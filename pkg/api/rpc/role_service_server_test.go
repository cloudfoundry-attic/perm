package rpc_test

import (
	"context"

	"code.cloudfoundry.org/lager/lagertest"
	"code.cloudfoundry.org/perm/pkg/api/protos"
	"code.cloudfoundry.org/perm/pkg/api/rpc"
	"code.cloudfoundry.org/perm/pkg/logx"
	"code.cloudfoundry.org/perm/pkg/logx/cef"
	"code.cloudfoundry.org/perm/pkg/logx/lagerx"
	"code.cloudfoundry.org/perm/pkg/logx/logxfakes"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("RoleRepoServer", func() {
	var (
		subject        *rpc.RoleServiceServer
		logger         logx.Logger
		securityLogger *logxfakes.FakeSecurityLogger

		inMemoryStore *rpc.InMemoryStore

		ctx context.Context
	)

	BeforeEach(func() {
		logger = lagerx.NewLogger(lagertest.NewTestLogger("perm-test"))
		securityLogger = new(logxfakes.FakeSecurityLogger)
		inMemoryStore = rpc.NewInMemoryStore()

		ctx = context.Background()

		subject = rpc.NewRoleServiceServer(logger, securityLogger, inMemoryStore)
	})

	Describe("#CreateRole", func() {
		var req *protos.CreateRoleRequest

		BeforeEach(func() {
			req = &protos.CreateRoleRequest{
				Name: "test-role",
			}
		})

		It("succeeds if no role with that name exists", func() {
			res, err := subject.CreateRole(ctx, req)

			Expect(err).NotTo(HaveOccurred())
			Expect(res).NotTo(BeNil())
		})

		It("fails if a role with that name already exists", func() {
			_, err := subject.CreateRole(ctx, req)

			Expect(err).NotTo(HaveOccurred())

			res, err := subject.CreateRole(ctx, req)

			Expect(res).To(BeNil())
			Expect(err).To(HaveOccurred())
		})

		It("logs a security event", func() {
			_, err := subject.CreateRole(ctx, req)
			expectedExtensions := []cef.CustomExtension{{Key: "roleName", Value: "test-role"}}

			Expect(err).NotTo(HaveOccurred())
			Expect(securityLogger.LogCallCount()).To(Equal(1))
			_, signature, name, extensions := securityLogger.LogArgsForCall(0)
			Expect(signature).To(Equal("CreateRole"))
			Expect(name).To(Equal("Role creation"))
			Expect(extensions).To(Equal(expectedExtensions))
		})
	})

	Describe("#DeleteRole", func() {
		It("deletes the role if it exists", func() {
			name := "test-role"
			_, err := subject.CreateRole(ctx, &protos.CreateRoleRequest{
				Name: name,
				Permissions: []*protos.Permission{
					{"a", "b"},
				},
			})
			Expect(err).NotTo(HaveOccurred())

			resp, err := subject.ListRolePermissions(ctx, &protos.ListRolePermissionsRequest{
				RoleName: name,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(len(resp.GetPermissions())).To(Equal(1))

			res, err := subject.DeleteRole(ctx, &protos.DeleteRoleRequest{
				Name: name,
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(res).NotTo(BeNil())

			resp, err = subject.ListRolePermissions(ctx, &protos.ListRolePermissionsRequest{
				RoleName: name,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(res).NotTo(BeNil())
			Expect(len(resp.GetPermissions())).To(Equal(0))
		})

		It("fails if the role does not exist", func() {
			res, err := subject.DeleteRole(ctx, &protos.DeleteRoleRequest{
				Name: "test-role",
			})

			Expect(res).To(BeNil())
			Expect(err).To(HaveOccurred())
		})

		It("deletes any role assignments for the role", func() {
			name := "test-role"

			_, err := subject.CreateRole(ctx, &protos.CreateRoleRequest{
				Name: name,
			})

			Expect(err).NotTo(HaveOccurred())

			actor := &protos.Actor{
				ID:        "actor-id",
				Namespace: "namespace",
			}

			_, err = subject.AssignRole(ctx, &protos.AssignRoleRequest{
				Actor:    actor,
				RoleName: name,
			})

			Expect(err).NotTo(HaveOccurred())

			hasRoleRes, err := subject.HasRole(ctx, &protos.HasRoleRequest{
				Actor:    actor,
				RoleName: name,
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(hasRoleRes).NotTo(BeNil())
			Expect(hasRoleRes.GetHasRole()).To(BeTrue())

			res, err := subject.DeleteRole(ctx, &protos.DeleteRoleRequest{
				Name: name,
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(res).NotTo(BeNil())

			hasRoleRes, err = subject.HasRole(ctx, &protos.HasRoleRequest{
				Actor:    actor,
				RoleName: name,
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(hasRoleRes).NotTo(BeNil())
			Expect(hasRoleRes.GetHasRole()).To(BeFalse())
		})

		It("logs a security event", func() {
			req := &protos.DeleteRoleRequest{
				Name: "test-role",
			}
			expectedExtensions := []cef.CustomExtension{{Key: "roleName", Value: "test-role"}}
			subject.DeleteRole(ctx, req)

			Expect(securityLogger.LogCallCount()).To(Equal(1))
			_, signature, name, extensions := securityLogger.LogArgsForCall(0)
			Expect(signature).To(Equal("DeleteRole"))
			Expect(name).To(Equal("Role deletion"))
			Expect(extensions).To(Equal(expectedExtensions))
		})
	})

	Describe("#AssignRole", func() {
		It("succeeds if the role exists", func() {
			name := "role"
			actor := &protos.Actor{
				ID:        "actor-id",
				Namespace: "fake-namespace",
			}
			_, err := subject.CreateRole(ctx, &protos.CreateRoleRequest{
				Name: name,
			})

			Expect(err).NotTo(HaveOccurred())

			req := &protos.AssignRoleRequest{
				Actor:    actor,
				RoleName: name,
			}
			res, err := subject.AssignRole(ctx, req)

			Expect(err).NotTo(HaveOccurred())
			Expect(res).NotTo(BeNil())
		})

		It("fails if the user has already been assigned the role", func() {
			name := "role"
			actor := &protos.Actor{
				ID:        "actor-id",
				Namespace: "fake-namespace",
			}
			_, err := subject.CreateRole(ctx, &protos.CreateRoleRequest{
				Name: name,
			})

			Expect(err).NotTo(HaveOccurred())

			req := &protos.AssignRoleRequest{
				Actor:    actor,
				RoleName: name,
			}
			_, err = subject.AssignRole(ctx, req)

			Expect(err).NotTo(HaveOccurred())

			res, err := subject.AssignRole(ctx, req)

			Expect(err).To(HaveOccurred())
			Expect(res).To(BeNil())
		})

		It("fails if the role does not exist", func() {
			actor := &protos.Actor{
				ID:        "actor",
				Namespace: "namespace",
			}
			res, err := subject.AssignRole(ctx, &protos.AssignRoleRequest{
				Actor:    actor,
				RoleName: "does-not-exist",
			})

			Expect(res).To(BeNil())
			Expect(err).To(HaveOccurred())
		})

		It("fails when the actor namespace is not provided", func() {
			actor := &protos.Actor{
				ID:        "actor",
				Namespace: "",
			}
			res, err := subject.AssignRole(ctx, &protos.AssignRoleRequest{
				Actor:    actor,
				RoleName: "role",
			})

			expectedErr := status.Errorf(codes.InvalidArgument, "actor namespace cannot be empty")
			Expect(res).To(BeNil())
			Expect(err).To(MatchError(expectedErr))
		})

		It("logs a security event", func() {
			actor := &protos.Actor{
				ID:        "actor-id",
				Namespace: "fake-namespace",
			}
			req := &protos.AssignRoleRequest{
				Actor:    actor,
				RoleName: "role",
			}
			subject.AssignRole(ctx, req)
			expectedExtensions := []cef.CustomExtension{
				{Key: "roleName", Value: "role"},
				{Key: "userID", Value: "actor-id"},
			}

			Expect(securityLogger.LogCallCount()).To(Equal(1))
			_, signature, name, extensions := securityLogger.LogArgsForCall(0)
			Expect(signature).To(Equal("AssignRole"))
			Expect(name).To(Equal("Role assignment"))
			Expect(extensions).To(Equal(expectedExtensions))
		})
	})

	Describe("#UnassignRole", func() {
		It("removes role binding if the user has that role", func() {
			name := "role"
			actor := &protos.Actor{
				ID:        "actor-id",
				Namespace: "fake-namespace",
			}
			_, err := subject.CreateRole(ctx, &protos.CreateRoleRequest{
				Name: name,
			})

			Expect(err).NotTo(HaveOccurred())

			_, err = subject.AssignRole(ctx, &protos.AssignRoleRequest{
				Actor:    actor,
				RoleName: name,
			})

			Expect(err).NotTo(HaveOccurred())

			req := &protos.UnassignRoleRequest{
				Actor:    actor,
				RoleName: name,
			}
			res, err := subject.UnassignRole(ctx, req)

			Expect(err).NotTo(HaveOccurred())
			Expect(res).NotTo(BeNil())
		})

		It("fails if the user is not assigned to the role", func() {
			name := "role"
			actor := &protos.Actor{
				ID:        "actor",
				Namespace: "namespace",
			}
			_, err := subject.CreateRole(ctx, &protos.CreateRoleRequest{
				Name: name,
			})

			Expect(err).NotTo(HaveOccurred())

			req := &protos.UnassignRoleRequest{
				Actor:    actor,
				RoleName: name,
			}
			res, err := subject.UnassignRole(ctx, req)

			Expect(err).To(HaveOccurred())
			Expect(res).To(BeNil())
		})

		It("fails if the role does not exist", func() {
			name := "fake-role"
			actor := &protos.Actor{
				ID:        "actor",
				Namespace: "namespace",
			}
			req := &protos.UnassignRoleRequest{
				Actor:    actor,
				RoleName: name,
			}
			res, err := subject.UnassignRole(ctx, req)

			Expect(err).To(HaveOccurred())
			Expect(res).To(BeNil())
		})

		It("fails when the actor namespace is not provided", func() {
			actor := &protos.Actor{
				ID:        "actor",
				Namespace: "",
			}
			res, err := subject.UnassignRole(ctx, &protos.UnassignRoleRequest{
				Actor:    actor,
				RoleName: "role",
			})

			expectedErr := status.Errorf(codes.InvalidArgument, "actor namespace cannot be empty")
			Expect(res).To(BeNil())
			Expect(err).To(MatchError(expectedErr))
		})

		It("logs a security event", func() {
			actor := &protos.Actor{
				ID:        "actor-id",
				Namespace: "fake-namespace",
			}
			req := &protos.UnassignRoleRequest{
				Actor:    actor,
				RoleName: "role",
			}
			subject.UnassignRole(ctx, req)
			expectedExtensions := []cef.CustomExtension{
				{Key: "roleName", Value: "role"},
				{Key: "userID", Value: "actor-id"},
			}

			Expect(securityLogger.LogCallCount()).To(Equal(1))
			_, signature, name, extensions := securityLogger.LogArgsForCall(0)
			Expect(signature).To(Equal("UnassignRole"))
			Expect(name).To(Equal("Role unassignment"))
			Expect(extensions).To(Equal(expectedExtensions))
		})
	})

	Describe("#UnassignRoleFromGroup", func() {
		It("removes role binding if the group has that role", func() {
			name := "role"
			group := &protos.Group{
				ID: "group-id",
			}
			_, err := subject.CreateRole(ctx, &protos.CreateRoleRequest{
				Name: name,
			})

			Expect(err).NotTo(HaveOccurred())

			_, err = subject.AssignRoleToGroup(ctx, &protos.AssignRoleToGroupRequest{
				Group:    group,
				RoleName: name,
			})
			Expect(err).NotTo(HaveOccurred())

			hasRoleResp, err := subject.HasRoleForGroup(ctx, &protos.HasRoleForGroupRequest{
				Group:    group,
				RoleName: name,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(hasRoleResp.GetHasRole()).To(BeTrue())

			req := &protos.UnassignRoleFromGroupRequest{
				Group:    group,
				RoleName: name,
			}
			res, err := subject.UnassignRoleFromGroup(ctx, req)

			Expect(err).NotTo(HaveOccurred())
			Expect(res).NotTo(BeNil())

			hasRoleResp, err = subject.HasRoleForGroup(ctx, &protos.HasRoleForGroupRequest{
				Group:    group,
				RoleName: name,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(hasRoleResp.GetHasRole()).To(BeFalse())

		})

		It("fails if the user is not assigned to the role", func() {
			name := "role"
			group := &protos.Group{
				ID: "group",
			}
			_, err := subject.CreateRole(ctx, &protos.CreateRoleRequest{
				Name: name,
			})

			Expect(err).NotTo(HaveOccurred())

			req := &protos.UnassignRoleFromGroupRequest{
				Group:    group,
				RoleName: name,
			}
			res, err := subject.UnassignRoleFromGroup(ctx, req)

			Expect(err).To(HaveOccurred())
			Expect(res).To(BeNil())
		})

		It("fails if the role does not exist", func() {
			name := "fake-role"
			group := &protos.Group{
				ID: "group",
			}
			req := &protos.UnassignRoleFromGroupRequest{
				Group:    group,
				RoleName: name,
			}
			res, err := subject.UnassignRoleFromGroup(ctx, req)

			Expect(err).To(HaveOccurred())
			Expect(res).To(BeNil())
		})

		It("logs a security event", func() {
			group := &protos.Group{
				ID: "group-id",
			}
			req := &protos.UnassignRoleFromGroupRequest{
				Group:    group,
				RoleName: "role",
			}
			subject.UnassignRoleFromGroup(ctx, req)
			expectedExtensions := []cef.CustomExtension{
				{Key: "roleName", Value: "role"},
				{Key: "userID", Value: "group-id"},
			}

			Expect(securityLogger.LogCallCount()).To(Equal(1))
			_, signature, name, extensions := securityLogger.LogArgsForCall(0)
			Expect(signature).To(Equal("UnassignRoleFromGroup"))
			Expect(name).To(Equal("Role group unassignment"))
			Expect(extensions).To(Equal(expectedExtensions))
		})
	})

	Describe("#HasRole", func() {
		It("returns true if the actor has the role", func() {
			roleName := "role"
			actor := &protos.Actor{
				ID:        "actor",
				Namespace: "namespace",
			}
			_, err := subject.CreateRole(ctx, &protos.CreateRoleRequest{
				Name: roleName,
			})

			Expect(err).NotTo(HaveOccurred())

			_, err = subject.AssignRole(ctx, &protos.AssignRoleRequest{
				Actor:    actor,
				RoleName: roleName,
			})

			Expect(err).NotTo(HaveOccurred())

			res, err := subject.HasRole(ctx, &protos.HasRoleRequest{
				Actor:    actor,
				RoleName: roleName,
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(res).NotTo(BeNil())
			Expect(res.GetHasRole()).To(BeTrue())
		})

		It("returns false if only an actor with the same name but different namespace is assigned", func() {
			roleName := "role"
			actor1 := &protos.Actor{
				ID:        "actor",
				Namespace: "namespace1",
			}
			actor2 := &protos.Actor{
				ID:        "actor",
				Namespace: "namespace2",
			}
			_, err := subject.CreateRole(ctx, &protos.CreateRoleRequest{
				Name: roleName,
			})

			Expect(err).NotTo(HaveOccurred())

			_, err = subject.AssignRole(ctx, &protos.AssignRoleRequest{
				Actor:    actor1,
				RoleName: roleName,
			})

			Expect(err).NotTo(HaveOccurred())

			res, err := subject.HasRole(ctx, &protos.HasRoleRequest{
				Actor:    actor2,
				RoleName: roleName,
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(res).NotTo(BeNil())
			Expect(res.GetHasRole()).To(BeFalse())
		})

		It("returns false if the actor is not assigned", func() {
			roleName := "role"
			actor := &protos.Actor{
				ID:        "actor",
				Namespace: "namespace",
			}
			_, err := subject.CreateRole(ctx, &protos.CreateRoleRequest{
				Name: roleName,
			})

			Expect(err).NotTo(HaveOccurred())

			res, err := subject.HasRole(ctx, &protos.HasRoleRequest{
				Actor:    actor,
				RoleName: roleName,
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(res).NotTo(BeNil())
			Expect(res.GetHasRole()).To(BeFalse())
		})

		It("returns false if the role does not exist", func() {
			roleName := "role"
			actor := &protos.Actor{
				ID:        "actor",
				Namespace: "namespace",
			}
			res, err := subject.HasRole(ctx, &protos.HasRoleRequest{
				Actor:    actor,
				RoleName: roleName,
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(res).NotTo(BeNil())
			Expect(res.GetHasRole()).To(BeFalse())
		})

		It("fails when the actor namespace is not provided", func() {
			actor := &protos.Actor{
				ID:        "actor",
				Namespace: "",
			}
			res, err := subject.HasRole(ctx, &protos.HasRoleRequest{
				Actor:    actor,
				RoleName: "role",
			})

			expectedErr := status.Errorf(codes.InvalidArgument, "actor namespace cannot be empty")
			Expect(res).To(BeNil())
			Expect(err).To(MatchError(expectedErr))
		})
	})

	Describe("#AssignRoleToGroup", func() {
		It("succeeds if the role exists", func() {
			name := "role"
			group := &protos.Group{
				ID: "group-id",
			}
			_, err := subject.CreateRole(ctx, &protos.CreateRoleRequest{
				Name: name,
			})

			Expect(err).NotTo(HaveOccurred())

			req := &protos.AssignRoleToGroupRequest{
				Group:    group,
				RoleName: name,
			}
			res, err := subject.AssignRoleToGroup(ctx, req)

			Expect(err).NotTo(HaveOccurred())
			Expect(res).NotTo(BeNil())
		})

		It("fails if the user has already been assigned the role", func() {
			name := "role"
			group := &protos.Group{
				ID: "group-id",
			}
			_, err := subject.CreateRole(ctx, &protos.CreateRoleRequest{
				Name: name,
			})

			Expect(err).NotTo(HaveOccurred())

			req := &protos.AssignRoleToGroupRequest{
				Group:    group,
				RoleName: name,
			}
			_, err = subject.AssignRoleToGroup(ctx, req)

			Expect(err).NotTo(HaveOccurred())

			res, err := subject.AssignRoleToGroup(ctx, req)

			Expect(err).To(HaveOccurred())
			Expect(res).To(BeNil())
		})

		It("fails if the role does not exist", func() {
			group := &protos.Group{
				ID: "group",
			}
			res, err := subject.AssignRoleToGroup(ctx, &protos.AssignRoleToGroupRequest{
				Group:    group,
				RoleName: "does-not-exist",
			})

			Expect(res).To(BeNil())
			Expect(err).To(HaveOccurred())
		})

		It("logs a security event", func() {
			group := &protos.Group{
				ID: "group-id",
			}
			req := &protos.AssignRoleToGroupRequest{
				Group:    group,
				RoleName: "role",
			}
			subject.AssignRoleToGroup(ctx, req)
			expectedExtensions := []cef.CustomExtension{
				{Key: "roleName", Value: "role"},
				{Key: "groupID", Value: "group-id"},
			}

			Expect(securityLogger.LogCallCount()).To(Equal(1))
			_, signature, name, extensions := securityLogger.LogArgsForCall(0)
			Expect(signature).To(Equal("AssignRoleToGroup"))
			Expect(name).To(Equal("Role assignment"))
			Expect(extensions).To(Equal(expectedExtensions))
		})
	})

	Describe("#HasRoleForGroup", func() {
		It("returns true if the group has the role", func() {
			roleName := "role"
			group := &protos.Group{
				ID: "group",
			}
			_, err := subject.CreateRole(ctx, &protos.CreateRoleRequest{
				Name: roleName,
			})

			Expect(err).NotTo(HaveOccurred())

			_, err = subject.AssignRoleToGroup(ctx, &protos.AssignRoleToGroupRequest{
				Group:    group,
				RoleName: roleName,
			})

			Expect(err).NotTo(HaveOccurred())

			res, err := subject.HasRoleForGroup(ctx, &protos.HasRoleForGroupRequest{
				Group:    group,
				RoleName: roleName,
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(res).NotTo(BeNil())
			Expect(res.GetHasRole()).To(BeTrue())
		})

		It("returns false if the group is not assigned", func() {
			roleName := "role"
			group := &protos.Group{
				ID: "group",
			}
			_, err := subject.CreateRole(ctx, &protos.CreateRoleRequest{
				Name: roleName,
			})

			Expect(err).NotTo(HaveOccurred())

			res, err := subject.HasRoleForGroup(ctx, &protos.HasRoleForGroupRequest{
				Group:    group,
				RoleName: roleName,
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(res).NotTo(BeNil())
			Expect(res.GetHasRole()).To(BeFalse())
		})

		It("returns false if the role does not exist", func() {
			roleName := "role"
			group := &protos.Group{
				ID: "group",
			}
			res, err := subject.HasRoleForGroup(ctx, &protos.HasRoleForGroupRequest{
				Group:    group,
				RoleName: roleName,
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(res).NotTo(BeNil())
			Expect(res.GetHasRole()).To(BeFalse())
		})
	})

	Describe("#ListRolePermissions", func() {
		It("returns all the permissions that the role was created with", func() {
			roleName := "role1"
			permission1 := &protos.Permission{
				Action:          "action-1",
				ResourcePattern: "resource-pattern-1",
			}
			permission2 := &protos.Permission{
				Action:          "action-2",
				ResourcePattern: "resource-pattern-2",
			}

			_, err := subject.CreateRole(ctx, &protos.CreateRoleRequest{
				Name: roleName,
				Permissions: []*protos.Permission{
					permission1,
					permission2,
				},
			})

			Expect(err).NotTo(HaveOccurred())

			res, err := subject.ListRolePermissions(ctx, &protos.ListRolePermissionsRequest{
				RoleName: roleName,
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(res).NotTo(BeNil())

			var permissions []protos.Permission
			for _, p := range res.GetPermissions() {
				permissions = append(permissions, *p)
			}

			Expect(permissions).To(HaveLen(2))
			Expect(permissions).To(ContainElement(*permission1))
			Expect(permissions).To(ContainElement(*permission2))
		})

		It("returns an empty list if the actor has not been assigned to any roles", func() {
			roleName := "role1"
			_, err := subject.CreateRole(ctx, &protos.CreateRoleRequest{
				Name: roleName,
			})

			Expect(err).NotTo(HaveOccurred())

			res, err := subject.ListRolePermissions(ctx, &protos.ListRolePermissionsRequest{
				RoleName: roleName,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(res).NotTo(BeNil())
			Expect(res.GetPermissions()).To(HaveLen(0))
		})
	})
})
