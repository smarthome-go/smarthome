package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateCameraPermissionsTable(t *testing.T) {
	if err := createCameraTable(); err != nil {
		t.Error(err.Error())
	}
}

func TestUserCameraPermissions(t *testing.T) {
	// Flush database first
	if err := initDB(true); err != nil {
		t.Error(err.Error())
	}
	// Create a test user
	if err := AddUser(FullUser{Username: "cam_perm"}); err != nil {
		t.Error(err.Error())
	}
	// Create test room
	if err := CreateRoom(RoomData{Id: "test"}); err != nil {
		t.Error(err.Error())
	}
	// Create test cameras
	cams := []Camera{
		{
			Id:     "perm1",
			RoomId: "test",
		},
		{
			Id:     "perm2",
			RoomId: "test",
		},
		{
			Id:     "perm3",
			RoomId: "test",
		},
	}
	for _, cam := range cams {
		if err := CreateCamera(cam); err != nil {
			t.Error(err.Error())
		}
	}
	// Check if permissions are initially empty
	emptyPerm, err := GetUserCameraPermissions("cam_perm")
	if err != nil {
		t.Error(err.Error())
	}
	assert.Empty(t, emptyPerm)
	// Grant user permission to every camera in cams
	for _, cam := range cams {
		if _, err := AddUserCameraPermission("cam_perm", cam.Id); err != nil {
			t.Error(err.Error())
		}
	}
	// Check permissions again
	wantedPerm := make([]string, 0)
	for _, cam := range cams {
		wantedPerm = append(wantedPerm, cam.Id)
	}
	allPerms, err := GetUserCameraPermissions("cam_perm")
	if err != nil {
		t.Error(err.Error())
	}
	assert.Equal(t, wantedPerm, allPerms)
	// Check if the camera is also granted using the `UserHasCameraPermission` function
	for _, cam := range allPerms {
		hasPermission, err := UserHasCameraPermission("cam_perm", cam)
		if err != nil {
			t.Error(err.Error())
		}
		assert.True(t, hasPermission)
	}
	// Remove every permission from the user
	if err := RemoveAllCameraPermissionsOfUser("cam_perm"); err != nil {
		t.Error(err.Error())
	}
	permsAfterDel, err := GetUserCameraPermissions("cam_perm")
	if err != nil {
		t.Error(err.Error())
	}
	assert.Empty(t, permsAfterDel)
	// Remove every permission from the user using a different function
	// Grant user permission to every camera in cams again
	for _, cam := range cams {
		if _, err := AddUserCameraPermission("cam_perm", cam.Id); err != nil {
			t.Error(err.Error())
		}
		if _, err := RemoveUserCameraPermission("cam_perm", cam.Id); err != nil {
			t.Error(err.Error())
		}
		// Test if the `UserHasCameraPermission` function is able to output `false`
		hasPermission, err := UserHasCameraPermission("cam_perm", cam.Id)
		if err != nil {
			t.Error(err.Error())
		}
		assert.False(t, hasPermission)
	}
	permsAfterDel2, err := GetUserCameraPermissions("cam_perm")
	if err != nil {
		t.Error(err.Error())
	}
	assert.Empty(t, permsAfterDel2)
}
