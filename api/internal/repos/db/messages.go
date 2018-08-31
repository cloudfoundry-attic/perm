package db

const (
	failedToStartTransaction = "failed-to-start-transaction"

	failedToRetrieveID        = "failed-to-retrieve-id"
	failedToCountRowsAffected = "failed-to-count-rows-affected"
	failedToScanRow           = "failed-to-scan-row"
	failedToIterateOverRows   = "failed-to-iterate-over-rows"

	errRoleAlreadyExists = "role-already-exists"
	errRoleNotFound      = "role-not-found"

	failedToCreateRole = "failed-to-create-role"
	failedToFindRole   = "failed-to-find-role"
	failedToDeleteRole = "failed-to-delete-role"

	errActorAlreadyExists = "actor-already-exists"
	errActorNotFound      = "actor-not-found"

	failedToCreateActor = "failed-to-create-actor"
	failedToFindActor   = "failed-to-find-actor"

	errRoleAssignmentAlreadyExists = "role-assignment-already-exists"
	errRoleAssignmentNotFound      = "role-assignment-not-found"

	failedToCreateRoleAssignment = "failed-to-create-role-assignment"
	failedToDeleteRoleAssignment = "failed-to-delete-role-assignment"
	failedToFindRoleAssignment   = "failed-to-find-role-assignment"
	failedToFindRoleAssignments  = "failed-to-find-role-assignments"

	errActionAlreadyExists = "action-already-exists"
	errActionNotFound      = "action-not-found"

	failedToCreateAction = "failed-to-create-action"
	failedToFindAction   = "failed-to-find-action"

	errPermissionAlreadyExists = "permission-already-exists"

	failedToCreatePermission = "failed-to-create-permission"
	failedToFindPermissions  = "failed-to-find-permissions"

	failedToListResourcePatterns = "failed-to-list-resource-patterns"
)
