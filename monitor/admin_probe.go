package monitor

import (
	"context"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/messages"
	"code.cloudfoundry.org/perm/protos"
	"github.com/cactus/go-statsd-client/statsd"
	multierror "github.com/hashicorp/go-multierror"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AdminProbe struct {
	RoleServiceClient protos.RoleServiceClient
	StatsDClient      statsd.Statter
}

const (
	AlwaysSendMetric = 1.0

	AdminProbeRoleName = "system.admin-probe"
)

var AdminProbeActor = &protos.Actor{
	ID:     "admin-probe",
	Issuer: "system",
}

func (p *AdminProbe) Cleanup(ctx context.Context, logger lager.Logger) error {
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
	var result error
	var err error

	defer func() {
		err = p.StatsDClient.Inc("perm.probe.admin.runs.total", 1, AlwaysSendMetric)

		if err != nil {
			logger.Error(messages.FailedToSendMetric, err, lager.Data{
				"metric": "perm.probe.admin.runs.total",
			})
			result = multierror.Append(result, err)
		}
	}()

	// CreateRole
	createRoleRequest := &protos.CreateRoleRequest{
		Name: AdminProbeRoleName,
	}
	_, err = p.RoleServiceClient.CreateRole(ctx, createRoleRequest)
	if err != nil {
		logger.Error(messages.FailedToCreateRole, err, lager.Data{
			"roleName": createRoleRequest.GetName(),
		})
		result = multierror.Append(result, err)

		err = p.StatsDClient.Inc("perm.probe.admin.runs.failed", 1, AlwaysSendMetric)
		if err != nil {
			logger.Error(messages.FailedToSendMetric, err, lager.Data{
				"metric": "perm.probe.admin.runs.failed",
			})
			result = multierror.Append(result, err)
		}

		return result
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
		result = multierror.Append(result, err)

		err = p.StatsDClient.Inc("perm.probe.admin.runs.failed", 1, AlwaysSendMetric)
		if err != nil {
			logger.Error(messages.FailedToSendMetric, err, lager.Data{
				"metric": "perm.probe.admin.runs.failed",
			})
			result = multierror.Append(result, err)
		}

		return result
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
		result = multierror.Append(result, err)

		err = p.StatsDClient.Inc("perm.probe.admin.runs.failed", 1, AlwaysSendMetric)
		if err != nil {
			logger.Error(messages.FailedToSendMetric, err, lager.Data{
				"metric": "perm.probe.admin.runs.failed",
			})
			result = multierror.Append(result, err)
		}
		return result
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
		result = multierror.Append(result, err)

		err = p.StatsDClient.Inc("perm.probe.admin.runs.failed", 1, AlwaysSendMetric)
		if err != nil {
			logger.Error(messages.FailedToSendMetric, err, lager.Data{
				"metric": "perm.probe.admin.runs.failed",
			})
			result = multierror.Append(result, err)
		}
		return result
	}

	return result
}
