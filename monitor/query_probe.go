package monitor

import (
	"context"

	"time"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/messages"
	"code.cloudfoundry.org/perm/protos"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	QueryProbeRoleName = "system.query-probe"
)

var QueryProbeActor = &protos.Actor{
	ID:     "query-probe",
	Issuer: "system",
}

var QueryProbeAssignedPermission = &protos.Permission{
	Name:            "system.query-probe.assigned-permission.name",
	ResourcePattern: "system.query-probe.assigned-permission.resource-id",
}

var QueryProbeUnassignedPermission = &protos.Permission{
	Name:            "system.query-probe.unassigned-permission.name",
	ResourcePattern: "system.query-probe.unassigned-permission.resource-id",
}

//go:generate counterfeiter code.cloudfoundry.org/perm/protos.PermissionServiceClient

type QueryProbe struct {
	RoleServiceClient       protos.RoleServiceClient
	PermissionServiceClient protos.PermissionServiceClient
}

func (p *QueryProbe) Setup(ctx context.Context, logger lager.Logger) error {
	logger.Debug(messages.Starting)
	defer logger.Debug(messages.Finished)

	err := p.setupCreateRole(ctx, logger)
	if err != nil {
		return err
	}

	return p.setupAssignRole(ctx, logger)
}

func (p *QueryProbe) setupCreateRole(ctx context.Context, logger lager.Logger) error {
	createRoleRequest := &protos.CreateRoleRequest{
		Name: QueryProbeRoleName,
		Permissions: []*protos.Permission{
			QueryProbeAssignedPermission,
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

func (p *QueryProbe) setupAssignRole(ctx context.Context, logger lager.Logger) error {
	assignRoleRequest := &protos.AssignRoleRequest{
		Actor:    QueryProbeActor,
		RoleName: QueryProbeRoleName,
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

func (p *QueryProbe) Cleanup(ctx context.Context, logger lager.Logger) error {
	logger.Debug(messages.Starting)
	defer logger.Debug(messages.Finished)

	deleteRoleRequest := &protos.DeleteRoleRequest{
		Name: QueryProbeRoleName,
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

func (p *QueryProbe) Run(ctx context.Context, logger lager.Logger) (bool, []time.Duration, error) {
	logger.Debug(messages.Starting)
	defer logger.Debug(messages.Finished)

	var (
		correct  bool
		duration time.Duration
		err      error

		durations []time.Duration
	)

	correct, duration, err = p.runAssignedPermission(ctx, logger)
	durations = append(durations, duration)

	if err != nil {
		return false, durations, err
	}
	if !correct {
		return false, durations, nil
	}

	correct, duration, err = p.runUnassignedPermission(ctx, logger)
	durations = append(durations, duration)

	if err != nil {
		return false, durations, err
	}
	if !correct {
		return false, durations, nil
	}

	return true, durations, nil
}

func (p *QueryProbe) runAssignedPermission(ctx context.Context, logger lager.Logger) (bool, time.Duration, error) {
	logger = logger.Session("has-assigned-permission").WithData(lager.Data{
		"actor.id":                    QueryProbeActor.GetID(),
		"actor.issuer":                QueryProbeActor.GetIssuer(),
		"permission.name":             QueryProbeAssignedPermission.GetName(),
		"permission.resource_pattern": QueryProbeAssignedPermission.GetResourcePattern(),
	})
	req := &protos.HasPermissionRequest{
		Actor:          QueryProbeActor,
		PermissionName: QueryProbeAssignedPermission.Name,
		ResourceId:     QueryProbeAssignedPermission.ResourcePattern,
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

func (p *QueryProbe) runUnassignedPermission(ctx context.Context, logger lager.Logger) (bool, time.Duration, error) {
	logger = logger.Session("has-unassigned-permission").WithData(lager.Data{
		"actor.id":                    QueryProbeActor.GetID(),
		"actor.issuer":                QueryProbeActor.GetIssuer(),
		"permission.name":             QueryProbeUnassignedPermission.GetName(),
		"permission.resource_pattern": QueryProbeUnassignedPermission.GetResourcePattern(),
	})
	req := &protos.HasPermissionRequest{
		Actor:          QueryProbeActor,
		PermissionName: QueryProbeUnassignedPermission.Name,
		ResourceId:     QueryProbeUnassignedPermission.ResourcePattern,
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
