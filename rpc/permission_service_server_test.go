package rpc_test

import (
	"context"

	"code.cloudfoundry.org/lager/lagertest"
	"code.cloudfoundry.org/perm/repos/reposfakes"
	"code.cloudfoundry.org/perm/rpc"

	"errors"

	"code.cloudfoundry.org/perm-go"
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

		inMemoryStore *rpc.InMemoryStore

		ctx context.Context
	)

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("perm-test")
		inMemoryStore = rpc.NewInMemoryStore()

		ctx = context.Background()

		roleServiceServer = rpc.NewRoleServiceServer(logger, inMemoryStore, inMemoryStore)
		subject = rpc.NewPermissionServiceServer(logger, inMemoryStore)
	})

	Describe("#HasPermission", func() {
		It("returns true if they have been assigned a role with a permission with a name "+
			"matching the permission name and a resource pattern that matches the resourceID of the query", func() {
			roleName := "role"
			actor := &protos.Actor{
				ID:     "actor",
				Issuer: "issuer",
			}

			permission1 := &protos.Permission{
				Name:            "some-permission",
				ResourcePattern: "some-resource-ID",
			}
			permission2 := &protos.Permission{
				Name:            "some-other-permission",
				ResourcePattern: "some-other-resource-ID",
			}

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

		It("returns false if they mismatch the permission name and resourceID", func() {
			roleName := "role"
			actor := &protos.Actor{
				ID:     "actor",
				Issuer: "issuer",
			}

			permission1 := &protos.Permission{
				Name:            "some-permission",
				ResourcePattern: "some-resource-ID",
			}
			permission2 := &protos.Permission{
				Name:            "some-other-permission",
				ResourcePattern: "some-other-resource-ID",
			}

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
			roleName := "role"
			actor := &protos.Actor{
				ID:     "actor",
				Issuer: "issuer",
			}

			permission := &protos.Permission{
				Name:            "some-permission",
				ResourcePattern: "some-resource-ID",
			}

			_, err := roleServiceServer.CreateRole(ctx, &protos.CreateRoleRequest{
				Name: roleName,
				Permissions: []*protos.Permission{
					permission,
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
			roleName := "role"
			actor := &protos.Actor{
				ID:     "actor",
				Issuer: "issuer",
			}

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
				Issuer: "test-issuer",
				ID:     "fancy-id",
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
					ID:     "123",
					Issuer: "issuer34",
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
			subject := rpc.NewPermissionServiceServer(logger, permissionRepo)

			testErr := errors.New("test-error")

			permissionRepo.ListResourcePatternsReturns(nil, testErr)

			req := &protos.ListResourcePatternsRequest{
				Actor: &protos.Actor{
					ID:     "123",
					Issuer: "issuer34",
				},
				PermissionName: "p12",
			}

			_, err := subject.ListResourcePatterns(ctx, req)

			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(status.Errorf(codes.Unknown, "test-error")))
		})
	})
})
