package rpc_test

import (
	"context"

	"code.cloudfoundry.org/lager/lagertest"
	"code.cloudfoundry.org/perm/rpc"

	"code.cloudfoundry.org/perm-go"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("RoleServiceServer", func() {
	var (
		subject *rpc.RoleServiceServer
		logger  *lagertest.TestLogger

		inMemoryStore *rpc.InMemoryStore

		ctx context.Context
	)

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("perm-test")
		inMemoryStore = rpc.NewInMemoryStore()

		ctx = context.Background()

		subject = rpc.NewRoleServiceServer(logger, inMemoryStore, inMemoryStore)
	})

	Describe("#CreateRole", func() {
		It("succeeds if no role with that name exists", func() {
			req := &perm_go.CreateRoleRequest{
				Name: "test-role",
			}
			res, err := subject.CreateRole(ctx, req)

			Expect(err).NotTo(HaveOccurred())
			Expect(res).NotTo(BeNil())
		})

		It("fails if a role with that name already exists", func() {
			req := &perm_go.CreateRoleRequest{
				Name: "test-role",
			}
			_, err := subject.CreateRole(ctx, req)

			Expect(err).NotTo(HaveOccurred())

			res, err := subject.CreateRole(ctx, req)

			Expect(res).To(BeNil())
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("#GetRole", func() {
		It("returns the role if a match exists", func() {
			name := "test"
			_, err := subject.CreateRole(ctx, &perm_go.CreateRoleRequest{
				Name: name,
			})

			Expect(err).NotTo(HaveOccurred())

			req := &perm_go.GetRoleRequest{
				Name: name,
			}
			res, err := subject.GetRole(ctx, req)

			Expect(err).NotTo(HaveOccurred())
			Expect(res).NotTo(BeNil())

			role := res.Role

			Expect(role.Name).To(Equal(name))
		})

		It("returns an error if no match exists", func() {
			res, err := subject.GetRole(ctx, &perm_go.GetRoleRequest{
				Name: "does-not-exist",
			})

			Expect(res).To(BeNil())
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("#DeleteRole", func() {
		It("deletes the role if it exists", func() {
			name := "test-role"
			_, err := subject.CreateRole(ctx, &perm_go.CreateRoleRequest{
				Name: name,
			})

			Expect(err).NotTo(HaveOccurred())

			res, err := subject.DeleteRole(ctx, &perm_go.DeleteRoleRequest{
				Name: name,
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(res).NotTo(BeNil())

			_, err = subject.GetRole(ctx, &perm_go.GetRoleRequest{
				Name: name,
			})

			Expect(err).To(HaveOccurred())
		})

		It("fails if the role does not exist", func() {
			res, err := subject.DeleteRole(ctx, &perm_go.DeleteRoleRequest{
				Name: "test-role",
			})

			Expect(res).To(BeNil())
			Expect(err).To(HaveOccurred())
		})

		It("deletes any role assignments for the role", func() {
			name := "test-role"

			_, err := subject.CreateRole(ctx, &perm_go.CreateRoleRequest{
				Name: name,
			})

			Expect(err).NotTo(HaveOccurred())

			actor := &perm_go.Actor{
				ID:     "actor-id",
				Issuer: "issuer",
			}

			_, err = subject.AssignRole(ctx, &perm_go.AssignRoleRequest{
				Actor:    actor,
				RoleName: name,
			})

			Expect(err).NotTo(HaveOccurred())

			hasRoleRes, err := subject.HasRole(ctx, &perm_go.HasRoleRequest{
				Actor:    actor,
				RoleName: name,
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(hasRoleRes).NotTo(BeNil())
			Expect(hasRoleRes.GetHasRole()).To(BeTrue())

			res, err := subject.DeleteRole(ctx, &perm_go.DeleteRoleRequest{
				Name: name,
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(res).NotTo(BeNil())

			hasRoleRes, err = subject.HasRole(ctx, &perm_go.HasRoleRequest{
				Actor:    actor,
				RoleName: name,
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(hasRoleRes).NotTo(BeNil())
			Expect(hasRoleRes.GetHasRole()).To(BeFalse())
		})
	})

	Describe("#AssignRole", func() {
		It("succeeds if the role exists", func() {
			name := "role"
			actor := &perm_go.Actor{
				ID:     "actor-id",
				Issuer: "fake-issuer",
			}
			_, err := subject.CreateRole(ctx, &perm_go.CreateRoleRequest{
				Name: name,
			})

			Expect(err).NotTo(HaveOccurred())

			req := &perm_go.AssignRoleRequest{
				Actor:    actor,
				RoleName: name,
			}
			res, err := subject.AssignRole(ctx, req)

			Expect(err).NotTo(HaveOccurred())
			Expect(res).NotTo(BeNil())
		})

		It("fails if the user has already been assigned the role", func() {
			name := "role"
			actor := &perm_go.Actor{
				ID:     "actor-id",
				Issuer: "fake-issuer",
			}
			_, err := subject.CreateRole(ctx, &perm_go.CreateRoleRequest{
				Name: name,
			})

			Expect(err).NotTo(HaveOccurred())

			req := &perm_go.AssignRoleRequest{
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
			actor := &perm_go.Actor{
				ID:     "actor",
				Issuer: "issuer",
			}
			res, err := subject.AssignRole(ctx, &perm_go.AssignRoleRequest{
				Actor:    actor,
				RoleName: "does-not-exist",
			})

			Expect(res).To(BeNil())
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("#UnassignRole", func() {
		It("removes role binding if the user has that role", func() {
			name := "role"
			actor := &perm_go.Actor{
				ID:     "actor-id",
				Issuer: "fake-issuer",
			}
			_, err := subject.CreateRole(ctx, &perm_go.CreateRoleRequest{
				Name: name,
			})

			Expect(err).NotTo(HaveOccurred())

			_, err = subject.AssignRole(ctx, &perm_go.AssignRoleRequest{
				Actor:    actor,
				RoleName: name,
			})

			Expect(err).NotTo(HaveOccurred())

			req := &perm_go.UnassignRoleRequest{
				Actor:    actor,
				RoleName: name,
			}
			res, err := subject.UnassignRole(ctx, req)

			Expect(err).NotTo(HaveOccurred())
			Expect(res).NotTo(BeNil())
		})

		It("fails if the user is not assigned to the role", func() {
			name := "role"
			actor := &perm_go.Actor{
				ID:     "actor",
				Issuer: "issuer",
			}
			_, err := subject.CreateRole(ctx, &perm_go.CreateRoleRequest{
				Name: name,
			})

			Expect(err).NotTo(HaveOccurred())

			req := &perm_go.UnassignRoleRequest{
				Actor:    actor,
				RoleName: name,
			}
			res, err := subject.UnassignRole(ctx, req)

			Expect(err).To(HaveOccurred())
			Expect(res).To(BeNil())
		})

		It("fails if the role does not exist", func() {
			name := "fake-role"
			actor := &perm_go.Actor{
				ID:     "actor",
				Issuer: "issuer",
			}
			req := &perm_go.UnassignRoleRequest{
				Actor:    actor,
				RoleName: name,
			}
			res, err := subject.UnassignRole(ctx, req)

			Expect(err).To(HaveOccurred())
			Expect(res).To(BeNil())
		})
	})

	Describe("#HasRole", func() {
		It("returns true if the actor has the role", func() {
			roleName := "role"
			actor := &perm_go.Actor{
				ID:     "actor",
				Issuer: "issuer",
			}
			_, err := subject.CreateRole(ctx, &perm_go.CreateRoleRequest{
				Name: roleName,
			})

			Expect(err).NotTo(HaveOccurred())

			_, err = subject.AssignRole(ctx, &perm_go.AssignRoleRequest{
				Actor:    actor,
				RoleName: roleName,
			})

			Expect(err).NotTo(HaveOccurred())

			res, err := subject.HasRole(ctx, &perm_go.HasRoleRequest{
				Actor:    actor,
				RoleName: roleName,
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(res).NotTo(BeNil())
			Expect(res.GetHasRole()).To(BeTrue())
		})

		It("returns false if only an actor with the same name but different issuer is assigned", func() {
			roleName := "role"
			actor1 := &perm_go.Actor{
				ID:     "actor",
				Issuer: "issuer1",
			}
			actor2 := &perm_go.Actor{
				ID:     "actor",
				Issuer: "issuer2",
			}
			_, err := subject.CreateRole(ctx, &perm_go.CreateRoleRequest{
				Name: roleName,
			})

			Expect(err).NotTo(HaveOccurred())

			_, err = subject.AssignRole(ctx, &perm_go.AssignRoleRequest{
				Actor:    actor1,
				RoleName: roleName,
			})

			Expect(err).NotTo(HaveOccurred())

			res, err := subject.HasRole(ctx, &perm_go.HasRoleRequest{
				Actor:    actor2,
				RoleName: roleName,
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(res).NotTo(BeNil())
			Expect(res.GetHasRole()).To(BeFalse())
		})

		It("returns false if the actor is not assigned", func() {
			roleName := "role"
			actor := &perm_go.Actor{
				ID:     "actor",
				Issuer: "issuer",
			}
			_, err := subject.CreateRole(ctx, &perm_go.CreateRoleRequest{
				Name: roleName,
			})

			Expect(err).NotTo(HaveOccurred())

			res, err := subject.HasRole(ctx, &perm_go.HasRoleRequest{
				Actor:    actor,
				RoleName: roleName,
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(res).NotTo(BeNil())
			Expect(res.GetHasRole()).To(BeFalse())
		})

		It("returns false if the role does not exist", func() {
			roleName := "role"
			actor := &perm_go.Actor{
				ID:     "actor",
				Issuer: "issuer",
			}
			res, err := subject.HasRole(ctx, &perm_go.HasRoleRequest{
				Actor:    actor,
				RoleName: roleName,
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(res).NotTo(BeNil())
			Expect(res.GetHasRole()).To(BeFalse())
		})
	})

	Describe("#ListActorRoles", func() {
		It("returns all the roles that the actor has been assigned to", func() {
			role1 := "role1"
			role2 := "role2"
			actor := &perm_go.Actor{
				ID:     "actor",
				Issuer: "issuer",
			}

			_, err := subject.CreateRole(ctx, &perm_go.CreateRoleRequest{
				Name: role1,
			})

			Expect(err).NotTo(HaveOccurred())

			_, err = subject.AssignRole(ctx, &perm_go.AssignRoleRequest{
				Actor:    actor,
				RoleName: role1,
			})

			Expect(err).NotTo(HaveOccurred())

			_, err = subject.CreateRole(ctx, &perm_go.CreateRoleRequest{
				Name: role2,
			})

			Expect(err).NotTo(HaveOccurred())

			_, err = subject.AssignRole(ctx, &perm_go.AssignRoleRequest{
				Actor:    actor,
				RoleName: role2,
			})

			Expect(err).NotTo(HaveOccurred())

			res, err := subject.ListActorRoles(ctx, &perm_go.ListActorRolesRequest{
				Actor: actor,
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(res).NotTo(BeNil())

			roles := []string{}
			for _, role := range res.GetRoles() {
				roles = append(roles, role.GetName())
			}

			Expect(roles).To(HaveLen(2))
			Expect(roles).To(ContainElement(role1))
			Expect(roles).To(ContainElement(role2))
		})

		It("returns an empty list if the actor has not been assigned to any roles", func() {
			actor := &perm_go.Actor{
				ID:     "actor",
				Issuer: "issuer",
			}
			res, err := subject.ListActorRoles(ctx, &perm_go.ListActorRolesRequest{
				Actor: actor,
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(res).NotTo(BeNil())
			Expect(res.GetRoles()).To(HaveLen(0))
		})
	})

	Describe("#ListRolePermissions", func() {
		It("returns all the permissions that the role was created with", func() {
			roleName := "role1"
			permission1 := &perm_go.Permission{
				Name:            "permission-1",
				ResourcePattern: "resource-pattern-1",
			}
			permission2 := &perm_go.Permission{
				Name:            "permission-2",
				ResourcePattern: "resource-pattern-2",
			}

			_, err := subject.CreateRole(ctx, &perm_go.CreateRoleRequest{
				Name: roleName,
				Permissions: []*perm_go.Permission{
					permission1,
					permission2,
				},
			})

			Expect(err).NotTo(HaveOccurred())

			res, err := subject.ListRolePermissions(ctx, &perm_go.ListRolePermissionsRequest{
				RoleName: roleName,
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(res).NotTo(BeNil())

			var permissions []perm_go.Permission
			for _, p := range res.GetPermissions() {
				permissions = append(permissions, *p)
			}

			Expect(permissions).To(HaveLen(2))
			Expect(permissions).To(ContainElement(*permission1))
			Expect(permissions).To(ContainElement(*permission2))
		})

		It("returns an empty list if the actor has not been assigned to any roles", func() {
			roleName := "role1"
			_, err := subject.CreateRole(ctx, &perm_go.CreateRoleRequest{
				Name: roleName,
			})

			Expect(err).NotTo(HaveOccurred())

			res, err := subject.ListRolePermissions(ctx, &perm_go.ListRolePermissionsRequest{
				RoleName: roleName,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(res).NotTo(BeNil())
			Expect(res.GetPermissions()).To(HaveLen(0))
		})
	})
})
