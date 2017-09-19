package rpc_test

import (
	"code.cloudfoundry.org/perm/rpc"

	"code.cloudfoundry.org/perm/protos"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	uuid "github.com/satori/go.uuid"
)

var _ = Describe("RoleServiceServer", func() {
	var (
		subject *rpc.RoleServiceServer
	)

	BeforeEach(func() {
		subject = rpc.NewRoleServiceServer()
	})

	Describe("#CreateRole", func() {
		It("succeeds if no role with that name exists", func() {
			req := &protos.CreateRoleRequest{
				Name: "test-role",
			}
			res, err := subject.CreateRole(nil, req)

			Expect(err).NotTo(HaveOccurred())
			Expect(res).NotTo(BeNil())
		})
	})

	Describe("#GetRole", func() {
		It("returns the role if a match exists", func() {
			name := "test"
			_, err := subject.CreateRole(nil, &protos.CreateRoleRequest{
				Name: name,
			})

			Expect(err).NotTo(HaveOccurred())

			req := &protos.GetRoleRequest{
				Name: name,
			}
			res, err := subject.GetRole(nil, req)

			Expect(err).NotTo(HaveOccurred())
			Expect(res).NotTo(BeNil())

			role := res.Role
			_, err = uuid.FromString(role.ID)

			Expect(role.Name).To(Equal(name))
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns an error if no match exists", func() {
			res, err := subject.GetRole(nil, &protos.GetRoleRequest{
				Name: "does-not-exist",
			})

			Expect(res).To(BeNil())
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("#AssignRole", func() {
		It("succeeds if the role exists", func() {
			name := "role"
			actor := &protos.Actor{
				ID:     "actor-id",
				Issuer: "fake-issuer",
			}
			_, err := subject.CreateRole(nil, &protos.CreateRoleRequest{
				Name: name,
			})

			Expect(err).NotTo(HaveOccurred())

			req := &protos.AssignRoleRequest{
				Actor:    actor,
				RoleName: name,
			}
			res, err := subject.AssignRole(nil, req)

			Expect(err).NotTo(HaveOccurred())
			Expect(res).NotTo(BeNil())
		})

		It("succeeds if the user has already been assigned the role", func() {
			name := "role"
			actor := &protos.Actor{
				ID:     "actor-id",
				Issuer: "fake-issuer",
			}
			_, err := subject.CreateRole(nil, &protos.CreateRoleRequest{
				Name: name,
			})

			Expect(err).NotTo(HaveOccurred())

			req := &protos.AssignRoleRequest{
				Actor:    actor,
				RoleName: name,
			}
			_, err = subject.AssignRole(nil, req)

			Expect(err).NotTo(HaveOccurred())

			res, err := subject.AssignRole(nil, req)

			Expect(err).NotTo(HaveOccurred())
			Expect(res).NotTo(BeNil())
		})
	})

	Describe("#HasRole", func() {
		It("returns true if the actor has the role", func() {
			roleName := "role"
			actor := &protos.Actor{
				ID:     "actor",
				Issuer: "issuer",
			}
			_, err := subject.CreateRole(nil, &protos.CreateRoleRequest{
				Name: roleName,
			})

			Expect(err).NotTo(HaveOccurred())

			_, err = subject.AssignRole(nil, &protos.AssignRoleRequest{
				Actor:    actor,
				RoleName: roleName,
			})

			Expect(err).NotTo(HaveOccurred())

			res, err := subject.HasRole(nil, &protos.HasRoleRequest{
				Actor:    actor,
				RoleName: roleName,
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(res).NotTo(BeNil())
			Expect(res.GetHasRole()).To(BeTrue())
		})

		It("returns false if only an actor with the same name but different issuer is assigned", func() {
			roleName := "role"
			actor1 := &protos.Actor{
				ID:     "actor",
				Issuer: "issuer1",
			}
			actor2 := &protos.Actor{
				ID:     "actor",
				Issuer: "issuer2",
			}
			_, err := subject.CreateRole(nil, &protos.CreateRoleRequest{
				Name: roleName,
			})

			Expect(err).NotTo(HaveOccurred())

			_, err = subject.AssignRole(nil, &protos.AssignRoleRequest{
				Actor:    actor1,
				RoleName: roleName,
			})

			Expect(err).NotTo(HaveOccurred())

			res, err := subject.HasRole(nil, &protos.HasRoleRequest{
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
				ID:     "actor",
				Issuer: "issuer",
			}
			_, err := subject.CreateRole(nil, &protos.CreateRoleRequest{
				Name: roleName,
			})

			Expect(err).NotTo(HaveOccurred())

			res, err := subject.HasRole(nil, &protos.HasRoleRequest{
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
				ID:     "actor",
				Issuer: "issuer",
			}
			res, err := subject.HasRole(nil, &protos.HasRoleRequest{
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
			actor := &protos.Actor{
				ID:     "actor",
				Issuer: "issuer",
			}

			_, err := subject.CreateRole(nil, &protos.CreateRoleRequest{
				Name: role1,
			})

			Expect(err).NotTo(HaveOccurred())

			_, err = subject.AssignRole(nil, &protos.AssignRoleRequest{
				Actor:    actor,
				RoleName: role1,
			})

			Expect(err).NotTo(HaveOccurred())

			_, err = subject.CreateRole(nil, &protos.CreateRoleRequest{
				Name: role2,
			})

			Expect(err).NotTo(HaveOccurred())

			_, err = subject.AssignRole(nil, &protos.AssignRoleRequest{
				Actor:    actor,
				RoleName: role2,
			})

			Expect(err).NotTo(HaveOccurred())

			res, err := subject.ListActorRoles(nil, &protos.ListActorRolesRequest{
				Actor: actor,
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(res).NotTo(BeNil())

			roles := []string{}
			for _, role := range res.GetRoles() {
				_, err := uuid.FromString(role.GetID())

				Expect(err).NotTo(HaveOccurred())

				roles = append(roles, role.GetName())
			}

			Expect(roles).To(HaveLen(2))
			Expect(roles).To(ContainElement(role1))
			Expect(roles).To(ContainElement(role2))
		})

		It("returns an empty list if the actor has not been assigned to any roles", func() {
			actor := &protos.Actor{
				ID:     "actor",
				Issuer: "issuer",
			}
			res, err := subject.ListActorRoles(nil, &protos.ListActorRolesRequest{
				Actor: actor,
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(res).NotTo(BeNil())
			Expect(res.GetRoles()).To(HaveLen(0))
		})
	})
})
