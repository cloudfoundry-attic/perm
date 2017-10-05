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

//go:generate counterfeiter code.cloudfoundry.org/perm/protos.RoleServiceClient
//go:generate counterfeiter github.com/cactus/go-statsd-client/statsd.Statter

type AdminProbe struct {
	RoleServiceClient protos.RoleServiceClient
	StatsDClient      statsd.Statter
}

const (
	AlwaysSendMetric = 1.0

	AdminProbeRoleName = "system.admin-probe"

	MetricAdminProbeRunsTotal  = "perm.probe.admin.runs.total"
	MetricAdminProbeRunsFailed = "perm.probe.admin.runs.failed"
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
		err = p.IncrementMetricRunsTotal()

		if err != nil {
			logger.Error(messages.FailedToSendMetric, err, lager.Data{
				"metric": MetricAdminProbeRunsTotal,
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

		err = p.IncrementMetricRunsFailed()
		if err != nil {
			logger.Error(messages.FailedToSendMetric, err, lager.Data{
				"metric": MetricAdminProbeRunsFailed,
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

		err = p.IncrementMetricRunsFailed()
		if err != nil {
			logger.Error(messages.FailedToSendMetric, err, lager.Data{
				"metric": MetricAdminProbeRunsFailed,
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

		err = p.IncrementMetricRunsFailed()
		if err != nil {
			logger.Error(messages.FailedToSendMetric, err, lager.Data{
				"metric": MetricAdminProbeRunsFailed,
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

		err = p.IncrementMetricRunsFailed()
		if err != nil {
			logger.Error(messages.FailedToSendMetric, err, lager.Data{
				"metric": MetricAdminProbeRunsFailed,
			})
			result = multierror.Append(result, err)
		}
		return result
	}

	return result
}

func (p *AdminProbe) IncrementMetricRunsTotal() error {
	return p.StatsDClient.Inc(MetricAdminProbeRunsTotal, 1, AlwaysSendMetric)
}

func (p *AdminProbe) IncrementMetricRunsFailed() error {
	return p.StatsDClient.Inc(MetricAdminProbeRunsFailed, 1, AlwaysSendMetric)
}
