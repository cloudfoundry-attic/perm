package rpc_test

import (
	"context"

	"code.cloudfoundry.org/lager/lagertest"
	"code.cloudfoundry.org/perm/api/internal/repos/inmemory"
	"code.cloudfoundry.org/perm/api/internal/repos/reposfakes"
	"code.cloudfoundry.org/perm/api/internal/rpc"
	"code.cloudfoundry.org/perm/internal/protos"
	"code.cloudfoundry.org/perm/logx"
	"code.cloudfoundry.org/perm/logx/lagerx"
	"code.cloudfoundry.org/perm/logx/logxfakes"

	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ = Describe("PermissionServiceServer", func() {
	var (
		subject           *rpc.PermissionServiceServer
		roleServiceServer *rpc.RoleServiceServer
		logger            logx.Logger
		securityLogger    *logxfakes.FakeSecurityLogger

		inMemoryStore *inmemory.Store

		ctx context.Context
	)

	BeforeEach(func() {
		logger = lagerx.NewLogger(lagertest.NewTestLogger("perm-test"))
		securityLogger = new(logxfakes.FakeSecurityLogger)
		inMemoryStore = inmemory.NewStore()

		ctx = context.Background()
		roleServiceServer = rpc.NewRoleServiceServer(logger, securityLogger, inMemoryStore)
		subject = rpc.NewPermissionServiceServer(logger, securityLogger, inMemoryStore)
	})

	Describe("#HasPermission", func() {
		var (
			roleName                 string
			actor                    *protos.Actor
			permission1, permission2 *protos.Permission
		)

		BeforeEach(func() {
			roleName = "role"
			actor = &protos.Actor{ID: "actor", Namespace: "namespace"}
			permission1 = &protos.Permission{
				Action:          "some-action",
				ResourcePattern: "some-resource",
			}
			permission2 = &protos.Permission{
				Action:          "some-other-action",
				ResourcePattern: "some-other-resource",
			}
		})

		It("returns true when there is a matching permission action and resource", func() {
			_, err := roleServiceServer.CreateRole(ctx, &protos.CreateRoleRequest{
				Name: roleName,
				Permissions: []*protos.Permission{
					permission1,
					permission2,
				},
			})
			Expect(err).NotTo(HaveOccurred())

			_, err = roleServiceServer.AssignRole(ctx, &protos.AssignRoleRequest{
				Actor:    actor,
				RoleName: roleName,
			})
			Expect(err).NotTo(HaveOccurred())

			res, err := subject.HasPermission(ctx, &protos.HasPermissionRequest{
				Actor:    actor,
				Action:   "some-other-action",
				Resource: "some-other-resource",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(res.GetHasPermission()).To(BeTrue())
		})

		It("returns false when mismatch the permission action and resource", func() {
			_, err := roleServiceServer.CreateRole(ctx, &protos.CreateRoleRequest{
				Name: roleName,
				Permissions: []*protos.Permission{
					permission1,
					permission2,
				},
			})
			Expect(err).NotTo(HaveOccurred())

			_, err = roleServiceServer.AssignRole(ctx, &protos.AssignRoleRequest{
				Actor:    actor,
				RoleName: roleName,
			})
			Expect(err).NotTo(HaveOccurred())

			res, err := subject.HasPermission(ctx, &protos.HasPermissionRequest{
				Actor:    actor,
				Action:   "some-permission",
				Resource: "some-other-resource",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(res.GetHasPermission()).To(BeFalse())
		})

		It("returns false if they have not been assigned the role", func() {
			_, err := roleServiceServer.CreateRole(ctx, &protos.CreateRoleRequest{
				Name: roleName,
				Permissions: []*protos.Permission{
					permission1,
				},
			})
			Expect(err).NotTo(HaveOccurred())

			res, err := subject.HasPermission(ctx, &protos.HasPermissionRequest{
				Actor:    actor,
				Action:   "some-action",
				Resource: "some-resource",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(res.GetHasPermission()).To(BeFalse())
		})

		It("returns false if they have no permissions", func() {
			_, err := roleServiceServer.CreateRole(ctx, &protos.CreateRoleRequest{
				Name:        roleName,
				Permissions: []*protos.Permission{},
			})
			Expect(err).NotTo(HaveOccurred())

			_, err = roleServiceServer.AssignRole(ctx, &protos.AssignRoleRequest{
				Actor:    actor,
				RoleName: roleName,
			})
			Expect(err).NotTo(HaveOccurred())

			res, err := subject.HasPermission(ctx, &protos.HasPermissionRequest{
				Actor:    actor,
				Action:   "some-action",
				Resource: "some-resource",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(res.GetHasPermission()).To(BeFalse())
		})

		It("fails when the actor namespace is not provided", func() {
			actorWithoutNamespace := &protos.Actor{
				ID:        "actor",
				Namespace: "",
			}

			res, err := subject.HasPermission(ctx, &protos.HasPermissionRequest{
				Actor:    actorWithoutNamespace,
				Action:   "some-action",
				Resource: "some-resource",
			})
			expectedErr := status.Errorf(codes.InvalidArgument, "actor namespace cannot be empty")
			Expect(res).To(BeNil())
			Expect(err).To(MatchError(expectedErr))
		})

		It("logs a security event", func() {
			_, err := subject.HasPermission(ctx, &protos.HasPermissionRequest{
				Actor:    actor,
				Action:   "some-action",
				Resource: "some-resource",
			})
			Expect(err).NotTo(HaveOccurred())
			expectedExtensions := []logx.SecurityData{
				{Key: "actorID", Value: "actor"},
				{Key: "actorNS", Value: "namespace"},
				{Key: "action", Value: "some-action"},
				{Key: "resource", Value: "some-resource"},
			}

			Expect(securityLogger.LogCallCount()).To(Equal(1))
			_, signature, name, extensions := securityLogger.LogArgsForCall(0)
			Expect(signature).To(Equal("HasPermission"))
			Expect(name).To(Equal("Permission check"))
			Expect(extensions).To(Equal(expectedExtensions))
		})

		Context("when there are groups provided to the request", func() {
			Context("when the actor does not permission", func() {
				It("returns true when the group has a matching permission action and resource", func() {
					_, err := roleServiceServer.CreateRole(ctx, &protos.CreateRoleRequest{
						Name: roleName,
						Permissions: []*protos.Permission{
							permission1,
							permission2,
						},
					})
					Expect(err).NotTo(HaveOccurred())

					group := protos.Group{ID: "some-group-id"}
					actor.Groups = append(actor.Groups, &group)
					_, err = roleServiceServer.AssignRoleToGroup(ctx, &protos.AssignRoleToGroupRequest{
						Group:    &group,
						RoleName: roleName,
					})
					Expect(err).NotTo(HaveOccurred())

					res, err := subject.HasPermission(ctx, &protos.HasPermissionRequest{
						Actor:    actor,
						Action:   "some-other-action",
						Resource: "some-other-resource",
					})
					Expect(err).NotTo(HaveOccurred())
					Expect(res.GetHasPermission()).To(BeTrue())
				})

				It("returns false when the group does not have a matching permission action and resource", func() {
					_, err := roleServiceServer.CreateRole(ctx, &protos.CreateRoleRequest{
						Name: roleName,
						Permissions: []*protos.Permission{
							permission1,
						},
					})
					Expect(err).NotTo(HaveOccurred())

					group := protos.Group{ID: "some-group-id"}
					actor.Groups = append(actor.Groups, &group)
					_, err = roleServiceServer.AssignRoleToGroup(ctx, &protos.AssignRoleToGroupRequest{
						Group:    &group,
						RoleName: roleName,
					})
					Expect(err).NotTo(HaveOccurred())

					res, err := subject.HasPermission(ctx, &protos.HasPermissionRequest{
						Actor:    actor,
						Action:   "some-other-action",
						Resource: "some-other-resource",
					})
					Expect(err).NotTo(HaveOccurred())
					Expect(res.GetHasPermission()).To(BeFalse())
				})
			})
		})
	})

	Describe("#ListResourcePatterns", func() {
		It("returns the list of resource patterns", func() {
			p1 := &protos.Permission{Action: "test-permission-action", ResourcePattern: "foo"}
			p2 := &protos.Permission{Action: "another-permission-action", ResourcePattern: "bar"}
			p3 := &protos.Permission{Action: "test-permission-action", ResourcePattern: "baz"}

			_, err := roleServiceServer.CreateRole(ctx, &protos.CreateRoleRequest{
				Name:        "r1",
				Permissions: []*protos.Permission{p1, p2},
			})
			Expect(err).NotTo(HaveOccurred())

			_, err = roleServiceServer.CreateRole(ctx, &protos.CreateRoleRequest{
				Name:        "r2",
				Permissions: []*protos.Permission{p3},
			})
			Expect(err).NotTo(HaveOccurred())

			actor := &protos.Actor{
				Namespace: "test-namespace",
				ID:        "fancy-id",
			}
			_, err = roleServiceServer.AssignRole(ctx, &protos.AssignRoleRequest{
				Actor:    actor,
				RoleName: "r1",
			})
			Expect(err).NotTo(HaveOccurred())

			_, err = roleServiceServer.AssignRole(ctx, &protos.AssignRoleRequest{
				Actor:    actor,
				RoleName: "r2",
			})
			Expect(err).NotTo(HaveOccurred())

			request := &protos.ListResourcePatternsRequest{
				Actor:  actor,
				Action: "test-permission-action",
			}
			response, err := subject.ListResourcePatterns(ctx, request)
			Expect(err).NotTo(HaveOccurred())
			Expect(response).NotTo(BeNil())
			Expect(response.ResourcePatterns).To(HaveLen(2))
			Expect(response.ResourcePatterns).To(ConsistOf("foo", "baz"))
		})

		It("returns an empty list if there are no resource patterns", func() {
			req := &protos.ListResourcePatternsRequest{
				Actor: &protos.Actor{
					ID:        "123",
					Namespace: "namespace34",
				},
				Action: "p12",
			}

			response, err := subject.ListResourcePatterns(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			Expect(response).NotTo(BeNil())
			Expect(response.ResourcePatterns).To(BeEmpty())
		})

		It("returns a relevant error if the query fails", func() {
			permissionRepo := new(reposfakes.FakePermissionRepo)
			subject = rpc.NewPermissionServiceServer(logger, securityLogger, permissionRepo)
			testErr := errors.New("test-error")
			permissionRepo.ListResourcePatternsReturns(nil, testErr)

			req := &protos.ListResourcePatternsRequest{
				Actor: &protos.Actor{
					ID:        "123",
					Namespace: "namespace34",
				},
				Action: "p12",
			}
			_, err := subject.ListResourcePatterns(ctx, req)
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(status.Errorf(codes.Unknown, "test-error")))
		})

		It("fails when the actor namespace is not provided", func() {
			actor := &protos.Actor{
				ID:        "actor",
				Namespace: "",
			}
			res, err := subject.ListResourcePatterns(ctx, &protos.ListResourcePatternsRequest{
				Actor: actor,
			})

			expectedErr := status.Errorf(codes.InvalidArgument, "actor namespace cannot be empty")
			Expect(res).To(BeNil())
			Expect(err).To(MatchError(expectedErr))
		})

		It("logs a security event", func() {
			actor := &protos.Actor{
				ID:        "actor",
				Namespace: "actor-namespace",
			}
			_, err := subject.ListResourcePatterns(ctx, &protos.ListResourcePatternsRequest{
				Actor:  actor,
				Action: "some-action",
			})
			Expect(err).NotTo(HaveOccurred())
			expectedExtensions := []logx.SecurityData{
				{Key: "actorID", Value: "actor"},
				{Key: "actorNS", Value: "actor-namespace"},
				{Key: "action", Value: "some-action"},
			}

			Expect(securityLogger.LogCallCount()).To(Equal(1))
			_, signature, name, extensions := securityLogger.LogArgsForCall(0)
			Expect(signature).To(Equal("ListResourcePatterns"))
			Expect(name).To(Equal("Resource pattern list"))
			Expect(extensions).To(Equal(expectedExtensions))
		})

		Context("when there are groups provided to the request", func() {
			var (
				actor                                 *protos.Actor
				role1Name, role2Name                  string
				action1, action2                      string
				resource1, resource2, resource3       string
				permission1, permission2, permission3 *protos.Permission
			)

			BeforeEach(func() {
				actor = &protos.Actor{ID: "actor", Namespace: "namespace"}
				role1Name = "role1"
				role2Name = "role2"

				action1 = "some-action"
				action2 = "some-other-action"

				resource1 = "some-resource-1"
				resource2 = "some-resource-2"
				resource3 = "some-resource-3"

				permission1 = &protos.Permission{
					Action:          action1,
					ResourcePattern: resource1,
				}
				permission2 = &protos.Permission{
					Action:          action2,
					ResourcePattern: resource2,
				}
				permission3 = &protos.Permission{
					Action:          action2,
					ResourcePattern: resource3,
				}
			})

			It("returns a list of resource patterns from both actor and groups for the action", func() {
				_, err := roleServiceServer.CreateRole(ctx, &protos.CreateRoleRequest{
					Name: role1Name,
					Permissions: []*protos.Permission{
						permission1,
						permission2,
					},
				})
				Expect(err).NotTo(HaveOccurred())

				_, err = roleServiceServer.AssignRole(ctx, &protos.AssignRoleRequest{
					Actor:    actor,
					RoleName: role1Name,
				})
				Expect(err).NotTo(HaveOccurred())

				_, err = roleServiceServer.CreateRole(ctx, &protos.CreateRoleRequest{
					Name: role2Name,
					Permissions: []*protos.Permission{
						permission3,
					},
				})
				Expect(err).NotTo(HaveOccurred())

				group := protos.Group{ID: "some-group-id"}
				actor.Groups = append(actor.Groups, &group)
				_, err = roleServiceServer.AssignRoleToGroup(ctx, &protos.AssignRoleToGroupRequest{
					Group:    &group,
					RoleName: role2Name,
				})
				Expect(err).NotTo(HaveOccurred())

				res, err := subject.ListResourcePatterns(ctx, &protos.ListResourcePatternsRequest{
					Actor:  actor,
					Action: action2,
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(res.ResourcePatterns).To(HaveLen(2))
				Expect(res.ResourcePatterns).To(ConsistOf(resource2, resource3))
			})
		})
	})
})
