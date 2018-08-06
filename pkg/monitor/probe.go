package monitor

import (
	"context"
	"fmt"
	"time"

	"code.cloudfoundry.org/perm/pkg/logx"
	"code.cloudfoundry.org/perm/pkg/monitor/recording"
	"code.cloudfoundry.org/perm/pkg/perm"
	uuid "github.com/satori/go.uuid"
)

const (
	// metric names
	probeRunsCorrect    = "perm.probe.runs.correct"
	probeRunsSuccess    = "perm.probe.runs.success"
	probeAPIRunsSuccess = "perm.probe.api.runs.success"

	alwaysSend = 1
)

var (
	assignedActor = perm.Actor{
		ID:        "actor-with-role",
		Namespace: "probe.system",
	}
	unassignedActor = perm.Actor{
		ID:        "actor-without-role",
		Namespace: "probe.system",
	}
)

//go:generate counterfeiter . Client

type Client interface {
	AssignRole(ctx context.Context, roleName string, actor perm.Actor) (time.Duration, error)
	CreateRole(ctx context.Context, name string, permissions ...perm.Permission) (perm.Role, time.Duration, error)
	DeleteRole(ctx context.Context, name string) (time.Duration, error)
	HasPermission(ctx context.Context, actor perm.Actor, action, resource string) (bool, time.Duration, error)
	UnassignRole(ctx context.Context, roleName string, actor perm.Actor) (time.Duration, error)
}

//go:generate counterfeiter . Store

type Store interface {
	Collect() map[string]int64
}

//go:generate counterfeiter . Sender

type Sender interface {
	Gauge(string, int64, float32) error
}

type Probe struct {
	client         Client
	store          Store
	sender         Sender
	logger         logx.Logger
	timeout        time.Duration
	cleanupTimeout time.Duration
	maxLatency     time.Duration
}

func NewProbe(client Client, store Store, sender Sender, logger logx.Logger, opts ...Option) *Probe {
	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}

	return &Probe{
		client:         client,
		store:          store,
		sender:         sender,
		logger:         logger.WithName("probe"),
		timeout:        o.timeout,
		cleanupTimeout: o.cleanupTimeout,
		maxLatency:     o.maxLatency,
	}
}

func (p *Probe) Run() {
	p.logger.Debug("starting")
	defer p.logger.Debug("finished")

	err := p.Probe()
	if err == nil {
		p.sendGauge(probeRunsCorrect, 1)
		p.sendGauge(probeRunsSuccess, 1)
		for metric, value := range p.store.Collect() {
			p.sendGauge(metric, value)
		}
		return
	}

	switch err.(type) {
	case HasAssignedPermissionError:
		p.logger.Error("incorrect-permission", err)
		p.sendGauge(probeRunsCorrect, 0)
		p.sendGauge(probeRunsSuccess, 0)
	case HasUnassignedPermissionError:
		p.logger.Error("incorrect-permission", err)
		p.sendGauge(probeRunsCorrect, 0)
		p.sendGauge(probeRunsSuccess, 0)
	case ExceededMaxLatencyError:
		p.logger.Error("exceeded-max-latency", err)
		p.sendGauge(probeRunsCorrect, 1)
		p.sendGauge(probeRunsSuccess, 0)
	default: // error from API call
		p.logger.Error("api-call-failed", err)
		p.sendGauge(probeRunsSuccess, 0)
	}
}

func (p *Probe) Probe() error {
	var (
		hasPermission      bool
		duration           time.Duration
		runErr             error
		exceededMaxLatency bool
	)

	suffix := uuid.NewV4().String()
	roleName := fmt.Sprintf("probe-role-%s", suffix)
	permission := perm.Permission{
		Action:          "probe.run",
		ResourcePattern: suffix,
	}

	// cleanup
	defer func() {
		if runErr != nil {
			ctx, cancel := context.WithTimeout(context.Background(), p.cleanupTimeout)
			defer cancel()
			_, _ = p.client.DeleteRole(ctx, roleName)
		}
	}()

	// create role
	ctx, cancel := context.WithTimeout(context.Background(), p.timeout)
	defer cancel()

	if _, duration, runErr = p.client.CreateRole(ctx, roleName, permission); runErr != nil {
		switch runErr.(type) {
		case recording.FailedToObserveDurationError:
			// do nothing
		default:
			p.sendGauge(probeAPIRunsSuccess, 0)
			return runErr
		}
	}

	if p.exceededMaxLatency(duration) {
		exceededMaxLatency = true
	}

	// assign role
	ctx, cancel = context.WithTimeout(context.Background(), p.timeout)
	defer cancel()

	if duration, runErr = p.client.AssignRole(ctx, roleName, assignedActor); runErr != nil {
		switch runErr.(type) {
		case recording.FailedToObserveDurationError:
			// do nothing
		default:
			p.sendGauge(probeAPIRunsSuccess, 0)
			return runErr
		}
	}

	if p.exceededMaxLatency(duration) {
		exceededMaxLatency = true
	}

	// check has permission
	ctx, cancel = context.WithTimeout(context.Background(), p.timeout)
	defer cancel()

	if hasPermission, duration, runErr = p.client.HasPermission(ctx, assignedActor, permission.Action, permission.ResourcePattern); runErr != nil {
		switch runErr.(type) {
		case recording.FailedToObserveDurationError:
			// do nothing
		default:
			p.sendGauge(probeAPIRunsSuccess, 0)
			return runErr
		}
	}

	if p.exceededMaxLatency(duration) {
		exceededMaxLatency = true
	}

	if !hasPermission {
		runErr = HasAssignedPermissionError{}
		return runErr
	}

	// check has no permission
	ctx, cancel = context.WithTimeout(context.Background(), p.timeout)
	defer cancel()

	if hasPermission, duration, runErr = p.client.HasPermission(ctx, unassignedActor, permission.Action, permission.ResourcePattern); runErr != nil {
		switch runErr.(type) {
		case recording.FailedToObserveDurationError:
			// do nothing
		default:
			p.sendGauge(probeAPIRunsSuccess, 0)
			return runErr
		}
	}

	if p.exceededMaxLatency(duration) {
		exceededMaxLatency = true
	}

	if hasPermission {
		runErr = HasUnassignedPermissionError{}
		return runErr
	}

	// unassign role
	ctx, cancel = context.WithTimeout(context.Background(), p.timeout)
	defer cancel()

	if duration, runErr = p.client.UnassignRole(ctx, roleName, assignedActor); runErr != nil {
		switch runErr.(type) {
		case recording.FailedToObserveDurationError:
			// do nothing
		default:
			p.sendGauge(probeAPIRunsSuccess, 0)
			return runErr
		}
	}

	if p.exceededMaxLatency(duration) {
		exceededMaxLatency = true
	}

	// delete role
	ctx, cancel = context.WithTimeout(context.Background(), p.timeout)
	defer cancel()

	if duration, runErr = p.client.DeleteRole(ctx, roleName); runErr != nil {
		switch runErr.(type) {
		case recording.FailedToObserveDurationError:
			// do nothing
		default:
			p.sendGauge(probeAPIRunsSuccess, 0)
			return runErr
		}
	}

	if p.exceededMaxLatency(duration) {
		exceededMaxLatency = true
	}

	if exceededMaxLatency {
		runErr = ExceededMaxLatencyError{}
		return runErr
	}

	runErr = nil
	return runErr
}

func (p *Probe) exceededMaxLatency(duration time.Duration) bool {
	if duration > p.maxLatency {
		p.sendGauge(probeAPIRunsSuccess, 0)
		return true
	}

	p.sendGauge(probeAPIRunsSuccess, 1)
	return false
}

func (p *Probe) sendGauge(metric string, value int64) {
	if err := p.sender.Gauge(metric, value, alwaysSend); err != nil {
		p.logger.Error(fmt.Sprintf("failed-to-send-%s=%d", metric, value), err)
	}
}
