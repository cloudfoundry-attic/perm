package probe

import (
	"context"
	"fmt"
	"time"

	"code.cloudfoundry.org/clock"
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
	CreateRole(ctx context.Context, name string, permissions ...perm.Permission) (perm.Role, error)
	DeleteRole(ctx context.Context, name string) error
	AssignRole(ctx context.Context, roleName string, actor perm.Actor) error
	UnassignRole(ctx context.Context, roleName string, actor perm.Actor) error
	HasPermission(ctx context.Context, actor perm.Actor, action, resource string) (bool, error)
}

type Probe struct {
	client         Client
	clock          clock.Clock
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
		clock:          o.clock,
		timeout:        o.timeout,
		cleanupTimeout: o.cleanupTimeout,
		maxLatency:     o.maxLatency,
	}
}

func (p *Probe) Run() error {
	var (
		err    error
		failed bool
	)

	suffix := uuid.NewV4().String()
	roleName := fmt.Sprintf("probe-role-%s", suffix)
	permission := perm.Permission{
		Action:          "probe.run",
		ResourcePattern: suffix,
	}

	defer func() {
		if err != nil {
			ctx, cancel := context.WithTimeout(context.Background(), p.cleanupTimeout)
			defer cancel()

			p.client.DeleteRole(ctx, roleName)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), p.timeout)
	defer cancel()

	start := p.clock.Now()
	if _, err = p.client.CreateRole(ctx, roleName, permission); err != nil {
		return err
	}
	if p.clock.Since(start) > p.maxLatency {
		failed = true
	}

	ctx, cancel = context.WithTimeout(context.Background(), p.timeout)
	defer cancel()

	start = p.clock.Now()
	if err = p.client.AssignRole(ctx, roleName, assignedActor); err != nil {
		return err
	}
	if p.clock.Since(start) > p.maxLatency {
		failed = true
	}

	ctx, cancel = context.WithTimeout(context.Background(), p.timeout)
	defer cancel()

	var hasPermission bool
	start = p.clock.Now()
	hasPermission, err = p.client.HasPermission(ctx, assignedActor, permission.Action, permission.ResourcePattern)
	if p.clock.Since(start) > p.maxLatency {
		failed = true
	}
	if !hasPermission {
		err = ErrIncorrectHasPermission
	}
	if err != nil {
		return err
	}

	ctx, cancel = context.WithTimeout(context.Background(), p.timeout)
	defer cancel()

	start = p.clock.Now()
	hasPermission, err = p.client.HasPermission(ctx, unassignedActor, permission.Action, permission.ResourcePattern)
	if p.clock.Since(start) > p.maxLatency {
		failed = true
	}
	if hasPermission {
		err = ErrIncorrectHasPermission
	}
	if err != nil {
		return err
	}

	start = p.clock.Now()
	ctx, cancel = context.WithTimeout(context.Background(), p.timeout)
	defer cancel()

	start = p.clock.Now()
	if err = p.client.UnassignRole(ctx, roleName, assignedActor); err != nil {
		return err
	}
	if p.clock.Since(start) > p.maxLatency {
		failed = true
	}

	ctx, cancel = context.WithTimeout(context.Background(), p.timeout)
	defer cancel()

	start = p.clock.Now()
	if err = p.client.DeleteRole(ctx, roleName); err != nil {
		return err
	}
	if p.clock.Since(start) > p.maxLatency {
		failed = true
	}

	if failed {
		return ErrExceededMaxLatency
	}

	return nil
}
