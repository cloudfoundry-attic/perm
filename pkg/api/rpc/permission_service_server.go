package rpc

import (
	"context"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm-go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"code.cloudfoundry.org/perm/pkg/api/logging"
	"code.cloudfoundry.org/perm/pkg/api/repos"
	"code.cloudfoundry.org/perm/pkg/perm"
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
	err := validateActor(pActor)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}
	actor := perm.Actor{
		ID:        pActor.GetID(),
		Namespace: pActor.GetNamespace(),
	}
	action := req.GetPermissionName()
	resourcePattern := req.GetResourceId()
	extensions := []logging.CustomExtension{
		{Key: "userID", Value: pActor.GetID()},
		{Key: "permission", Value: action},
		{Key: "resourceID", Value: resourcePattern},
	}

	s.securityLogger.Log(ctx, "HasPermission", "Permission check", extensions...)

	logger := s.logger.Session("has-permission").WithData(lager.Data{
		"actor.id":                   actor.ID,
		"actor.namespace":            actor.Namespace,
		"permission.action":          action,
		"permission.resourcePattern": resourcePattern,
	})
	logger.Debug(starting)

	query := repos.HasPermissionQuery{
		Actor:           actor,
		Action:          action,
		ResourcePattern: resourcePattern,
	}

	found, err := s.permissionRepo.HasPermission(ctx, logger, query)
	if err != nil {
		if err == perm.ErrRoleNotFound {
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
	err := validateActor(pActor)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}
	actor := perm.Actor{
		ID:        pActor.GetID(),
		Namespace: pActor.GetNamespace(),
	}
	action := req.GetPermissionName()

	logger := s.logger.Session("list-resource-patterns").
		WithData(lager.Data{
			"actor.id":          actor.ID,
			"actor.namespace":   actor.Namespace,
			"permission.action": action,
		})

	logger.Debug(starting)

	query := repos.ListResourcePatternsQuery{
		Actor:  actor,
		Action: action,
	}

	resourcePatterns, err := s.permissionRepo.ListResourcePatterns(ctx, logger, query)
	if err != nil {
		return nil, togRPCError(err)
	}

	var resourcePatternStrings []string

	for _, rp := range resourcePatterns {
		resourcePatternStrings = append(resourcePatternStrings, rp)
	}

	logger.Debug(success)

	return &protos.ListResourcePatternsResponse{
		ResourcePatterns: resourcePatternStrings,
	}, nil
}
