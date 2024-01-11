package device

import (
	"fmt"

	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/homescript"
)

func ExtractDriverInformationFromDevice(device database.Device) error {
	// TODO: implementation is missing

	// Get the driver
	driver, found, err := database.GetDeviceDriver(device.VendorId, device.ModelId)
	if err != nil {
		return err
	}

	if !found {
		panic("BUG: device has no driver entry")
	}

	// Invoke the extractor
	info, err := homescript.ExtractDriverInfo(driver)
	if err != nil {
		return err
	}

	// TODO: do smth with info
	fmt.Printf("INFO: %v\n", info)

	return nil
}
