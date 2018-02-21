package monitor

import (
	"context"

	"time"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm-go"
	"code.cloudfoundry.org/perm/messages"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	QueryProbeRoleName = "system.query-probe"

	QueryProbeAssignedPermissionName       = "system.query-probe.assigned-permission.name"
	QueryProbeAssignedPermissionResourceID = "system.query-probe.assigned-permission.resource-id"

	QueryProbeUnassignedPermissionName       = "system.query-probe.unassigned-permission.name"
	QueryProbeUnassignedPermissionResourceID = "system.query-probe.unassigned-permission.resource-id"
)

var QueryProbeActor = &protos.Actor{
	ID:     "query-probe",
	Issuer: "system",
}

type QueryProbe struct {
	RoleServiceClient       protos.RoleServiceClient
	PermissionServiceClient protos.PermissionServiceClient
}

func (p *QueryProbe) Setup(ctx context.Context, logger lager.Logger, uniqueSuffix string) error {
	logger.Debug(messages.Starting)
	defer logger.Debug(messages.Finished)

	err := p.setupCreateRole(ctx, logger, uniqueSuffix)
	if err != nil {
		return err
	}

	return p.setupAssignRole(ctx, logger, uniqueSuffix)
}

func (p *QueryProbe) setupCreateRole(ctx context.Context, logger lager.Logger, uniqueSuffix string) error {
	roleName := QueryProbeRoleName + "." + uniqueSuffix

	assignedPermission := &protos.Permission{
		Name:            QueryProbeAssignedPermissionName,
		ResourcePattern: QueryProbeAssignedPermissionResourceID + "." + uniqueSuffix,
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
		logger.Error(messages.FailedToCreateRole, err, lager.Data{
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
			logger.Error(messages.FailedToCreateRole, err, lager.Data{
				"roleName":    createRoleRequest.GetName(),
				"permissions": createRoleRequest.GetPermissions(),
			})
			return err
		}
	}

	return nil
}

func (p *QueryProbe) setupAssignRole(ctx context.Context, logger lager.Logger, uniqueSuffix string) error {
	roleName := QueryProbeRoleName + "." + uniqueSuffix

	assignRoleRequest := &protos.AssignRoleRequest{
		Actor:    QueryProbeActor,
		RoleName: roleName,
	}

	_, err := p.RoleServiceClient.AssignRole(ctx, assignRoleRequest)
	s, ok := status.FromError(err)

	// Not a GRPC error
	if err != nil && !ok {
		logger.Error(messages.FailedToAssignRole, err, lager.Data{
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
			logger.Error(messages.FailedToAssignRole, err, lager.Data{
				"roleName":     assignRoleRequest.GetRoleName(),
				"actor.id":     assignRoleRequest.GetActor().GetID(),
				"actor.issuer": assignRoleRequest.GetActor().GetIssuer(),
			})
			return err
		}
	}

	return nil
}

func (p *QueryProbe) Cleanup(ctx context.Context, logger lager.Logger, uniqueSuffix string) error {
	logger.Debug(messages.Starting)
	defer logger.Debug(messages.Finished)

	roleName := QueryProbeRoleName + "." + uniqueSuffix

	deleteRoleRequest := &protos.DeleteRoleRequest{
		Name: roleName,
	}
	_, err := p.RoleServiceClient.DeleteRole(ctx, deleteRoleRequest)
	s, ok := status.FromError(err)

	// Not a GRPC error
	if err != nil && !ok {
		logger.Error(messages.FailedToDeleteRole, err, lager.Data{
			"roleName": deleteRoleRequest.GetName(),
		})
		return err
	}

	// GRPC error
	if err != nil && ok {
		switch s.Code() {
		case codes.NotFound:

		default:
			logger.Error(messages.FailedToDeleteRole, err, lager.Data{
				"roleName": deleteRoleRequest.GetName(),
			})
			return err
		}
	}

	return nil
}

func (p *QueryProbe) Run(
	ctx context.Context,
	logger lager.Logger,
	uniqueSuffix string,
) (bool, []time.Duration, error) {
	logger.Debug(messages.Starting)
	defer logger.Debug(messages.Finished)

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

func (p *QueryProbe) runAssignedPermission(
	ctx context.Context,
	logger lager.Logger,
	uniqueSuffix string,
) (bool, time.Duration, error) {
	assignedPermission := &protos.Permission{
		Name:            QueryProbeAssignedPermissionName,
		ResourcePattern: QueryProbeAssignedPermissionResourceID + "." + uniqueSuffix,
	}

	logger = logger.Session("has-assigned-permission").WithData(lager.Data{
		"actor.id":                    QueryProbeActor.GetID(),
		"actor.issuer":                QueryProbeActor.GetIssuer(),
		"permission.name":             assignedPermission.GetName(),
		"permission.resource_pattern": assignedPermission.GetResourcePattern(),
	})
	req := &protos.HasPermissionRequest{
		Actor:          QueryProbeActor,
		PermissionName: assignedPermission.Name,
		ResourceId:     assignedPermission.ResourcePattern,
	}

	start := time.Now()
	res, err := p.PermissionServiceClient.HasPermission(ctx, req)
	end := time.Now()
	duration := end.Sub(start)

	if err != nil {
		logger.Error(messages.FailedToFindPermissions, err)
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

func (p *QueryProbe) runUnassignedPermission(
	ctx context.Context,
	logger lager.Logger,
	uniqueSuffix string,
) (bool, time.Duration, error) {
	unassignedPermission := &protos.Permission{
		Name:            QueryProbeUnassignedPermissionName,
		ResourcePattern: QueryProbeUnassignedPermissionResourceID + "." + uniqueSuffix,
	}

	logger = logger.Session("has-unassigned-permission").WithData(lager.Data{
		"actor.id":                    QueryProbeActor.GetID(),
		"actor.issuer":                QueryProbeActor.GetIssuer(),
		"permission.name":             unassignedPermission.GetName(),
		"permission.resource_pattern": unassignedPermission.GetResourcePattern(),
	})
	req := &protos.HasPermissionRequest{
		Actor:          QueryProbeActor,
		PermissionName: unassignedPermission.Name,
		ResourceId:     unassignedPermission.ResourcePattern,
	}

	start := time.Now()
	res, err := p.PermissionServiceClient.HasPermission(ctx, req)
	end := time.Now()
	duration := end.Sub(start)

	if err != nil {
		logger.Error(messages.FailedToFindPermissions, err)
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
