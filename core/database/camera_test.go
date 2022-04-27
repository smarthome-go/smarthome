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
		assert.Equal(t, cameras, table, "listed cameras do not match table")
	})
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
