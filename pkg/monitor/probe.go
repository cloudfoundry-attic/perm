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
	defer logger.Debug(finished)

	err := p.setupCreateRole(ctx, logger, uniqueSuffix)
	if err != nil {
		return err
	}

	return p.setupAssignRole(ctx, logger, uniqueSuffix)
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

func (p *Probe) Cleanup(ctx context.Context, logger lager.Logger, uniqueSuffix string) error {
	logger.Debug(starting)
	defer logger.Debug(finished)

	roleName := ProbeRoleName + "." + uniqueSuffix

	deleteRoleRequest := &protos.DeleteRoleRequest{
		Name: roleName,
	}
	_, err := p.RoleServiceClient.DeleteRole(ctx, deleteRoleRequest)
	s, ok := status.FromError(err)

	// Not a GRPC error
	if err != nil && !ok {
		logger.Error(failedToDeleteRole, err, lager.Data{
			"roleName": deleteRoleRequest.GetName(),
		})
		return err
	}

	// GRPC error
	if err != nil && ok {
		switch s.Code() {
		case codes.NotFound:

		default:
			logger.Error(failedToDeleteRole, err, lager.Data{
				"roleName": deleteRoleRequest.GetName(),
			})
			return err
		}
	}

	return nil
}

func (p *Probe) Run(
	ctx context.Context,
	logger lager.Logger,
	uniqueSuffix string,
) (bool, []time.Duration, error) {
	logger.Debug(starting)
	defer logger.Debug(finished)

	var (
		correct  bool
		duration time.Duration
		err      error

		durations []time.Duration
	)

	correct, duration, err = p.runAssignedPermission(ctx, logger, uniqueSuffix)
	durations = append(durations, duration)

	if err != nil {
		return false, durations, err
	}
	if !correct {
		return false, durations, nil
	}

	correct, duration, err = p.runUnassignedPermission(ctx, logger, uniqueSuffix)
	durations = append(durations, duration)

	if err != nil {
		return false, durations, err
	}
	if !correct {
		return false, durations, nil
	}

	return true, durations, nil
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