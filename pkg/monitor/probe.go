package monitor

import (
	"context"

	"time"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm-go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	ProbeRoleName = "system.probe"

	ProbeAssignedPermissionName       = "system.probe.assigned-permission.name"
	ProbeAssignedPermissionResourceID = "system.probe.assigned-permission.resource-id"

	ProbeUnassignedPermissionName       = "system.probe.unassigned-permission.name"
	ProbeUnassignedPermissionResourceID = "system.probe.unassigned-permission.resource-id"
)

var ProbeActor = &protos.Actor{
	ID:     "probe",
	Issuer: "system",
}

type Probe struct {
	RoleServiceClient       protos.RoleServiceClient
	PermissionServiceClient protos.PermissionServiceClient
}

func (p *Probe) Setup(ctx context.Context, logger lager.Logger, uniqueSuffix string) error {
	logger.Debug(starting)
	doneChan := make(chan error)
	defer logger.Debug(finished)

	go func() {
		err := p.setupCreateRole(ctx, logger, uniqueSuffix)
		if err != nil {
			doneChan <- err
			return
		}
		doneChan <- p.setupAssignRole(ctx, logger, uniqueSuffix)
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-doneChan:
			return err
		}
	}
}

func (p *Probe) setupCreateRole(ctx context.Context, logger lager.Logger, uniqueSuffix string) error {
	roleName := ProbeRoleName + "." + uniqueSuffix

	assignedPermission := &protos.Permission{
		Name:            ProbeAssignedPermissionName,
		ResourcePattern: ProbeAssignedPermissionResourceID + "." + uniqueSuffix,
	}

	createRoleRequest := &protos.CreateRoleRequest{
		Name: roleName,
		Permissions: []*protos.Permission{
			assignedPermission,
		},
	}
	_, err := p.RoleServiceClient.CreateRole(ctx, createRoleRequest)
	s, ok := status.FromError(err)

	// Not a GRPC error
	if err != nil && !ok {
		logger.Error(failedToCreateRole, err, lager.Data{
			"roleName":    createRoleRequest.GetName(),
			"permissions": createRoleRequest.GetPermissions(),
		})
		return err
	}

	// GRPC error
	if err != nil && ok {
		switch s.Code() {
		case codes.AlreadyExists:

		default:
			logger.Error(failedToCreateRole, err, lager.Data{
				"roleName":    createRoleRequest.GetName(),
				"permissions": createRoleRequest.GetPermissions(),
			})
			return err
		}
	}

	return nil
}

func (p *Probe) setupAssignRole(ctx context.Context, logger lager.Logger, uniqueSuffix string) error {
	roleName := ProbeRoleName + "." + uniqueSuffix

	assignRoleRequest := &protos.AssignRoleRequest{
		Actor:    ProbeActor,
		RoleName: roleName,
	}

	_, err := p.RoleServiceClient.AssignRole(ctx, assignRoleRequest)
	s, ok := status.FromError(err)

	// Not a GRPC error
	if err != nil && !ok {
		logger.Error(failedToAssignRole, err, lager.Data{
			"roleName":     assignRoleRequest.GetRoleName(),
			"actor.id":     assignRoleRequest.GetActor().GetID(),
			"actor.issuer": assignRoleRequest.GetActor().GetIssuer(),
		})
		return err
	}

	// GRPC error
	if err != nil && ok {
		switch s.Code() {
		case codes.AlreadyExists:

		default:
			logger.Error(failedToAssignRole, err, lager.Data{
				"roleName":     assignRoleRequest.GetRoleName(),
				"actor.id":     assignRoleRequest.GetActor().GetID(),
				"actor.issuer": assignRoleRequest.GetActor().GetIssuer(),
			})
			return err
		}
	}

	return nil
}

func (p *Probe) Cleanup(ctx context.Context, cleanupTimeout time.Duration, logger lager.Logger, uniqueSuffix string) ([]time.Duration, error) {

	type cleanupResult struct {
		Error     error
		Durations []time.Duration
	}

	doneChan := make(chan cleanupResult)

	cleanupTimeoutTimer := time.After(cleanupTimeout)
	go func() {
		result := cleanupResult{}
		logger.Debug(starting)
		defer logger.Debug(finished)

		roleName := ProbeRoleName + "." + uniqueSuffix
		deleteRoleRequest := &protos.DeleteRoleRequest{
			Name: roleName,
		}

		start := time.Now()
		_, err := p.RoleServiceClient.DeleteRole(ctx, deleteRoleRequest)
		end := time.Now()
		result.Durations = append(result.Durations, end.Sub(start))
		s, ok := status.FromError(err)

		// Not a GRPC error
		if err != nil && !ok {
			logger.Error(failedToDeleteRole, err, lager.Data{
				"roleName": deleteRoleRequest.GetName(),
			})
			result.Error = err
			doneChan <- result
			return
		}

		// GRPC error
		if err != nil && ok {
			switch s.Code() {
			case codes.NotFound:

			default:
				logger.Error(failedToDeleteRole, err, lager.Data{
					"roleName": deleteRoleRequest.GetName(),
				})
				result.Error = err
				doneChan <- result
				return
			}
		}

		doneChan <- result
		return
	}()

	for {
		select {
		case <-cleanupTimeoutTimer:
			return []time.Duration{}, context.DeadlineExceeded
		case result := <-doneChan:
			select {
			case <-ctx.Done():
				return []time.Duration{}, ctx.Err()
			default:
				return result.Durations, result.Error
			}
		}
	}
}

func (p *Probe) Run(
	ctx context.Context,
	logger lager.Logger,
	uniqueSuffix string,
) (correct bool, durations []time.Duration, err error) {
	logger.Debug(starting)
	defer logger.Debug(finished)

	//	var duration time.Duration

	type result struct {
		Correct   bool
		Durations []time.Duration
		Err       error
	}

	doneChan := make(chan result)
	go func() {
		correct, duration, err := p.runAssignedPermission(ctx, logger, uniqueSuffix)
		r := result{}
		r.Durations = append(r.Durations, duration)
		if err != nil || !correct {
			r.Err = err
			r.Correct = correct
			doneChan <- r
			return
		}

		correct, duration, err = p.runUnassignedPermission(ctx, logger, uniqueSuffix)
		r.Durations = append(r.Durations, duration)
		r.Err = err
		r.Correct = correct
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
) (bool, time.Duration, error) {
	assignedPermission := &protos.Permission{
		Name:            ProbeAssignedPermissionName,
		ResourcePattern: ProbeAssignedPermissionResourceID + "." + uniqueSuffix,
	}

	logger = logger.Session("has-assigned-permission").WithData(lager.Data{
		"actor.id":                    ProbeActor.GetID(),
		"actor.issuer":                ProbeActor.GetIssuer(),
		"permission.name":             assignedPermission.GetName(),
		"permission.resource_pattern": assignedPermission.GetResourcePattern(),
	})
	req := &protos.HasPermissionRequest{
		Actor:          ProbeActor,
		PermissionName: assignedPermission.Name,
		ResourceId:     assignedPermission.ResourcePattern,
	}

	start := time.Now()
	res, err := p.PermissionServiceClient.HasPermission(ctx, req)
	end := time.Now()
	duration := end.Sub(start)

	if err != nil {
		logger.Error(failedToFindPermissions, err)
		return false, duration, err
	}

	if !res.GetHasPermission() {
		logger.Debug("incorrect-response", lager.Data{
			"expected": "true",
			"got":      "false",
		})
		return false, duration, nil
	}

	return true, duration, nil
}

func (p *Probe) runUnassignedPermission(
	ctx context.Context,
	logger lager.Logger,
	uniqueSuffix string,
) (bool, time.Duration, error) {
	unassignedPermission := &protos.Permission{
		Name:            ProbeUnassignedPermissionName,
		ResourcePattern: ProbeUnassignedPermissionResourceID + "." + uniqueSuffix,
	}

	logger = logger.Session("has-unassigned-permission").WithData(lager.Data{
		"actor.id":                    ProbeActor.GetID(),
		"actor.issuer":                ProbeActor.GetIssuer(),
		"permission.name":             unassignedPermission.GetName(),
		"permission.resource_pattern": unassignedPermission.GetResourcePattern(),
	})
	req := &protos.HasPermissionRequest{
		Actor:          ProbeActor,
		PermissionName: unassignedPermission.Name,
		ResourceId:     unassignedPermission.ResourcePattern,
	}

	start := time.Now()
	res, err := p.PermissionServiceClient.HasPermission(ctx, req)
	end := time.Now()
	duration := end.Sub(start)

	if err != nil {
		logger.Error(failedToFindPermissions, err)
		return false, duration, err
	}

	if res.GetHasPermission() {
		logger.Debug("incorrect-response", lager.Data{
			"expected": "false",
			"got":      "true",
		})
		return false, duration, nil
	}

	return true, duration, nil
}
