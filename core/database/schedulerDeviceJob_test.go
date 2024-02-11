package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateSchedulerSwitchesTable(t *testing.T) {
	assert.NoError(t, createSchedulerDeviceJobTable())
}
