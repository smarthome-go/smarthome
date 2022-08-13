package database

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCreatePowerUsageTable(t *testing.T) {
	assert.NoError(t, createPowerUsageTable())
}

func TestAddPowerUsagePoint(t *testing.T) {
	// Create mock data
	onData := PowerDrawData{
		SwitchCount: 10,
		Watts:       1000,
		Percent:     83.33,
	}
	offData := PowerDrawData{
		SwitchCount: 2,
		Watts:       200,
		Percent:     16.66,
	}
	// Insert a new record into the database
	insertedId, err := AddPowerUsagePoint(onData, offData)
	assert.NoError(t, err)
	// Search the records for the inserted id and assert data equality
	records, err := GetPowerUsageRecords(1)
	assert.NoError(t, err)
	for _, point := range records {
		if point.Id == insertedId {
			// Assert data equality
			assert.Equal(t, onData, point.On)
			assert.Equal(t, offData, point.Off)
			return
		}
	}
	t.Errorf("Id %d was not found in records after insertion", insertedId)
}

func TestFlushPowerUsagePoints(t *testing.T) {
	// Mock data
	onData := PowerDrawData{
		Watts: 1000,
	}
	offData := PowerDrawData{
		Watts: 200,
	}
	t.Run("flush_all", func(t *testing.T) {
		// Insert a new record into the database
		insertedId, err := AddPowerUsagePoint(onData, offData)
		assert.NoError(t, err)
		// Search the records for the inserted id
		records, err := GetPowerUsageRecords(1)
		assert.NoError(t, err)
		found := false
		for _, point := range records {
			if point.Id == insertedId {
				found = true
			}
		}
		if !found {
			t.Errorf("Id %d was not found in records after insertion", insertedId)
		}
		// Must wait around 5 Seconds for the point's time to be considered old
		time.Sleep(time.Second * 5)
		// Delete every record from the table
		affectedRows, err := FlushPowerUsageRecords(0)
		assert.NoError(t, err)
		if affectedRows == 0 {
			t.Errorf("Flushing off all power usage points failed: affected 0 rows but should have affected at least 1")
		}
		// Search the records for the inserted id again
		records, err = GetPowerUsageRecords(1)
		assert.NoError(t, err)
		found = false
		for _, point := range records {
			if point.Id == insertedId {
				found = true
			}
		}
		if found {
			t.Errorf("Id %d was found in records after record deletion", insertedId)
		}
	})
	t.Run("flush_old", func(t *testing.T) {
		// Insert another point into the tatabase
		insertedId, err := AddPowerUsagePoint(onData, offData)
		assert.NoError(t, err)
		// Must wait around 5 Seconds for the point's time to be considered old (safety measure to be sure)
		time.Sleep(time.Second * 5)
		// Delete recors again, this time those older than 1 hour (should be none due to previous deletion)
		affectedRows, err := FlushPowerUsageRecords(1)
		assert.NoError(t, err)
		if affectedRows > 0 {
			t.Errorf("Flushing off old power usage points failed: affected > 0 rows but should have affected 0")
		}
		// Search the records for the inserted id again
		records, err := GetPowerUsageRecords(1)
		assert.NoError(t, err)
		found := false
		for _, point := range records {
			if point.Id == insertedId {
				found = true
			}
		}
		if !found {
			t.Errorf("Id %d was bot found in records after record deletion which should have not affected this id", insertedId)
		}
	})
}
