package hardware

import (
	"fmt"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/event"
)

// This file's functions are being used for calculating new power usage summaries (which is triggered on every switch power change)

// Just like the equivalent in the database module
// except the time is represented using Unix-millis
type PowerDrawDataPointUnixMillis struct {
	Id   uint                   `json:"id"`
	Time uint                   `json:"time"` // Is represented as Unix-millis
	On   database.PowerDrawData `json:"on"`
	Off  database.PowerDrawData `json:"off"`
}

// Takes a slice of power data points as an input and outputs it whilst filtering the data for semantic and visual imperfections, such as redundant measurements
func filterPowerData(input []database.PowerDataPoint) (newData []database.PowerDataPoint, iDsToBeDeleted []uint) {
	// Step 1: filter out redundant measurements
	// Calculate the length once (for performance)
	dataPoints := len(input)
	// Contains the final, filtered data
	newData = make([]database.PowerDataPoint, 0)
	// Specifies which ids can be safely deleted from the data set (the filtered out data)
	iDsToBeDeleted = make([]uint, 0)
	// Filter out the data
	for pointIndex, point := range input {
		// Check if one lookback and one lookahead is possible
		if /* Lookback is not possible*/ pointIndex-1 < 0 || /* Lookahead is not possible */ pointIndex+1 > dataPoints-1 {
			newData = append(newData, point)
			continue
		}
		lookback := input[pointIndex-1]
		lookahead := input[pointIndex+1]
		if lookback.On.Watts == point.On.Watts && point.On.Watts == lookahead.On.Watts {
			iDsToBeDeleted = append(iDsToBeDeleted, point.Id)
			continue
		}
		newData = append(newData, point)
	}
	return newData, iDsToBeDeleted
}

// Takes a snapshot of the current power states and transforms them into a power data point
// Returns `onData`, `offData` and an `error`
func generateSnapshot() (database.PowerDrawData, database.PowerDrawData, error) {
	// Get the current power states
	powerStates, err := database.GetPowerStates()
	if err != nil {
		return database.PowerDrawData{}, database.PowerDrawData{}, err
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

	// NOTE: If the total watts are equal to 0, stop here and do not calculate the percent (it will lead to errors)
	if totalWatts == 0 {
		return onData, offData, nil
	}

	// After the on + off data has been calculated, leverage the grand total watt count in order to calculate the individual percent numbers
	onData.Percent = float64(onData.Watts) / float64(totalWatts) * 100
	offData.Percent = float64(offData.Watts) / float64(totalWatts) * 100

	return onData, offData, nil
}

// Takes a snapshot of the current power draw and inserts it into the database
func SaveCurrentPowerUsage() error {
	// Generate a snapshot
	onData, offData, err := generateSnapshot()
	if err != nil {
		return err
	}
	// Insert the snapshot data into the database
	if _, err = database.AddPowerUsagePoint(
		onData,
		offData,
	); err != nil {
		return err
	}
	// Filter the data after the insertion and delete redundant data records
	powerUsageData, err := database.GetPowerUsageRecords(24)
	if err != nil {
		return err
	}
	// Delete the redundant records one by one
	_, toBeDeleted := filterPowerData(powerUsageData)
	for _, record := range toBeDeleted {
		if err := database.DeletePowerUsagePointById(record); err != nil {
			return err
		}
		log.Debug(fmt.Sprintf("Deleted redundant power usage data point from dataset. (ID: %d)", record))
	}
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
	event.Trace("Power Draw Snapshot Saved", "A snapshot of the current power draw has been generated and saved in the database")
	log.Debug("A snapshot of the current power draw has been generated and saved in the database")
}

// Acts like a wrapper for the `database.GetPowerUsageRecords`
// The main difference is that dates are transformed into unix-millis (which are easier to parse for any API client)
func GetPowerUsageRecordsUnixMillis(maxAgeHours int) ([]PowerDrawDataPointUnixMillis, error) {
	dbData, err := database.GetPowerUsageRecords(maxAgeHours)
	if err != nil {
		return nil, err
	}
	// Transform the data into a slice which uses the new struct
	returnValue := make([]PowerDrawDataPointUnixMillis, 0)
	for _, record := range dbData {
		returnValue = append(returnValue, PowerDrawDataPointUnixMillis{
			Id:   record.Id,
			Time: uint(record.Time.UnixMilli()),
			On:   record.On,
			Off:  record.Off,
		})
	}
	return returnValue, err
}

// Sets up a scheduler which triggers the flushing of old power usage records
func StartPowerUsageSnapshotScheduler() error {
	scheduler := gocron.NewScheduler(time.Local)
	if _, err := scheduler.Every(1).Hours().Do(SaveCurrentPowerUsageWithLogs); err != nil {
		return err
	}
	scheduler.StartAsync()
	log.Debug("Successfully started power usage snapshot scheduler")
	return nil
}
