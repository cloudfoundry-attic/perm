package rpc

import (
	"context"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm-go"
	"code.cloudfoundry.org/perm/messages"
	"code.cloudfoundry.org/perm/models"
	"code.cloudfoundry.org/perm/repos"
)

type PermissionServiceServer struct {
	logger lager.Logger

	permissionRepo repos.PermissionRepo
}

func NewPermissionServiceServer(
	logger lager.Logger,
	permissionRepo repos.PermissionRepo,
) *PermissionServiceServer {
	return &PermissionServiceServer{
		logger:         logger,
		permissionRepo: permissionRepo,
	}
}

func (s *PermissionServiceServer) HasPermission(
	ctx context.Context,
	req *protos.HasPermissionRequest,
) (*protos.HasPermissionResponse, error) {
	pActor := req.GetActor()
	domainID := models.ActorDomainID(pActor.GetID())
	issuer := models.ActorIssuer(pActor.GetIssuer())

	permissionName := models.PermissionName(req.GetPermissionName())
	resourcePattern := models.PermissionResourcePattern(req.GetResourceId())

	logger := s.logger.Session("has-role").WithData(lager.Data{
		"actor.id":                   domainID,
		"actor.issuer":               issuer,
		"permission.name":            permissionName,
		"permission.resourcePattern": resourcePattern,
	})
	logger.Debug(messages.Starting)

	query := repos.HasPermissionQuery{
		ActorQuery: repos.ActorQuery{
			DomainID: domainID,
			Issuer:   issuer,
		},
		PermissionQuery: repos.PermissionQuery{
			PermissionName:  permissionName,
			ResourcePattern: resourcePattern,
		},
	}

	found, err := s.permissionRepo.HasPermission(ctx, logger, query)
	if err != nil {
		if err == models.ErrRoleNotFound {
			return &protos.HasPermissionResponse{HasPermission: false}, nil
		}

		return nil, togRPCError(err)
	}

	logger.Debug(messages.Success)
	return &protos.HasPermissionResponse{HasPermission: found}, nil
}

func (s *PermissionServiceServer) ListResourcePatterns(
	ctx context.Context,
	req *protos.ListResourcePatternsRequest,
) (*protos.ListResourcePatternsResponse, error) {
	pActor := req.GetActor()
	domainID := models.ActorDomainID(pActor.GetID())
	issuer := models.ActorIssuer(pActor.GetIssuer())

	permissionName := models.PermissionName(req.GetPermissionName())

	logger := s.logger.Session("list-resource-patterns").
		WithData(lager.Data{
			"actor.id":        domainID,
			"actor.issuer":    issuer,
			"permission.name": permissionName,
		})

	logger.Debug(messages.Starting)

	query := repos.ListResourcePatternsQuery{
		ActorQuery: repos.ActorQuery{
			DomainID: models.ActorDomainID(req.Actor.ID),
			Issuer:   models.ActorIssuer(req.Actor.Issuer),
		},
		PermissionDefinitionQuery: repos.PermissionDefinitionQuery{
			Name: models.PermissionName(req.PermissionName),
		},
	}

	resourcePatterns, err := s.permissionRepo.ListResourcePatterns(ctx, logger, query)
	if err != nil {
		return nil, togRPCError(err)
	}

	var resourcePatternStrings []string

	for _, rp := range resourcePatterns {
		resourcePatternStrings = append(resourcePatternStrings, string(rp))
	}

	logger.Debug(messages.Success)

	return &protos.ListResourcePatternsResponse{
		ResourcePatterns: resourcePatternStrings,
	}, nil
}
