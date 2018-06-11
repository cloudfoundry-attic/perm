package rpc

import (
	"context"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/pkg/api/protos"
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
	action := req.GetAction()
	groups := make([]perm.Group, 0)
	for _, group := range req.GetGroups() {
		groups = append(groups, perm.Group{
			ID: group.GetID(),
		})
	}
	resourcePattern := req.GetResource()
	extensions := []logging.CustomExtension{
		{Key: "userID", Value: pActor.GetID()},
		{Key: "action", Value: action},
		{Key: "resource", Value: resourcePattern},
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
		Groups:          groups,
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

	pGroups := req.GetGroups()
	groups := []perm.Group{}
	for _, pGroup := range pGroups {
		groups = append(groups, perm.Group{ID: pGroup.GetID()})
	}

	action := req.GetAction()

	groupsStr := ""
	for _, g := range groups {
		groupsStr = groupsStr + ", " + g.ID
	}
	logger := s.logger.Session("list-resource-patterns").
		WithData(lager.Data{
			"actor.id":          actor.ID,
			"actor.namespace":   actor.Namespace,
			"groups":            groupsStr,
			"permission.action": action,
		})

	logger.Debug(starting)

	query := repos.ListResourcePatternsQuery{
		Actor:  actor,
		Action: action,
		Groups: groups,
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
