package db

import (
	"code.cloudfoundry.org/perm/pkg/api/db/migrations"
	"code.cloudfoundry.org/perm/pkg/sqlx"
)

var MigrationsTableName = "perm_migrations"

var Migrations = []sqlx.Migration{
	{
		Name: "create_roles_table",
		Up:   migrations.CreateRolesTableUp,
		Down: migrations.CreateRolesTableDown,
	},
	{
		Name: "create_actors_table",
		Up:   migrations.CreateActorsTableUp,
		Down: migrations.CreateActorsTableDown,
	},
	{
		Name: "create_role_assignments_table",
		Up:   migrations.CreateRoleAssignmentsTableUp,
		Down: migrations.CreateRoleAssignmentsTableDown,
	},
	{
		Name: "create_permission_definitions_table",
		Up:   migrations.CreatePermissionDefinitionsTableUp,
		Down: migrations.CreatePermissionDefinitionsTableDown,
	},
	{
		Name: "create_permissions_table",
		Up:   migrations.CreatePermissionsTableUp,
		Down: migrations.CreatePermissionsTableDown,
	},
	{
		Name: "rename_permission_definition_to_action",
		Up:   migrations.RenamePermissionDefinitionToActionUp,
		Down: migrations.RenamePermissionDefinitionToActionDown,
	},
	{
		Name: "combine_actor_and_role_assignment_tables",
		Up:   migrations.CombineActorAndRoleAssignmentTablesUp,
		Down: migrations.CombineActorAndRoleAssignmentTablesDown,
	},
}
