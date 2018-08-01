package recording

import (
	"context"
	"time"

	"code.cloudfoundry.org/clock"
	"code.cloudfoundry.org/perm/pkg/monitor"
	"code.cloudfoundry.org/perm/pkg/perm"
)

//go:generate counterfeiter . DurationRecorder

type DurationRecorder interface {
	Observe(duration time.Duration) error
}

type Client struct {
	client   monitor.Client
	recorder DurationRecorder
	clock    clock.Clock
}

func NewClient(client monitor.Client, recorder DurationRecorder, opts ...Option) *Client {
	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}

	return &Client{
		client:   client,
		recorder: recorder,
		clock:    o.clock,
	}
}

func (c *Client) AssignRole(ctx context.Context, roleName string, actor perm.Actor) error {
	start := c.clock.Now()
	if err := c.client.AssignRole(ctx, roleName, actor); err != nil {
		return err
	}

	if err := c.recorder.Observe(c.clock.Since(start)); err != nil {
		return FailedToObserveDurationError{Err: err}
	}

	return nil
}

func (c *Client) CreateRole(ctx context.Context, roleName string, permissions ...perm.Permission) (perm.Role, error) {
	start := c.clock.Now()
	role, err := c.client.CreateRole(ctx, roleName, permissions...)
	if err != nil {
		return perm.Role{}, err
	}

	if err := c.recorder.Observe(c.clock.Since(start)); err != nil {
		return role, FailedToObserveDurationError{Err: err}
	}

	return role, nil
}

func (c *Client) DeleteRole(ctx context.Context, roleName string) error {
	start := c.clock.Now()
	if err := c.client.DeleteRole(ctx, roleName); err != nil {
		return err
	}

	if err := c.recorder.Observe(c.clock.Since(start)); err != nil {
		return FailedToObserveDurationError{Err: err}
	}

	return nil
}

func (c *Client) UnassignRole(ctx context.Context, roleName string, actor perm.Actor) error {
	start := c.clock.Now()
	if err := c.client.UnassignRole(ctx, roleName, actor); err != nil {
		return err
	}

	if err := c.recorder.Observe(c.clock.Since(start)); err != nil {
		return FailedToObserveDurationError{Err: err}
	}

	return nil
}

func (c *Client) HasPermission(ctx context.Context, actor perm.Actor, action, resource string) (bool, error) {
	start := c.clock.Now()
	hasPermission, err := c.client.HasPermission(ctx, actor, action, resource)
	if err != nil {
		return false, err
	}

	if err := c.recorder.Observe(c.clock.Since(start)); err != nil {
		return false, FailedToObserveDurationError{Err: err}
	}

	return hasPermission, nil
}
