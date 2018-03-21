package rpc

import (
	"context"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm-go"

	"code.cloudfoundry.org/perm/pkg/api/logging"
	"code.cloudfoundry.org/perm/pkg/api/models"
	"code.cloudfoundry.org/perm/pkg/api/repos"
)

type PermissionServiceServer struct {
	logger         lager.Logger
	securityLogger SecurityLogger
	permissionRepo repos.PermissionRepo
}

func NewPermissionServiceServer(
	logger lager.Logger,
	securityLogger SecurityLogger,
	permissionRepo repos.PermissionRepo,
) *PermissionServiceServer {
	return &PermissionServiceServer{
		logger:         logger,
		securityLogger: securityLogger,
		permissionRepo: permissionRepo,
	}
}

func (s *PermissionServiceServer) HasPermission(
	ctx context.Context,
	req *protos.HasPermissionRequest,
) (*protos.HasPermissionResponse, error) {
	pActor := req.GetActor()
	actor := models.Actor{
		DomainID: models.ActorDomainID(pActor.GetID()),
		Issuer:   models.ActorIssuer(pActor.GetIssuer()),
	}
	permissionName := models.PermissionName(req.GetPermissionName())
	resourcePattern := models.PermissionResourcePattern(req.GetResourceId())
	extensions := []logging.CustomExtension{
		{Key: "userID", Value: pActor.GetID()},
		{Key: "permission", Value: string(permissionName)},
		{Key: "resourceID", Value: string(resourcePattern)},
	}

	s.securityLogger.Log(ctx, "HasPermission", "Permission check", extensions...)

	logger := s.logger.Session("has-role").WithData(lager.Data{
		"actor.id":                   actor.DomainID,
		"actor.issuer":               actor.Issuer,
		"permission.name":            permissionName,
		"permission.resourcePattern": resourcePattern,
	})
	logger.Debug(starting)

	query := repos.HasPermissionQuery{
		Actor:           actor,
		PermissionName:  permissionName,
		ResourcePattern: resourcePattern,
	}

	found, err := s.permissionRepo.HasPermission(ctx, logger, query)
	if err != nil {
		if err == models.ErrRoleNotFound {
			return &protos.HasPermissionResponse{HasPermission: false}, nil
		}

		return nil, togRPCError(err)
	}

	logger.Debug(success)
	return &protos.HasPermissionResponse{HasPermission: found}, nil
}

func (s *PermissionServiceServer) ListResourcePatterns(
	ctx context.Context,
	req *protos.ListResourcePatternsRequest,
) (*protos.ListResourcePatternsResponse, error) {
	pActor := req.GetActor()
	actor := models.Actor{
		DomainID: models.ActorDomainID(pActor.GetID()),
		Issuer:   models.ActorIssuer(pActor.GetIssuer()),
	}
	permissionName := models.PermissionName(req.GetPermissionName())

	logger := s.logger.Session("list-resource-patterns").
		WithData(lager.Data{
			"actor.id":        actor.DomainID,
			"actor.issuer":    actor.Issuer,
			"permission.name": permissionName,
		})

	logger.Debug(starting)

	query := repos.ListResourcePatternsQuery{
		Actor:          actor,
		PermissionName: permissionName,
	}

	resourcePatterns, err := s.permissionRepo.ListResourcePatterns(ctx, logger, query)
	if err != nil {
		return nil, togRPCError(err)
	}

	var resourcePatternStrings []string

	for _, rp := range resourcePatterns {
		resourcePatternStrings = append(resourcePatternStrings, string(rp))
	}

	logger.Debug(success)

	return &protos.ListResourcePatternsResponse{
		ResourcePatterns: resourcePatternStrings,
	}, nil
}
