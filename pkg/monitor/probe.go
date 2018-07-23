package monitor

import (
	"context"

	"time"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/pkg/perm"
)

const (
	ProbeRoleName = "system.probe"

	ProbeAssignedPermissionAction   = "system.probe.assigned-permission.action"
	ProbeAssignedPermissionResource = "system.probe.assigned-permission.resource"

	ProbeUnassignedPermissionAction   = "system.probe.unassigned-permission.action"
	ProbeUnassignedPermissionResource = "system.probe.unassigned-permission.resource"
)

//go:generate counterfeiter . Client

type Client interface {
	CreateRole(ctx context.Context, name string, permissions ...perm.Permission) (perm.Role, error)
	DeleteRole(ctx context.Context, name string) error
	AssignRole(ctx context.Context, roleName string, actor perm.Actor) error
	HasPermission(ctx context.Context, actor perm.Actor, action, resource string) (bool, error)
}

var ProbeActor = perm.Actor{
	ID:        "probe",
	Namespace: "system",
}

type Probe struct {
	client Client
}

type LabeledDuration struct {
	Label    string
	Duration time.Duration
}

func NewProbe(client Client) *Probe {
	return &Probe{
		client: client,
	}
}

func (p *Probe) Setup(ctx context.Context, logger lager.Logger, uniqueSuffix string) ([]LabeledDuration, error) {
	type setupResult struct {
		Error     error
		Durations []LabeledDuration
	}

	logger.Debug(starting)
	doneChan := make(chan setupResult)
	defer logger.Debug(finished)

	go func() {
		duration, err := p.setupCreateRole(ctx, logger, uniqueSuffix)
		result := setupResult{err, []LabeledDuration{duration}}
		if err != nil {
			doneChan <- result
			return
		}
		duration, err = p.setupAssignRole(ctx, logger, uniqueSuffix)
		result.Error = err
		result.Durations = append(result.Durations, duration)
		doneChan <- result
		return
	}()

	for {
		select {
		case <-ctx.Done():
			return []LabeledDuration{}, ctx.Err()
		case result := <-doneChan:
			return result.Durations, result.Error
		}
	}
}

func (p *Probe) setupCreateRole(ctx context.Context, logger lager.Logger, uniqueSuffix string) (LabeledDuration, error) {
	roleName := ProbeRoleName + "." + uniqueSuffix

	start := time.Now()

	permissions := []perm.Permission{
		perm.Permission{
			Action:          ProbeAssignedPermissionAction,
			ResourcePattern: ProbeAssignedPermissionResource,
		},
	}
	_, err := p.client.CreateRole(ctx, roleName, permissions...)

	end := time.Now()

	duration := end.Sub(start)

	if err != nil && err != perm.ErrRoleAlreadyExists {
		logger.Error(failedToCreateRole, err, lager.Data{
			"roleName":    roleName,
			"permissions": permissions,
		})

		return LabeledDuration{}, err
	}

	return LabeledDuration{Label: "CreateRole", Duration: duration}, nil
}

func (p *Probe) setupAssignRole(ctx context.Context, logger lager.Logger, uniqueSuffix string) (LabeledDuration, error) {
	roleName := ProbeRoleName + "." + uniqueSuffix
	start := time.Now()

	err := p.client.AssignRole(ctx, roleName, ProbeActor)

	end := time.Now()
	duration := end.Sub(start)

	if err != nil && err != perm.ErrAssignmentAlreadyExists {
		logger.Error(failedToAssignRole, err, lager.Data{
			"roleName":        roleName,
			"actor.id":        ProbeActor.ID,
			"actor.namespace": ProbeActor.Namespace,
		})

		return LabeledDuration{}, err
	}

	return LabeledDuration{Label: "AssignRole", Duration: duration}, nil
}

func (p *Probe) Cleanup(ctx context.Context, cleanupTimeout time.Duration, logger lager.Logger, uniqueSuffix string) ([]LabeledDuration, error) {
	type cleanupResult struct {
		Error     error
		Durations []LabeledDuration
	}

	doneChan := make(chan cleanupResult)

	cleanupTimeoutTimer := time.After(cleanupTimeout)
	go func() {
		result := cleanupResult{}
		logger.Debug(starting)
		defer logger.Debug(finished)

		roleName := ProbeRoleName + "." + uniqueSuffix

		start := time.Now()
		err := p.client.DeleteRole(ctx, roleName)
		end := time.Now()
		result.Durations = append(result.Durations, LabeledDuration{Label: "DeleteRole", Duration: end.Sub(start)})

		if err != nil && err != perm.ErrRoleNotFound {
			logger.Error(failedToDeleteRole, err, lager.Data{
				"roleName": roleName,
			})
			result.Error = err
			doneChan <- result
			return
		}

		doneChan <- result
		return
	}()

	for {
		select {
		case result := <-doneChan:
			select {
			case <-ctx.Done():
				return []LabeledDuration{}, ctx.Err()
			default:
				return result.Durations, result.Error
			}
		case <-cleanupTimeoutTimer:
			return []LabeledDuration{}, context.DeadlineExceeded
		}
	}
}

func (p *Probe) Run(
	ctx context.Context,
	logger lager.Logger,
	uniqueSuffix string,
) (correct bool, durations []LabeledDuration, err error) {
	logger.Debug(starting)
	defer logger.Debug(finished)

	type result struct {
		Correct   bool
		Durations []LabeledDuration
		Err       error
	}

	doneChan := make(chan result)
	go func() {
		permission, duration, runErr := p.runAssignedPermission(ctx, logger, uniqueSuffix)
		r := result{}
		r.Durations = append(r.Durations, duration)
		if runErr != nil || !permission {
			r.Err = runErr
			r.Correct = permission
			doneChan <- r
			return
		}

		permission, duration, runErr = p.runUnassignedPermission(ctx, logger, uniqueSuffix)
		r.Durations = append(r.Durations, duration)
		r.Err = runErr
		r.Correct = permission
		doneChan <- r
	}()

	for {
		select {
		case <-ctx.Done():
			err = ctx.Err()
			return
		case result := <-doneChan:
			correct = result.Correct
			durations = result.Durations
			err = result.Err
			return
		}
	}
}

func (p *Probe) runAssignedPermission(
	ctx context.Context,
	logger lager.Logger,
	uniqueSuffix string,
) (bool, LabeledDuration, error) {
	logger = logger.Session("has-assigned-permission").WithData(lager.Data{
		"actor.id":                   ProbeActor.ID,
		"actor.namespace":            ProbeActor.Namespace,
		"permission.action":          ProbeAssignedPermissionAction,
		"permission.resourcePattern": ProbeAssignedPermissionResource,
	})

	start := time.Now()
	hasPermission, err := p.client.HasPermission(ctx, ProbeActor, ProbeAssignedPermissionAction, ProbeAssignedPermissionResource)
	end := time.Now()
	duration := end.Sub(start)

	if err != nil {
		logger.Error(failedToFindPermissions, err)
		return false, LabeledDuration{}, err
	}

	if !hasPermission {
		logger.Info("incorrect-response", lager.Data{
			"expected": "true",
			"got":      "false",
		})
		return false, LabeledDuration{}, nil
	}

	return true, LabeledDuration{Label: "HasPermission", Duration: duration}, nil
}

func (p *Probe) runUnassignedPermission(
	ctx context.Context,
	logger lager.Logger,
	uniqueSuffix string,
) (bool, LabeledDuration, error) {
	logger = logger.Session("has-unassigned-permission").WithData(lager.Data{
		"actor.id":                   ProbeActor.ID,
		"actor.namespace":            ProbeActor.Namespace,
		"permission.action":          ProbeUnassignedPermissionAction,
		"permission.resourcePattern": ProbeUnassignedPermissionResource,
	})

	start := time.Now()
	hasPermission, err := p.client.HasPermission(ctx, ProbeActor, ProbeUnassignedPermissionAction, ProbeUnassignedPermissionResource)
	end := time.Now()
	duration := end.Sub(start)

	if err != nil {
		logger.Error(failedToFindPermissions, err)
		return false, LabeledDuration{}, err
	}

	if hasPermission {
		logger.Info("incorrect-response", lager.Data{
			"expected": "false",
			"got":      "true",
		})
		return false, LabeledDuration{}, nil
	}

	return true, LabeledDuration{Label: "HasPermission", Duration: duration}, nil
}
