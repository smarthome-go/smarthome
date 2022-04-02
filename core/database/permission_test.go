package database

import (
	"testing"
)

func addAllPermissions() error {
	for _, permission := range Permissions {
		if _, err := AddUserPermission("admin", permission.Permission); err != nil {
			return err
		}
	}
	return nil
}

func TestCreatePermissionTable(t *testing.T) {
	if err := createPermissionTable(); err != nil {
		t.Error(err.Error())
		return
	}
}

func TestCreateHasPermissionTable(t *testing.T) {
	if err := createHasPermissionTable(); err != nil {
		t.Error(err.Error())
		return
	}
}

func TestInitializePermissions(t *testing.T) {
	if err := initializePermissions(); err != nil {
		t.Error(err.Error())
		return
	}
}

func TestAddUserPermission(t *testing.T) {
	if err := addAllPermissions(); err != nil {
		t.Error(err.Error())
	}
	for _, permission := range Permissions {
		alreadyHasPermission, err := AddUserPermission("admin", permission.Permission)
		if err != nil {
			t.Error(err.Error())
			return
		}
		if alreadyHasPermission {
			t.Errorf("Add User permission failed: user does not already have permission %s", permission.Permission)
			return
		}
	}
}

func TestRemoveAllPermissionOfUser(t *testing.T) {
	if err := addAllPermissions(); err != nil {
		t.Error(err.Error())
		return
	}
	if err := RemoveAllPermissionsOfUser("admin"); err != nil {
		t.Error(err.Error())
		return
	}
	permissions, err := GetUserPermissions("admin")
	if err != nil {
		t.Error(err.Error())
		return
	}
	if len(permissions) > 0 {
		t.Errorf("Failed to delete all permissions of user: amount of permissions is greater than 0. want: 0 got: %d", len(permissions))
		return
	}
}

func TestRemovePermissionOfUser(t *testing.T) {
	if err := addAllPermissions(); err != nil {
		t.Error(err.Error()) // The 10.000 line of code :)
		return
	}
	for _, permission := range Permissions {
		if _, err := RemoveUserPermission("admin", permission.Permission); err != nil {
			t.Error(err.Error())
			return
		}
	}
}
