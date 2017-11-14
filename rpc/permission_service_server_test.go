package rpc_test

import (
	"context"

	"code.cloudfoundry.org/lager/lagertest"
	"code.cloudfoundry.org/perm/rpc"

	"code.cloudfoundry.org/perm/protos"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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

	Describe("HasPermission", func() {
		It("returns true if they have been assigned a role with a permission with a name matching the permission name and a resource pattern that matches the resourceID of the query", func() {
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
})
