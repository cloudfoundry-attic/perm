package monitor

import (
	"context"
	"fmt"
	"time"

	"code.cloudfoundry.org/perm/pkg/perm"
	uuid "github.com/satori/go.uuid"
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

type Probe struct {
	client         Client
	timeout        time.Duration
	cleanupTimeout time.Duration
	maxLatency     time.Duration
}

func NewProbe(client Client, opts ...Option) *Probe {
	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}

	return &Probe{
		client:         client,
		timeout:        o.timeout,
		cleanupTimeout: o.cleanupTimeout,
		maxLatency:     o.maxLatency,
	}
}

func (p *Probe) Run() error {
	var (
		hasPermission      bool
		duration           time.Duration
		cleanupErr         error
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
		if cleanupErr == nil {
			return
		}

		switch cleanupErr.(type) {
		case HasAssignedPermissionError:
			// do nothing
		case HasUnassignedPermissionError:
			// do nothing
		case ExceededMaxLatencyError:
			// do nothing
		default:
			ctx, cancel := context.WithTimeout(context.Background(), p.cleanupTimeout)
			defer cancel()

			_, _ = p.client.DeleteRole(ctx, roleName)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), p.timeout)
	defer cancel()
	if _, duration, cleanupErr = p.client.CreateRole(ctx, roleName, permission); cleanupErr != nil {
		return cleanupErr
	}
	if duration > p.maxLatency {
		exceededMaxLatency = true
	}

	ctx, cancel = context.WithTimeout(context.Background(), p.timeout)
	defer cancel()
	if duration, cleanupErr = p.client.AssignRole(ctx, roleName, assignedActor); cleanupErr != nil {
		return cleanupErr
	}
	if duration > p.maxLatency {
		exceededMaxLatency = true
	}

	ctx, cancel = context.WithTimeout(context.Background(), p.timeout)
	defer cancel()
	hasPermission, duration, cleanupErr = p.client.HasPermission(ctx, assignedActor, permission.Action, permission.ResourcePattern)
	if cleanupErr != nil {
		return cleanupErr
	}
	if duration > p.maxLatency {
		exceededMaxLatency = true
	}
	if !hasPermission {
		cleanupErr = HasAssignedPermissionError{}
		return cleanupErr
	}

	ctx, cancel = context.WithTimeout(context.Background(), p.timeout)
	defer cancel()
	hasPermission, duration, cleanupErr = p.client.HasPermission(ctx, unassignedActor, permission.Action, permission.ResourcePattern)
	if cleanupErr != nil {
		return cleanupErr
	}
	if duration > p.maxLatency {
		exceededMaxLatency = true
	}
	if hasPermission {
		cleanupErr = HasUnassignedPermissionError{}
		return cleanupErr
	}

	ctx, cancel = context.WithTimeout(context.Background(), p.timeout)
	defer cancel()
	if duration, cleanupErr = p.client.UnassignRole(ctx, roleName, assignedActor); cleanupErr != nil {
		return cleanupErr
	}
	if duration > p.maxLatency {
		exceededMaxLatency = true
	}

	ctx, cancel = context.WithTimeout(context.Background(), p.timeout)
	defer cancel()
	if duration, cleanupErr = p.client.DeleteRole(ctx, roleName); cleanupErr != nil {
		return cleanupErr
	}
	if duration > p.maxLatency {
		exceededMaxLatency = true
	}

	if exceededMaxLatency {
		cleanupErr = ExceededMaxLatencyError{}
		return cleanupErr
	}

	return nil
}
