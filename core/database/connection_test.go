package database

import "testing"

func TestConnection(t *testing.T) {
	if _, err := connection(); err != nil {
		t.Error(err.Error())
		return
	}
}

func TestCheckDatabase(t *testing.T) {
	if err := CheckDatabase(); err != nil {
		t.Error(err.Error())
		return
	}
}
