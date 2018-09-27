package rpc

import (
	"context"

	"code.cloudfoundry.org/perm"
	"code.cloudfoundry.org/perm/api/internal/repos"
	"code.cloudfoundry.org/perm/internal/protos"
	"code.cloudfoundry.org/perm/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PermissionServiceServer struct {
	logger         logx.Logger
	securityLogger logx.SecurityLogger
	permissionRepo repos.PermissionRepo
}

func NewPermissionServiceServer(
	logger logx.Logger,
	securityLogger logx.SecurityLogger,
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

	for _, group := range pActor.GetGroups() {
		actor.Groups = append(actor.Groups, perm.Group{
			ID: group.GetID(),
		})
	}

	resourcePattern := req.GetResource()
	extensions := []logx.SecurityData{
		{Key: "actorID", Value: pActor.GetID()},
		{Key: "actorNS", Value: pActor.GetNamespace()},
		{Key: "action", Value: action},
		{Key: "resource", Value: resourcePattern},
	}

	s.securityLogger.Log(ctx, "HasPermission", "Permission check", extensions...)

	logger := s.logger.WithName("has-permission").WithData(
		logx.Data{Key: "actor.id", Value: actor.ID},
		logx.Data{Key: "actor.namespace", Value: actor.Namespace},
		logx.Data{Key: "permission.action", Value: action},
		logx.Data{Key: "permission.resourcePattern", Value: resourcePattern},
	)
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

	for _, group := range pActor.GetGroups() {
		actor.Groups = append(actor.Groups, perm.Group{ID: group.GetID()})
	}

	action := req.GetAction()

	groupsStr := ""
	for _, g := range actor.Groups {
		groupsStr = groupsStr + ", " + g.ID
	}

	extensions := []logx.SecurityData{
		{Key: "actorID", Value: pActor.GetID()},
		{Key: "actorNS", Value: pActor.GetNamespace()},
		{Key: "action", Value: action},
	}

	s.securityLogger.Log(ctx, "ListResourcePatterns", "Resource pattern list", extensions...)

	logger := s.logger.WithName("list-resource-patterns").WithData(
		logx.Data{Key: "actor.id", Value: actor.ID},
		logx.Data{Key: "actor.namespace", Value: actor.Namespace},
		logx.Data{Key: "groups", Value: groupsStr},
		logx.Data{Key: "permission.action", Value: action},
	)

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
