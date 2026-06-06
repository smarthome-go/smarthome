package driver

import (
	"fmt"
	"sync"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/smarthome-go/smarthome/core/database"
	driverTypes "github.com/smarthome-go/smarthome/core/device/driver/types"
	"github.com/smarthome-go/smarthome/core/event"
)

const savePowerUsageEveryNMinutes = 10

type PowerDrawDataPointUnixMillis struct {
	Id   uint64                 `json:"id"`
	Time uint64                 `json:"time"`
	On   database.PowerDrawData `json:"on"`
	Off  database.PowerDrawData `json:"off"`
}

type DevicePowerCacheEntry struct {
	State          bool
	PowerDrawWatts uint
}

var powerCache = struct {
	sync.RWMutex
	entries map[string]DevicePowerCacheEntry
}{
	entries: make(map[string]DevicePowerCacheEntry),
}

func PopulatePowerCache() {
	if !Manager.IsInitialized() {
		return
	}

	devices, err := Manager.ListAllDevicesRich()
	if err != nil {
		log.Errorf("Failed to populate power cache: %s", err.Error())
		return
	}

	powerCache.Lock()
	defer powerCache.Unlock()

	for _, dev := range devices {
		if !dev.Extractions.Config.Capabilities.Has(DeviceCapabilityPower) {
			continue
		}
		powerCache.entries[dev.Shallow.ID] = DevicePowerCacheEntry{
			State:          dev.Extractions.PowerInformation.State,
			PowerDrawWatts: dev.Extractions.PowerInformation.PowerDrawWatts,
		}
	}
}

func UpdateDevicePowerCache(deviceID, vendorID, modelID string) {
	ids := driverTypes.DriverInvocationIDs{
		DeviceID: &deviceID,
		VendorID: vendorID,
		ModelID:  modelID,
	}

	stateOut, _, err := Manager.InvokeDriverReportPowerState(ids)
	if err != nil {
		log.Warnf("Power cache update: failed to get power state for device %s: %s", deviceID, err.Error())
		return
	}

	drawOut, _, err := Manager.InvokeDriverReportPowerDraw(ids)
	if err != nil {
		log.Warnf("Power cache update: failed to get power draw for device %s: %s", deviceID, err.Error())
		return
	}

	powerCache.Lock()
	powerCache.entries[deviceID] = DevicePowerCacheEntry{
		State:          stateOut.State,
		PowerDrawWatts: drawOut.Watts,
	}
	powerCache.Unlock()
}

func filterPowerData(input []database.PowerDataPoint) (newData []database.PowerDataPoint, iDsToBeDeleted []uint64) {
	dataPoints := len(input)
	newData = make([]database.PowerDataPoint, 0)
	iDsToBeDeleted = make([]uint64, 0)

	for pointIndex, point := range input {
		if pointIndex-1 < 0 || pointIndex+1 > dataPoints-1 {
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

func generateSnapshot() (onData database.PowerDrawData, offData database.PowerDrawData, err error) {
	var totalWatts uint = 0

	powerCache.RLock()
	defer powerCache.RUnlock()

	for _, entry := range powerCache.entries {
		if entry.State {
			onData.SwitchCount++
			onData.Watts += entry.PowerDrawWatts
		} else {
			offData.SwitchCount++
			offData.Watts += entry.PowerDrawWatts
		}
		totalWatts += entry.PowerDrawWatts
	}

	if totalWatts == 0 {
		return onData, offData, nil
	}

	onData.Percent = float64(onData.Watts) / float64(totalWatts) * 100
	offData.Percent = float64(offData.Watts) / float64(totalWatts) * 100

	return onData, offData, nil
}

func SaveCurrentPowerUsage() error {
	config, _, err := database.GetServerConfiguration()
	if err != nil {
		return err
	}

	if config.LockDownMode {
		log.Trace("Lockdown mode is enabled, not generating power snapshot")
		return nil
	}

	if !Manager.IsInitialized() {
		log.Trace("Power usage manager is not initialized, not generating power snapshot")
		return nil
	}

	onData, offData, err := generateSnapshot()
	if err != nil {
		return err
	}

	if _, err = database.AddPowerUsagePoint(
		onData,
		offData,
		time.Now(),
	); err != nil {
		return err
	}

	powerUsageData, err := database.GetPowerUsageRecords(24)
	if err != nil {
		return err
	}

	_, toBeDeleted := filterPowerData(powerUsageData)
	for _, record := range toBeDeleted {
		if err := database.DeletePowerUsagePointById(record); err != nil {
			return err
		}
		log.Debug(fmt.Sprintf("Deleted redundant power usage data point from dataset. (ID: %d)", record))
	}

	return err
}

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

func GetPowerUsageRecordsUnixMillis(maxAgeHours int) ([]PowerDrawDataPointUnixMillis, error) {
	dbData, err := database.GetPowerUsageRecords(maxAgeHours)
	if err != nil {
		return nil, err
	}

	returnValue := make([]PowerDrawDataPointUnixMillis, 0)
	for _, record := range dbData {
		returnValue = append(returnValue, PowerDrawDataPointUnixMillis{
			Id:   record.Id,
			Time: uint64(record.Time.UnixMilli()),
			On:   record.On,
			Off:  record.Off,
		})
	}

	return returnValue, err
}

func StartPowerUsageSnapshotScheduler() error {
	PopulatePowerCache()

	scheduler := gocron.NewScheduler(time.Local)
	if _, err := scheduler.Every(savePowerUsageEveryNMinutes).Minute().Do(SaveCurrentPowerUsageWithLogs); err != nil {
		return err
	}
	scheduler.StartAsync()
	log.Debug("Successfully started power usage snapshot scheduler")
	return nil
}
