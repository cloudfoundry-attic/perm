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
	probeRunsCorrect     = "perm.probe.runs.correct"
	probeRunsSuccess     = "perm.probe.runs.success"
	probeAPICallsSuccess = "perm.probe.api.runs.success"

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

//go:generate counterfeiter . Sender

type Sender interface {
	Gauge(string, int64, float32) error
}

type Probe struct {
	client         Client
	sender         Sender
	logger         logx.Logger
	timeout        time.Duration
	cleanupTimeout time.Duration
	maxLatency     time.Duration
}

func NewProbe(client Client, sender Sender, logger logx.Logger, opts ...Option) *Probe {
	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}

	return &Probe{
		client:         client,
		sender:         sender,
		logger:         logger.WithName("probe"),
		timeout:        o.timeout,
		cleanupTimeout: o.cleanupTimeout,
		maxLatency:     o.maxLatency,
	}
}

func (p *Probe) Run() {
	var (
		hasPermission      bool
		duration           time.Duration
		runErr             error
		exceededMaxLatency bool
	)

	p.logger.Debug("starting")
	defer p.logger.Debug("finished")

	suffix := uuid.NewV4().String()
	roleName := fmt.Sprintf("probe-role-%s", suffix)
	permission := perm.Permission{
		Action:          "probe.run",
		ResourcePattern: suffix,
	}

	defer func() {
		switch runErr.(type) {
		case nil:
			p.sendGauge(probeRunsCorrect, 1)
			p.sendGauge(probeRunsSuccess, 1)
		case HasAssignedPermissionError:
			p.logger.Error("incorrect-permission", runErr)
			p.sendGauge(probeRunsCorrect, 0)
			p.sendGauge(probeRunsSuccess, 0)
		case HasUnassignedPermissionError:
			p.logger.Error("incorrect-permission", runErr)
			p.sendGauge(probeRunsCorrect, 0)
			p.sendGauge(probeRunsSuccess, 0)
		case ExceededMaxLatencyError:
			p.logger.Error("exceeded-max-latency", runErr)
			p.sendGauge(probeRunsCorrect, 1)
			p.sendGauge(probeRunsSuccess, 0)
		default: // error from API call
			p.logger.Error("api-call-failed", runErr)
			p.sendGauge(probeRunsSuccess, 0)
		}
	}()

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
			p.sendGauge(probeAPICallsSuccess, 0)
			return
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
			p.sendGauge(probeAPICallsSuccess, 0)
			return
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
			p.sendGauge(probeAPICallsSuccess, 0)
			return
		}
	}

	if p.exceededMaxLatency(duration) {
		exceededMaxLatency = true
	}

	if !hasPermission {
		runErr = HasAssignedPermissionError{}
		return
	}

	// check has no permission
	ctx, cancel = context.WithTimeout(context.Background(), p.timeout)
	defer cancel()

	if hasPermission, duration, runErr = p.client.HasPermission(ctx, unassignedActor, permission.Action, permission.ResourcePattern); runErr != nil {
		switch runErr.(type) {
		case recording.FailedToObserveDurationError:
			// do nothing
		default:
			p.sendGauge(probeAPICallsSuccess, 0)
			return
		}
	}

	if p.exceededMaxLatency(duration) {
		exceededMaxLatency = true
	}

	if hasPermission {
		runErr = HasUnassignedPermissionError{}
		return
	}

	// unassign role
	ctx, cancel = context.WithTimeout(context.Background(), p.timeout)
	defer cancel()

	if duration, runErr = p.client.UnassignRole(ctx, roleName, assignedActor); runErr != nil {
		switch runErr.(type) {
		case recording.FailedToObserveDurationError:
			// do nothing
		default:
			p.sendGauge(probeAPICallsSuccess, 0)
			return
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
			p.sendGauge(probeAPICallsSuccess, 0)
			return
		}
	}

	if p.exceededMaxLatency(duration) {
		exceededMaxLatency = true
	}

	if exceededMaxLatency {
		runErr = ExceededMaxLatencyError{}
		return
	}

	runErr = nil
}

func (p *Probe) exceededMaxLatency(duration time.Duration) bool {
	if duration > p.maxLatency {
		p.sendGauge(probeAPICallsSuccess, 0)
		return true
	}

	p.sendGauge(probeAPICallsSuccess, 1)
	return false
}

func (p *Probe) sendGauge(metric string, value int64) {
	if err := p.sender.Gauge(metric, value, alwaysSend); err != nil {
		p.logger.Error(fmt.Sprintf("failed-to-send-%s=%d", metric, value), err)
	}
}
