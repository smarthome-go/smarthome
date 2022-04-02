package database

import "testing"

func TestCreateCameraTable(t *testing.T) {
	if err := createCameraTable(); err != nil {
		t.Error(err.Error())
	}
}
