package database

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateCameraTable(t *testing.T) {
	if err := createCameraTable(); err != nil {
		t.Error(err.Error())
	}
}

func TestCreateCamera(t *testing.T) {
	// Delete database first
	if err := initDB(true); err != nil {
		t.Error(err.Error())
	}
	// Create test rooms
	if err := CreateRoom(RoomData{Id: "test"}); err != nil {
		t.Error(err.Error())
		return
	}
	if err := CreateRoom(RoomData{Id: "test2"}); err != nil {
		t.Error(err.Error())
	}
	table := []Camera{
		{
			Id:     "test_1",
			Name:   "test 1",
			Url:    "http://example/com/1",
			RoomId: "test",
		},
		{
			Id:     "test_2",
			Name:   "test 2",
			Url:    "http://example.com/2",
			RoomId: "test2",
		},
	}

	for _, camera := range table {
		t.Run(fmt.Sprintf("TestCreateCamera/%s", camera.Id), func(t *testing.T) {
			// Create
			if err := CreateCamera(camera); err != nil {
				t.Error(err.Error())
			}
			// Assert equality
			cam, found, err := GetCameraById(camera.Id)
			if err != nil {
				t.Error(err.Error())
			}
			if !found {
				t.Errorf("Camera %s not found after creation", camera.Id)
			}
			assert.Equal(t, camera, cam, "camera from id has invalid metadata")
		})
	}

	t.Run("TestCreateCamera/compare_lists", func(t *testing.T) {
		// Check the listing function
		cameras, err := ListCameras()
		if err != nil {
			t.Error(err.Error())
		}
		assert.Equal(t, table, cameras, "listed cameras do not match table")
	})
}

func TestListCamerasRedacted(t *testing.T) {
	// Existent cameras are deleted to prevent interference with the `assert.Equal` function
	oldCams, err := ListCameras()
	if err != nil {
		t.Error(err.Error())
	}
	for _, cam := range oldCams {
		if err := DeleteCamera(cam.Id); err != nil {
			t.Error(err.Error())
		}
	}
	// Create test data
	if err := CreateRoom(RoomData{
		Id:          "redacted",
		Name:        "redacted",
		Description: "redacted",
	}); err != nil {
		t.Error(err.Error())
	}
	data := []Camera{
		{
			Id:     "test_redacted_1",
			Name:   "1",
			Url:    "http://hidden.com",
			RoomId: "redacted",
		},
		{
			Id:     "test_redacted_2",
			Name:   "2",
			Url:    "http://hidden2.com",
			RoomId: "redacted",
		},
	}
	for _, camera := range data {
		if err := CreateCamera(camera); err != nil {
			t.Error(err.Error())
		}
	}
	// Create a copy of the `data` slice but without urls
	dataCpy := make([]RedactedCamera, 0)
	for _, camera := range data {
		dataCpy = append(dataCpy, RedactedCamera{
			Id:   camera.Id,
			Name: camera.Name,
		})
	}
	fromFunc, err := ListCamerasRedacted()
	if err != nil {
		t.Error(err.Error())
	}
	assert.Equal(t, dataCpy, fromFunc)
}

func TestModifyCamera(t *testing.T) {
	// Create test room
	if err := CreateRoom(RoomData{Id: "test"}); err != nil {
		t.Error(err.Error())
	}

	table := []struct {
		Name     string
		Original Camera
		Modified Camera
	}{
		{
			Name: "only_modify_URL",
			Original: Camera{
				Id:     "test_3",
				Name:   "old_name",
				Url:    "old_url",
				RoomId: "test",
			},
			Modified: Camera{
				Id:     "test_3",
				Name:   "old_name",
				Url:    "https://example.com/1",
				RoomId: "test",
			},
		},
		{
			Name: "only_modify_name",
			Original: Camera{
				Id:     "test_4",
				Name:   "old_name",
				Url:    "old_url",
				RoomId: "test",
			},
			Modified: Camera{
				Id:     "test_4",
				Name:   "new_name",
				Url:    "old_url",
				RoomId: "test",
			},
		},
		{
			Name: "modify_URL_and_name",
			Original: Camera{
				Id:     "test_5",
				Name:   "old_name",
				Url:    "old_url",
				RoomId: "test",
			},
			Modified: Camera{
				Id:     "test_5",
				Name:   "new_name",
				Url:    "https://example.com/2",
				RoomId: "test",
			},
		},
	}
	for _, test := range table {
		t.Run(fmt.Sprintf("TestModifyCamera/%s", test.Name), func(t *testing.T) {
			if err := CreateCamera(test.Original); err != nil {
				t.Error(err.Error())
			}
			if err := ModifyCamera(test.Original.Id, test.Modified.Name, test.Modified.Url); err != nil {
				t.Error(err.Error())
			}
			camera, found, err := GetCameraById(test.Original.Id)
			if err != nil {
				t.Error(err.Error())
			}
			if !found {
				t.Errorf("Camera with id %s does not exist after modification", test.Original.Id)
			}
			assert.Equal(t, test.Modified, camera)
		})
	}
}

func TestDeleteCameraById(t *testing.T) {
	// Create test rooms
	if err := CreateRoom(RoomData{Id: "test3"}); err != nil {
		t.Error(err.Error())
	}
	table := []Camera{
		{
			Id:     "test_delete_1",
			Name:   "test 1",
			Url:    "http://example/com/1",
			RoomId: "test3",
		},
		{
			Id:     "test_delete_2",
			Name:   "test 2",
			Url:    "http://example.com/2",
			RoomId: "test3",
		},
	}
	for _, camera := range table {
		t.Run(fmt.Sprintf("TestDeleteCameraById/%s", camera.Id), func(t *testing.T) {
			// Create
			if err := CreateCamera(camera); err != nil {
				t.Error(err.Error())
			}

			// Validate creation
			cam, found, err := GetCameraById(camera.Id)
			if err != nil {
				t.Error(err.Error())
			}
			if !found {
				t.Errorf("Camera %s not found after creation", camera.Id)
			}
			assert.Equal(t, camera, cam)

			// Delete
			if err := DeleteCamera(cam.Id); err != nil {
				t.Error(err.Error())
			}

			// Validate deletion
			cam, found, err = GetCameraById(camera.Id)
			if err != nil {
				t.Error(err.Error())
			}
			if found {
				t.Errorf("Camera %s found after deletion", camera.Id)
			}
			assert.Empty(t, cam)
		})
	}
}

func TestListUserCameras(t *testing.T) {
	// Flush database first
	if err := initDB(true); err != nil {
		t.Error(err.Error())
	}
	// Create a room for the cameras
	if err := CreateRoom(RoomData{Id: "test"}); err != nil {
		t.Error(err.Error())
	}
	// Create test users
	if err := AddUser(FullUser{
		Username: "cameras",
	}); err != nil {
		t.Error(err.Error())
	}
	// Is required later in order to check if the function is able to return all cameras to a user with the modifyRooms permission
	if err := AddUser(FullUser{
		Username: "room_admin",
	}); err != nil {
		t.Error(err.Error())
	}
	if err := AddUserPermission("room_admin", PermissionModifyRooms); err != nil {
		t.Error(err.Error())
	}
	// Create test cameras
	cams := []Camera{
		{
			Id:     "test_user_1",
			Name:   "test 1",
			Url:    "http://example/com/1",
			RoomId: "test",
		},
		{
			Id:     "test_user_2",
			Name:   "test 2",
			Url:    "http://example.com/2",
			RoomId: "test",
		},
	}
	for _, cam := range cams {
		if err := CreateCamera(cam); err != nil {
			t.Error(err.Error())
		}
	}
	// Add an additional camera which will not be added to the permissions in order to check if the function does not return ungranted cameras
	if err := CreateCamera(Camera{
		Id:     "unlisted",
		RoomId: "test",
	}); err != nil {
		t.Error(err.Error())
	}
	// Check if the method returns something despite the user having no permission
	userCamsBefPerm, err := ListUserCameras("cameras")
	if err != nil {
		t.Error(err.Error())
	}
	assert.Empty(t, userCamsBefPerm)
	// Grant the user permission to all cameras
	for _, cam := range cams {
		if _, err := AddUserCameraPermission("cameras", cam.Id); err != nil {
			t.Error(err.Error())
		}
	}
	// Check the cameras again after giving the user permission, expected are just the granted cameras without the `unlisted` one
	userCamsAfterPerm, err := ListUserCameras("cameras")
	if err != nil {
		t.Error(err.Error())
	}
	assert.Equal(t, cams, userCamsAfterPerm)
	// Check the user-cameras of the `room_admin` user, expected are all cameras, inluding the `unlisted` one
	userCamsAdmin, err := ListUserCameras("room_admin")
	if err != nil {
		t.Error(err.Error())
	}
	allCameras, err := ListCameras()
	if err != nil {
		t.Error(err.Error())
	}
	assert.Equal(t, allCameras, userCamsAdmin)
}
