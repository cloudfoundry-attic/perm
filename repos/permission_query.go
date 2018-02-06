package repos

import "code.cloudfoundry.org/perm/models"

type PermissionQuery struct {
	PermissionName  models.PermissionName
	ResourcePattern models.PermissionResourcePattern
}
