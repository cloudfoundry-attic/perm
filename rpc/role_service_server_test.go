package rpc_test

import (
	"database/sql"

	"code.cloudfoundry.org/lager/lagertest"
	"code.cloudfoundry.org/perm/rpc"

	"code.cloudfoundry.org/perm/protos"
	"github.com/DATA-DOG/go-sqlmock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("RoleServiceServer", func() {
	var (
		subject *rpc.RoleServiceServer
		logger  *lagertest.TestLogger

		fakeDBConn *sql.DB
		dbMock     sqlmock.Sqlmock

		deps *rpc.InMemoryStore
	)

	BeforeEach(func() {
		var err error
		fakeDBConn, dbMock, err = sqlmock.New()

		Expect(err).NotTo(HaveOccurred())

		logger = lagertest.NewTestLogger("perm-test")
		deps = rpc.NewInMemoryStore()

		subject = rpc.NewRoleServiceServer(logger, fakeDBConn, deps)
	})

	AfterEach(func() {
		Expect(dbMock.ExpectationsWereMet()).To(Succeed())
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

		It("fails if a role with that name already exists", func() {
			req := &protos.CreateRoleRequest{
				Name: "test-role",
			}
			_, err := subject.CreateRole(nil, req)

			Expect(err).NotTo(HaveOccurred())

			res, err := subject.CreateRole(nil, req)

			Expect(res).To(BeNil())
			Expect(err).To(HaveOccurred())
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

			Expect(role.Name).To(Equal(name))
		})

		It("returns an error if no match exists", func() {
			res, err := subject.GetRole(nil, &protos.GetRoleRequest{
				Name: "does-not-exist",
			})

			Expect(res).To(BeNil())
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("#DeleteRole", func() {
		It("deletes the role if it exists", func() {
			name := "test-role"
			_, err := subject.CreateRole(nil, &protos.CreateRoleRequest{
				Name: name,
			})

			Expect(err).NotTo(HaveOccurred())

			res, err := subject.DeleteRole(nil, &protos.DeleteRoleRequest{
				Name: name,
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(res).NotTo(BeNil())

			_, err = subject.GetRole(nil, &protos.GetRoleRequest{
				Name: name,
			})

			Expect(err).To(HaveOccurred())
		})

		It("fails if the role does not exist", func() {
			res, err := subject.DeleteRole(nil, &protos.DeleteRoleRequest{
				Name: "test-role",
			})

			Expect(res).To(BeNil())
			Expect(err).To(HaveOccurred())
		})

		It("deletes any role assignments for the role", func() {
			name := "test-role"

			_, err := subject.CreateRole(nil, &protos.CreateRoleRequest{
				Name: name,
			})

			Expect(err).NotTo(HaveOccurred())

			actor := &protos.Actor{
				ID:     "actor-id",
				Issuer: "issuer",
			}

			_, err = subject.AssignRole(nil, &protos.AssignRoleRequest{
				Actor:    actor,
				RoleName: name,
			})

			Expect(err).NotTo(HaveOccurred())

			hasRoleRes, err := subject.HasRole(nil, &protos.HasRoleRequest{
				Actor:    actor,
				RoleName: name,
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(hasRoleRes).NotTo(BeNil())
			Expect(hasRoleRes.GetHasRole()).To(BeTrue())

			res, err := subject.DeleteRole(nil, &protos.DeleteRoleRequest{
				Name: name,
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(res).NotTo(BeNil())

			hasRoleRes, err = subject.HasRole(nil, &protos.HasRoleRequest{
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

		It("fails if the user has already been assigned the role", func() {
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

			Expect(err).To(HaveOccurred())
			Expect(res).To(BeNil())
		})

		It("fails if the role does not exist", func() {
			actor := &protos.Actor{
				ID:     "actor",
				Issuer: "issuer",
			}
			res, err := subject.AssignRole(nil, &protos.AssignRoleRequest{
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
			actor := &protos.Actor{
				ID:     "actor-id",
				Issuer: "fake-issuer",
			}
			_, err := subject.CreateRole(nil, &protos.CreateRoleRequest{
				Name: name,
			})

			Expect(err).NotTo(HaveOccurred())

			_, err = subject.AssignRole(nil, &protos.AssignRoleRequest{
				Actor:    actor,
				RoleName: name,
			})

			Expect(err).NotTo(HaveOccurred())

			req := &protos.UnassignRoleRequest{
				Actor:    actor,
				RoleName: name,
			}
			res, err := subject.UnassignRole(nil, req)

			Expect(err).NotTo(HaveOccurred())
			Expect(res).NotTo(BeNil())
		})

		It("fails if the user is not assigned to the role", func() {
			name := "role"
			actor := &protos.Actor{
				ID:     "actor",
				Issuer: "issuer",
			}
			_, err := subject.CreateRole(nil, &protos.CreateRoleRequest{
				Name: name,
			})

			Expect(err).NotTo(HaveOccurred())

			req := &protos.UnassignRoleRequest{
				Actor:    actor,
				RoleName: name,
			}
			res, err := subject.UnassignRole(nil, req)

			Expect(err).To(HaveOccurred())
			Expect(res).To(BeNil())
		})

		It("fails if the role does not exist", func() {
			name := "fake-role"
			actor := &protos.Actor{
				ID:     "actor",
				Issuer: "issuer",
			}
			req := &protos.UnassignRoleRequest{
				Actor:    actor,
				RoleName: name,
			}
			res, err := subject.UnassignRole(nil, req)

			Expect(err).To(HaveOccurred())
			Expect(res).To(BeNil())
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
