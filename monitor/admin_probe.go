package monitor

import (
	"context"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/messages"
	"code.cloudfoundry.org/perm/protos"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

//go:generate counterfeiter code.cloudfoundry.org/perm/protos.RoleServiceClient

type AdminProbe struct {
	RoleServiceClient protos.RoleServiceClient
}

const (
	AdminProbeRoleName = "system.admin-probe"
)

var AdminProbeActor = &protos.Actor{
	ID:     "admin-probe",
	Issuer: "system",
}

func (p *AdminProbe) Cleanup(ctx context.Context, logger lager.Logger) error {
	logger.Debug(messages.Starting)
	defer logger.Debug(messages.Finished)

	deleteRoleRequest := &protos.DeleteRoleRequest{
		Name: AdminProbeRoleName,
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

func (p *AdminProbe) Run(ctx context.Context, logger lager.Logger) error {
	logger.Debug(messages.Starting)
	defer logger.Debug(messages.Finished)

	var err error

	// CreateRole
	createRoleRequest := &protos.CreateRoleRequest{
		Name: AdminProbeRoleName,
	}
	_, err = p.RoleServiceClient.CreateRole(ctx, createRoleRequest)
	if err != nil {
		logger.Error(messages.FailedToCreateRole, err, lager.Data{
			"roleName": createRoleRequest.GetName(),
		})

		return err
	}

	// AssignRole
	assignRoleRequest := &protos.AssignRoleRequest{
		RoleName: AdminProbeRoleName,
		Actor:    AdminProbeActor,
	}
	_, err = p.RoleServiceClient.AssignRole(ctx, assignRoleRequest)
	if err != nil {
		logger.Error(messages.FailedToAssignRole, err, lager.Data{
			"roleName":     assignRoleRequest.GetRoleName(),
			"actor.ID":     assignRoleRequest.GetActor().GetID(),
			"actor.Issuer": assignRoleRequest.GetActor().GetIssuer(),
		})
		return err
	}

	// UnassignRole
	unassignRoleRequest := &protos.UnassignRoleRequest{
		Actor:    AdminProbeActor,
		RoleName: AdminProbeRoleName,
	}
	_, err = p.RoleServiceClient.UnassignRole(ctx, unassignRoleRequest)
	if err != nil {
		logger.Error(messages.FailedToUnassignRole, err, lager.Data{
			"roleName":     unassignRoleRequest.GetRoleName(),
			"actor.ID":     unassignRoleRequest.GetActor().GetID(),
			"actor.Issuer": unassignRoleRequest.GetActor().GetIssuer(),
		})
		return err
	}

	// DeleteRole
	deleteRoleRequest := &protos.DeleteRoleRequest{
		Name: AdminProbeRoleName,
	}
	_, err = p.RoleServiceClient.DeleteRole(ctx, deleteRoleRequest)
	if err != nil {
		logger.Error(messages.FailedToDeleteRole, err, lager.Data{
			"roleName": deleteRoleRequest.GetName(),
		})

		return err
	}

	return nil
}
