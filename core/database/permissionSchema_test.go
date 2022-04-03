package database

import (
	"testing"
)

func TestPermissionsList(t *testing.T) {
	for _, permission := range Permissions {
		if len(permission.Name) < 5 {
			t.Errorf("Name ('%s') of permission %s is not long enough, short names are too hard to understand", permission.Permission, permission.Name)
			return
		}
		if len(permission.Description) < 10 {
			t.Errorf("Description ('%s') of permission %s is not long enough, short descriptions are often hard to understand", permission.Description, permission.Permission)
			return
		}
	}
}
