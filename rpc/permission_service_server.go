package rpc

import (
	"context"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm-go"
	"code.cloudfoundry.org/perm/messages"
	"code.cloudfoundry.org/perm/models"
)

type PermissionServiceServer struct {
	logger lager.Logger

	permissionService models.PermissionService
}

func NewPermissionServiceServer(
	logger lager.Logger,
	permissionService models.PermissionService,
) *PermissionServiceServer {
	return &PermissionServiceServer{
		logger:            logger,
		permissionService: permissionService,
	}
}

func (s *PermissionServiceServer) HasPermission(
	ctx context.Context,
	req *perm_go.HasPermissionRequest,
) (*perm_go.HasPermissionResponse, error) {
	pActor := req.GetActor()
	domainID := models.ActorDomainID(pActor.GetID())
	issuer := models.ActorIssuer(pActor.GetIssuer())

	permissionName := models.PermissionName(req.GetPermissionName())
	resourceID := models.ResourceID(req.GetResourceId())

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
			PermissionName: permissionName,
			ResourceID:     resourceID,
		},
	}

	found, err := s.permissionService.HasPermission(ctx, logger, query)
	if err != nil {
		if err == models.ErrRoleNotFound || err == models.ErrActorNotFound {
			return &perm_go.HasPermissionResponse{HasPermission: false}, nil
		}

		return nil, togRPCError(err)
	}

	logger.Debug(messages.Success)
	return &perm_go.HasPermissionResponse{HasPermission: found}, nil
}
