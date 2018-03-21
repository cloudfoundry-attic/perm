package monitor

import (
	"context"

	"code.cloudfoundry.org/lager"

	"code.cloudfoundry.org/perm-go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

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

func adminRoleName(s string) string {
	return AdminProbeRoleName + "." + s
}

func (p *AdminProbe) Cleanup(ctx context.Context, logger lager.Logger, uniqueSuffix string) error {
	logger.Debug(starting)
	defer logger.Debug(finished)

	roleName := adminRoleName(uniqueSuffix)

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

func (p *AdminProbe) Run(ctx context.Context, logger lager.Logger, uniqueSuffix string) error {
	logger.Debug(starting)
	defer logger.Debug(finished)

	roleName := adminRoleName(uniqueSuffix)
	var err error

	// CreateRole
	createRoleRequest := &protos.CreateRoleRequest{
		Name: roleName,
	}
	_, err = p.RoleServiceClient.CreateRole(ctx, createRoleRequest)
	if err != nil {
		logger.Error(failedToCreateRole, err, lager.Data{
			"roleName": createRoleRequest.GetName(),
		})

		return err
	}

	// AssignRole
	assignRoleRequest := &protos.AssignRoleRequest{
		RoleName: roleName,
		Actor:    AdminProbeActor,
	}
	_, err = p.RoleServiceClient.AssignRole(ctx, assignRoleRequest)
	if err != nil {
		logger.Error(failedToAssignRole, err, lager.Data{
			"roleName":     assignRoleRequest.GetRoleName(),
			"actor.ID":     assignRoleRequest.GetActor().GetID(),
			"actor.Issuer": assignRoleRequest.GetActor().GetIssuer(),
		})
		return err
	}

	// UnassignRole
	unassignRoleRequest := &protos.UnassignRoleRequest{
		Actor:    AdminProbeActor,
		RoleName: roleName,
	}
	_, err = p.RoleServiceClient.UnassignRole(ctx, unassignRoleRequest)
	if err != nil {
		logger.Error(failedToUnassignRole, err, lager.Data{
			"roleName":     unassignRoleRequest.GetRoleName(),
			"actor.ID":     unassignRoleRequest.GetActor().GetID(),
			"actor.Issuer": unassignRoleRequest.GetActor().GetIssuer(),
		})
		return err
	}

	// DeleteRole
	deleteRoleRequest := &protos.DeleteRoleRequest{
		Name: roleName,
	}
	_, err = p.RoleServiceClient.DeleteRole(ctx, deleteRoleRequest)
	if err != nil {
		logger.Error(failedToDeleteRole, err, lager.Data{
			"roleName": deleteRoleRequest.GetName(),
		})

		return err
	}

	return nil
}
