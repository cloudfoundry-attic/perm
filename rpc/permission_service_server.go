package rpc

import (
	"context"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/messages"
	"code.cloudfoundry.org/perm/models"
	"code.cloudfoundry.org/perm/protos"
)

type PermissionServiceServer struct {
	logger lager.Logger

	permissionService models.PermissionService
}

func NewPermissionServiceServer(logger lager.Logger, permissionService models.PermissionService) *PermissionServiceServer {
	return &PermissionServiceServer{
		logger:            logger,
		permissionService: permissionService,
	}
}

func (s *PermissionServiceServer) HasPermission(ctx context.Context, req *protos.HasPermissionRequest) (*protos.HasPermissionResponse, error) {
	pActor := req.GetActor()
	domainID := pActor.GetID()
	issuer := pActor.GetIssuer()

	permissionName := req.GetPermissionName()
	resourceID := req.GetResourceId()

	logger := s.logger.Session("has-role").WithData(lager.Data{
		"actor.id":              domainID,
		"actor.issuer":          issuer,
		"permission.name":       permissionName,
		"permission.resourceID": resourceID,
	})
	logger.Debug(messages.Starting)

	query := models.HasPermissionQuery{
		ActorQuery: models.ActorQuery{
			DomainID: domainID,
			Issuer:   issuer,
		},
		PermissionQuery: models.PermissionQuery{
			PermissionDefinitionQuery: models.PermissionDefinitionQuery{
				Name: permissionName,
			},
			ResourceID: resourceID,
		},
	}

	found, err := s.permissionService.HasPermission(ctx, logger, query)
	if err != nil {
		if err == models.ErrRoleNotFound || err == models.ErrActorNotFound {
			return &protos.HasPermissionResponse{HasPermission: false}, nil
		}

		return nil, togRPCError(err)
	}

	logger.Debug(messages.Success)
	return &protos.HasPermissionResponse{HasPermission: found}, nil
}
