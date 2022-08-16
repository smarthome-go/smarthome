package weather

import (
	"fmt"
	"time"

	owm "github.com/briandowns/openweathermap"
	"github.com/smarthome-go/smarthome/core/database"
)

type WeatherMeasurement struct {
	Id                 uint    `json:"id"`
	Time               uint    `json:"time"` // Time as Unix-millis
	WeatherTitle       string  `json:"weatherTitle"`
	WeatherDescription string  `json:"weatherDescription"`
	Temperature        float32 `json:"temperature"`
	FeelsLike          float32 `json:"feelsLike"`
	Humidity           uint8   `json:"humidity"`
}

var (
	ErrInvalidApiKey = fmt.Errorf("invalid owm api key")
	ErrLenWeather0   = fmt.Errorf("the owm response contains no weather information")
)

// Makes an API-request to OWM in order to get the latest weather data
func fetchWeather() (owm.CurrentWeatherData, error) {
	// Retrieve the server configuration
	config, found, err := database.GetServerConfiguration()
	if err != nil {
		return owm.CurrentWeatherData{}, err
	}
	if !found {
		return owm.CurrentWeatherData{}, fmt.Errorf("server configuration not found")
	}
	// Fetch the current weather data from their API
	w, err := owm.NewCurrent(
		"C",
		"en",
		config.OpenWeatherMapApiKey,
	)
	if err != nil {
		return owm.CurrentWeatherData{}, err
	}
	// Request the local weather using their coordinates API
	if err := w.CurrentByCoordinates(&owm.Coordinates{
		Longitude: float64(config.Longitude),
		Latitude:  float64(config.Latitude),
	}); err != nil {
		return owm.CurrentWeatherData{}, err
	}
	// Check if the weather list contains at least 1 item
	if len(*&w.Weather) == 0 {
		return owm.CurrentWeatherData{}, ErrLenWeather0
	}
	return *w, nil
}

func GetCurrentWeather() (WeatherMeasurement, error) {
	// Attempt to retrieve the weather records from the last 5 minutes (use cache if possible)
	cached, err := database.GetWeatherDataRecords(5)
	if err != nil {
		return WeatherMeasurement{}, err
	}
	// If there is already a cached version available, return the latest one
	if len(cached) > 0 {
		fmt.Println(cached[0].Id)
		return transformWeatherStruct(cached[0]), nil
	}
	// Otherwise, new data must be fetched and inserted into the database
	freshData, err := fetchWeather()
	if err != nil {
		return WeatherMeasurement{}, err
	}
	newLabel := freshData.Weather[0].Main
	newDescription := freshData.Weather[0].Main
	// Insert the new record into the datbase
	id, err := database.AddWeatherDataRecord(
		newLabel,
		newDescription,
		float32(freshData.Main.Temp),
		float32(freshData.Main.FeelsLike),
		uint8(freshData.Main.Humidity),
	)
	if err != nil {
		return WeatherMeasurement{}, err
	}
	// Return a final version
	return WeatherMeasurement{
		Id:                 id,
		Time:               uint(time.Now().UnixMilli()),
		WeatherTitle:       newLabel,
		WeatherDescription: newDescription,
		Temperature:        float32(freshData.Main.Temp),
		FeelsLike:          float32(freshData.Main.FeelsLike),
		Humidity:           uint8(freshData.Main.Humidity),
	}, nil
}

// Transforms the weather data struct from the database into the struct defined in this module
// The only difference is that the Go time.Time is transformed into Unix-Millis
func transformWeatherStruct(input database.WeatherMeasurement) WeatherMeasurement {
	return WeatherMeasurement{
		Id:                 input.Id,
		Time:               uint(input.Time.UnixMilli()),
		WeatherTitle:       input.WeatherTitle,
		WeatherDescription: input.WeatherDescription,
		Temperature:        input.Temperature,
		FeelsLike:          input.FeelsLike,
		Humidity:           input.Humidity,
	}
}
