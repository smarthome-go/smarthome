package homescript

import (
	"fmt"
	"testing"
	"time"

	"github.com/smarthome-go/smarthome/core/database"
	"github.com/stretchr/testify/assert"
)

func TestDefaultDriverHmsCode(t *testing.T) {
	driverBaseIDNumeric := time.Now().UnixMilli()
	driverBaseID := fmt.Sprint(driverBaseIDNumeric)

	vendorID := fmt.Sprintf("%s_vendor", driverBaseID)
	const modelID = "default_test"

	hmsErr, dbErr := CreateDriver(
		vendorID,
		modelID,
		"Default Driver",
		SemanticVersion{
			Major: 1,
			Minor: 2,
			Patch: 3,
		}.String(),
		&database.DefaultDriverHomescriptCode,
	)

	assert.NoError(t, dbErr)
	assert.NoError(t, hmsErr)

	// Create a room.
	assert.NoError(t, database.CreateRoom(database.RoomData{
		Id:          driverBaseID,
		Name:        "Default Driver Test",
		Description: "/",
	}))

	deviceID := fmt.Sprintf("dev_%s", driverBaseID)

	// Create a device for that driver.
	driverFound, hmsErr, dbErr := CreateDevice(
		database.DEVICE_TYPE_OUTPUT,
		deviceID,
		"Default Driver Device",
		driverBaseID,
		vendorID,
		modelID,
	)

	assert.NoError(t, dbErr)
	assert.NoError(t, hmsErr)
	assert.True(t, driverFound)

	// Run the driver.
	// TODO: allow invocation without being tied to a device.
	hmsErrs, dbErr := InvokeValidateCheckDriver(
		DriverInvocationIDs{
			deviceID: deviceID,
			vendorID: vendorID,
			modelID:  modelID,
		},
	)

	assert.NoError(t, dbErr)
	assert.Empty(t, hmsErrs)
}
