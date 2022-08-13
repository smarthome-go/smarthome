package hardware

import (
	"fmt"

	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/event"
)

// This file's functions are being used for calculating new power usage summaries (which is triggered on every switch power change)

type PowerDrawDataPoint struct {
	OnData  database.PowerDrawData
	OffData database.PowerDrawData
}

// Takes a snapshot of the current power states and transforms them into a power data point
func generateSnapshot() (PowerDrawDataPoint, error) {
	// Get the current power states
	powerStates, err := database.GetPowerStates()
	if err != nil {
		return PowerDrawDataPoint{}, err
	}
	// Will hold the sum off the power draw of all switches, regardless of whether they are active or disabled
	var totalWatts uint16 = 0
	// Collects information about the active switches
	onData := database.PowerDrawData{
		SwitchCount: 0,
		Watts:       0,
		Percent:     0,
	}
	// Collects information about the deactivated switches
	offData := database.PowerDrawData{
		SwitchCount: 0,
		Watts:       0,
		Percent:     0,
	}
	// Loop over all switches
	for _, sw := range powerStates {
		// If the current switch is active, account for in int the `onData`
		if sw.PowerOn {
			onData.SwitchCount++           // Increment the switch count of all active switches by one
			onData.Watts += uint(sw.Watts) // Add the power draw of the current switch to the total of the active switches
		} else {
			offData.SwitchCount++           // Increment the switch count of all passive switches by one
			offData.Watts += uint(sw.Watts) // Add the power draw of the current switch to the total of the passive switches
		}
		// Regardless of the power state, increment the total watt count
		totalWatts += sw.Watts
	}
	// After the on + off data has been calculated, leverage the grand total watt count in order to calculate the individual percent numbers
	onData.Percent = float64(onData.Watts) / float64(totalWatts) * 100
	offData.Percent = float64(offData.Watts) / float64(totalWatts) * 100

	// Create a data point from the individual data structs
	dataPoint := PowerDrawDataPoint{
		OnData:  onData,
		OffData: offData,
	}
	return dataPoint, nil
}

// Takes a snapshot of the current power draw and inserts it into the database
func SaveCurrentPowerUsage() error {
	// Generate a snapshot
	data, err := generateSnapshot()
	if err != nil {
		return err
	}
	// Insert the snapshot data into the database
	_, err = database.AddPowerUsagePoint(
		data.OnData,
		data.OffData,
	)
	return err
}

// Wrapper around `saveCurrentPowerUsage` which handles errors through logging
// Is also more verbose than the original function
func SaveCurrentPowerUsageWithLogs() {
	log.Trace("Saving snapshot of current power draw...")
	if err := SaveCurrentPowerUsage(); err != nil {
		log.Error("Could not save snapshot of current power draw: ", err.Error())
		event.Error("Power Draw Snapshot Error", fmt.Sprintf("Could not save snapshot of the current power draw: %s", err.Error()))
		return
	}
	event.Debug("Power Draw Snapshot Saved", "A snapshot of the current power draw has been generated and saved in the database")
	log.Debug("A snapshot of the current power draw has been generated and saved in the database")
}
