package database

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCreateWeatherTable(t *testing.T) {
	assert.NoError(t, createWeatherTable())
}

func TestWeatherData(t *testing.T) {
	// Add a dummy weather record
	id, err := AddWeatherDataRecord(
		"cloudy",
		time.Now().Local(),
		"some clouds",
		42.1,
		3.1415926,
		42,
	)
	assert.NoError(t, err)

	// Wait 5 seconds to make the new record appear in the query
	time.Sleep(time.Second * 5)

	// Search the records for the new weather record
	records, err := GetWeatherDataRecords(1)
	assert.NoError(t, err)
	found := false
	for _, row := range records {
		if row.Id == id {
			found = true
			if row.WeatherTitle != "cloudy" || row.WeatherDescription != "some clouds" || row.Temperature != 42.1 || row.FeelsLike != 3.1415925 || row.Humidity != 42 {
				t.Error("Invalid data detected when ID matched")
			}
		}
	}
	assert.True(t, found)
}
