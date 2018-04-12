package rpc_test

import (
	"context"

	"code.cloudfoundry.org/lager/lagertest"
	"code.cloudfoundry.org/perm/pkg/api/repos/reposfakes"
	"code.cloudfoundry.org/perm/pkg/api/rpc"

	"errors"

	"code.cloudfoundry.org/perm-go"
	"code.cloudfoundry.org/perm/pkg/api/logging"
	"code.cloudfoundry.org/perm/pkg/api/rpc/rpcfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ = Describe("PermissionServiceServer", func() {
	var (
		subject           *rpc.PermissionServiceServer
		roleServiceServer *rpc.RoleServiceServer
		logger            *lagertest.TestLogger
		securityLogger    *rpcfakes.FakeSecurityLogger

		inMemoryStore *rpc.InMemoryStore

		ctx context.Context
	)

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("perm-test")
		securityLogger = new(rpcfakes.FakeSecurityLogger)
		inMemoryStore = rpc.NewInMemoryStore()

		ctx = context.Background()
		roleServiceServer = rpc.NewRoleServiceServer(logger, securityLogger, inMemoryStore, inMemoryStore)
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
				Name:            "some-permission",
				ResourcePattern: "some-resource-ID",
			}
			permission2 = &protos.Permission{
				Name:            "some-other-permission",
				ResourcePattern: "some-other-resource-ID",
			}
		})

		It("returns true when there is a matching permission name and resourceID", func() {
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
				Actor:          actor,
				PermissionName: "some-other-permission",
				ResourceId:     "some-other-resource-ID",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(res.GetHasPermission()).To(BeTrue())
		})

		It("returns false when mismatch the permission name and resourceID", func() {
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
				Actor:          actor,
				PermissionName: "some-permission",
				ResourceId:     "some-other-resource-ID",
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
				Actor:          actor,
				PermissionName: "some-permission",
				ResourceId:     "some-resource-ID",
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
				Actor:          actor,
				PermissionName: "some-permission",
				ResourceId:     "some-resource-ID",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(res.GetHasPermission()).To(BeFalse())
		})

		It("fails when the actor namespace is not provided", func() {
			actor := &protos.Actor{
				ID:        "actor",
				Namespace: "",
			}

			res, err := subject.HasPermission(ctx, &protos.HasPermissionRequest{
				Actor:          actor,
				PermissionName: "some-permission",
				ResourceId:     "some-resource-ID",
			})
			expectedErr := status.Errorf(codes.InvalidArgument, "actor namespace cannot be empty")
			Expect(res).To(BeNil())
			Expect(err).To(MatchError(expectedErr))
		})

		It("logs a security event", func() {
			_, err := subject.HasPermission(ctx, &protos.HasPermissionRequest{
				Actor:          actor,
				PermissionName: "some-permission",
				ResourceId:     "some-resource-ID",
			})
			Expect(err).NotTo(HaveOccurred())
			expectedExtensions := []logging.CustomExtension{
				{Key: "userID", Value: "actor"},
				{Key: "permission", Value: "some-permission"},
				{Key: "resourceID", Value: "some-resource-ID"},
			}

			Expect(securityLogger.LogCallCount()).To(Equal(1))
			_, signature, name, extensions := securityLogger.LogArgsForCall(0)
			Expect(signature).To(Equal("HasPermission"))
			Expect(name).To(Equal("Permission check"))
			Expect(extensions).To(Equal(expectedExtensions))
		})
	})

	Describe("#ListResourcePatterns", func() {
		It("returns the list of resource patterns", func() {
			p1 := &protos.Permission{Name: "test-permission-name", ResourcePattern: "foo"}
			p2 := &protos.Permission{Name: "another-permission-name", ResourcePattern: "bar"}
			p3 := &protos.Permission{Name: "test-permission-name", ResourcePattern: "baz"}

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
				Actor:          actor,
				PermissionName: "test-permission-name",
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
				PermissionName: "p12",
			}

			response, err := subject.ListResourcePatterns(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			Expect(response).NotTo(BeNil())
			Expect(response.ResourcePatterns).To(BeEmpty())
		})

		It("returns a relevant error if the query fails", func() {
			permissionRepo := new(reposfakes.FakePermissionRepo)
			subject := rpc.NewPermissionServiceServer(logger, securityLogger, permissionRepo)
			testErr := errors.New("test-error")
			permissionRepo.ListResourcePatternsReturns(nil, testErr)

			req := &protos.ListResourcePatternsRequest{
				Actor: &protos.Actor{
					ID:        "123",
					Namespace: "namespace34",
				},
				PermissionName: "p12",
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
	})

})
