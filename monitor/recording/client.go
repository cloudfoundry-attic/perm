package recording

import (
	"context"
	"time"

	"code.cloudfoundry.org/clock"
	"code.cloudfoundry.org/perm"
)

//go:generate counterfeiter . Client

type Client interface {
	AssignRole(ctx context.Context, roleName string, actor perm.Actor) error
	CreateRole(ctx context.Context, name string, permissions ...perm.Permission) (perm.Role, error)
	DeleteRole(ctx context.Context, name string) error
	HasPermission(ctx context.Context, actor perm.Actor, action, resource string) (bool, error)
	UnassignRole(ctx context.Context, roleName string, actor perm.Actor) error
}

//go:generate counterfeiter . Recorder

type Recorder interface {
	Observe(duration time.Duration) error
}

type RecordingClient struct {
	client   Client
	recorder Recorder
	clock    clock.Clock
}

func NewClient(client Client, recorder Recorder, opts ...Option) *RecordingClient {
	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}

	return &RecordingClient{
		client:   client,
		recorder: recorder,
		clock:    o.clock,
	}
}

func (c *RecordingClient) AssignRole(ctx context.Context, roleName string, actor perm.Actor) (time.Duration, error) {
	start := c.clock.Now()

	if err := c.client.AssignRole(ctx, roleName, actor); err != nil {
		return 0, err
	}

	duration := c.clock.Since(start)

	if err := c.recorder.Observe(duration); err != nil {
		return duration, FailedToObserveDurationError{Err: err}
	}

	return duration, nil
}

func (c *RecordingClient) CreateRole(ctx context.Context, roleName string, permissions ...perm.Permission) (perm.Role, time.Duration, error) {
	start := c.clock.Now()

	role, err := c.client.CreateRole(ctx, roleName, permissions...)
	if err != nil {
		return perm.Role{}, 0, err
	}

	duration := c.clock.Since(start)

	if err := c.recorder.Observe(duration); err != nil {
		return role, duration, FailedToObserveDurationError{Err: err}
	}

	return role, duration, nil
}

func (c *RecordingClient) DeleteRole(ctx context.Context, roleName string) (time.Duration, error) {
	start := c.clock.Now()

	if err := c.client.DeleteRole(ctx, roleName); err != nil {
		return 0, err
	}

	duration := c.clock.Since(start)

	if err := c.recorder.Observe(duration); err != nil {
		return duration, FailedToObserveDurationError{Err: err}
	}

	return duration, nil
}

func (c *RecordingClient) HasPermission(ctx context.Context, actor perm.Actor, action, resource string) (bool, time.Duration, error) {
	start := c.clock.Now()

	hasPermission, err := c.client.HasPermission(ctx, actor, action, resource)
	if err != nil {
		return false, 0, err
	}

	duration := c.clock.Since(start)

	if err := c.recorder.Observe(duration); err != nil {
		return false, duration, FailedToObserveDurationError{Err: err}
	}

	return hasPermission, duration, nil
}

func (c *RecordingClient) UnassignRole(ctx context.Context, roleName string, actor perm.Actor) (time.Duration, error) {
	start := c.clock.Now()

	if err := c.client.UnassignRole(ctx, roleName, actor); err != nil {
		return 0, err
	}

	duration := c.clock.Since(start)

	if err := c.recorder.Observe(duration); err != nil {
		return duration, FailedToObserveDurationError{Err: err}
	}

	return duration, nil
}
