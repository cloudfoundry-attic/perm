package repos

import "code.cloudfoundry.org/perm/models"

type PermissionQuery struct {
	PermissionName models.PermissionName
	ResourceID     ResourceID
}
