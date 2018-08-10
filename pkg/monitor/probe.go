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
	probeAPICallsSuccess = "perm.probe.api.calls.success"

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
		ok            bool
		runErr        error
		slowEndpoints []string
	)

	suffix := uuid.NewV4().String()
	roleName := fmt.Sprintf("probe-role-%s", suffix)
	permission := perm.Permission{
		Action:          "probe.run",
		ResourcePattern: suffix,
	}

	logger := p.logger.WithData(logx.Data{Key: "uuid", Value: suffix})

	logger.Debug("starting")
	defer logger.Debug("finished")

	defer func() {
		switch err := runErr.(type) {
		case nil:
			p.sendGauge(probeRunsCorrect, 1)
			p.sendGauge(probeRunsSuccess, 1)
		case errIncorrectPermission:
			logger.Error("incorrect-permission", err, logx.Data{Key: "expected", Value: err.expected}, logx.Data{Key: "actual", Value: err.actual})
			p.sendGauge(probeRunsCorrect, 0)
			p.sendGauge(probeRunsSuccess, 0)
		case errExceededMaxLatency:
			logger.Error("exceeded-max-latency", err, logx.Data{Key: "endpoints", Value: slowEndpoints})
			p.sendGauge(probeRunsCorrect, 1)
			p.sendGauge(probeRunsSuccess, 0)
		default: // error from API call
			logger.Error("api-call-failed", err)
			p.sendGauge(probeRunsSuccess, 0)
		}
	}()

	defer func() {
		if runErr != nil {
			ctx, cancel := context.WithTimeout(context.Background(), p.cleanupTimeout)
			defer cancel()
			_, _ = p.client.DeleteRole(ctx, roleName)
		}

		if len(slowEndpoints) > 0 {
			runErr = errExceededMaxLatency{}
		}
	}()

	handler := func(ctx context.Context) (time.Duration, error) {
		_, duration, err := p.client.CreateRole(ctx, roleName, permission)
		return duration, err
	}
	ok, runErr = p.call(handler)
	if runErr != nil {
		return
	}
	if !ok {
		slowEndpoints = append(slowEndpoints, "CreateRole")
	}

	handler = func(ctx context.Context) (time.Duration, error) {
		return p.client.AssignRole(ctx, roleName, assignedActor)
	}
	ok, runErr = p.call(handler)
	if runErr != nil {
		return
	}
	if !ok {
		slowEndpoints = append(slowEndpoints, "AssignRole")
	}

	handler = func(ctx context.Context) (time.Duration, error) {
		hasPermission, duration, err := p.client.HasPermission(ctx, assignedActor, permission.Action, permission.ResourcePattern)
		if err != nil {
			return duration, err
		}
		if !hasPermission {
			return duration, errIncorrectPermission{
				expected: true,
				actual:   false,
			}
		}
		return duration, nil
	}
	ok, runErr = p.call(handler)
	if runErr != nil {
		return
	}
	if !ok {
		slowEndpoints = append(slowEndpoints, "HasPermission")
	}

	handler = func(ctx context.Context) (time.Duration, error) {
		hasPermission, duration, err := p.client.HasPermission(ctx, unassignedActor, permission.Action, permission.ResourcePattern)
		if err != nil {
			return duration, err
		}
		if hasPermission {
			return duration, errIncorrectPermission{
				expected: false,
				actual:   true,
			}
		}
		return duration, nil
	}
	ok, runErr = p.call(handler)
	if runErr != nil {
		return
	}
	if !ok {
		slowEndpoints = append(slowEndpoints, "HasPermission")
	}

	handler = func(ctx context.Context) (time.Duration, error) {
		return p.client.UnassignRole(ctx, roleName, assignedActor)
	}
	ok, runErr = p.call(handler)
	if runErr != nil {
		return
	}
	if !ok {
		slowEndpoints = append(slowEndpoints, "UnassignRole")
	}

	handler = func(ctx context.Context) (time.Duration, error) {
		return p.client.DeleteRole(ctx, roleName)
	}
	ok, runErr = p.call(handler)
	if runErr != nil {
		return
	}
	if !ok {
		slowEndpoints = append(slowEndpoints, "DeleteRole")
	}
}

func (p *Probe) call(handler func(context.Context) (time.Duration, error)) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), p.timeout)
	defer cancel()

	duration, err := handler(ctx)

	switch err.(type) {
	case nil:
	case recording.FailedToObserveDurationError:
	default:
		p.sendGauge(probeAPICallsSuccess, 0)
		return false, err
	}

	if duration > p.maxLatency {
		p.sendGauge(probeAPICallsSuccess, 0)
		return false, nil
	}

	p.sendGauge(probeAPICallsSuccess, 1)
	return true, nil
}

func (p *Probe) sendGauge(metric string, value int64) {
	if err := p.sender.Gauge(metric, value, alwaysSend); err != nil {
		p.logger.Error("failed-to-send-metric", err, logx.Data{Key: "metric", Value: metric}, logx.Data{Key: "value", Value: value})
	}
}
