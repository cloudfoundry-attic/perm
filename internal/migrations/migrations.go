package migrations

import (
	"code.cloudfoundry.org/perm/pkg/sqlx"
)

var TableName = "perm_migrations"

var Migrations = []sqlx.Migration{
	{
		Name: "create_roles_table",
		Up:   createRolesTableUp,
		Down: createRolesTableDown,
	},
	{
		Name: "create_actors_table",
		Up:   createActorsTableUp,
		Down: createActorsTableDown,
	},
	{
		Name: "create_role_assignments_table",
		Up:   createRoleAssignmentsTableUp,
		Down: createRoleAssignmentsTableDown,
	},
	{
		Name: "create_permission_definitions_table",
		Up:   createPermissionDefinitionsTableUp,
		Down: createPermissionDefinitionsTableDown,
	},
	{
		Name: "create_permissions_table",
		Up:   createPermissionsTableUp,
		Down: createPermissionsTableDown,
	},
	{
		Name: "rename_permission_definition_to_action",
		Up:   renamePermissionDefinitionToActionUp,
		Down: renamePermissionDefinitionToActionDown,
	},
	{
		Name: "combine_actor_and_role_assignment_tables",
		Up:   combineActorAndRoleAssignmentTablesUp,
		Down: combineActorAndRoleAssignmentTablesDown,
	},
	{
		Name: "create_group_assignment_table",
		Up:   createGroupAssignmentTableUp,
		Down: createGroupAssignmentTableDown,
	},
}
