ipackage database

import (
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
		return
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
		if err := CreateCamera(camera); err != nil {
			t.Error(err.Error())
		}

		cam, found, err := GetCameraById(camera.Id)
		if err != nil {
			t.Error(err.Error())
		}
		if !found {
			t.Errorf("Camera %s not found after creation", camera.Id)
		}
		assert.Equal(t, camera, cam, "camera from id has invalid metadata")
	}

	// Check the listing function
	cameras, err := ListCameras()
	if err != nil {
		t.Error(err.Error())
	}
	assert.Equal(t, cameras, table, "listed cameras do not match table")
}

func TestModifyCamera(t * testing.T) {
  // Create test room
  if err := CreateRoom(Room{Id: "test"}); err != nil {
    t.Error(err.Error())
  }

  table := []struct {
    Original Camera
    Modified Camera
  } {
    {
      Original: Camera {
        Id: "test_3"
        RoomId: "test"
      },
      Modified: Camera {
        Id: "test_3",
        RoomId: "test",
        Name: "Test Name",
        Url: "https://example.com/1"
      }
    }
  }
  for _, test := range table {
    if err := CreateCamera(test.Original); err != nil {
      t.Error(err.Error())
    }
  }
}
