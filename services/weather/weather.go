package weather

import (
	"fmt"
	"time"

	owm "github.com/briandowns/openweathermap"
	"github.com/nathan-osman/go-sunrise"
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

	Sunrise uint `json:"sunrise"`
	Sunset  uint `json:"sunset"`
}

var (
	ErrInvalidApiKey = fmt.Errorf("invalid owm api key")
	ErrLenWeather0   = fmt.Errorf("the owm response contains no weather information")
)

// Makes an API-request to OWM in order to get the latest weather data
func fetchWeather(latitude float64, longitude float64, owmKey string) (owm.CurrentWeatherData, error) {
	// Fetch the current weather data from their API
	w, err := owm.NewCurrent(
		"C",
		"en",
		owmKey,
	)
	if err != nil {
		return owm.CurrentWeatherData{}, err
	}
	// Request the local weather using their coordinates API
	if err := w.CurrentByCoordinates(&owm.Coordinates{
		Longitude: longitude,
		Latitude:  latitude,
	}); err != nil {
		return owm.CurrentWeatherData{}, err
	}
	// Check if the weather list contains at least 1 item
	if len(w.Weather) == 0 {
		return owm.CurrentWeatherData{}, ErrLenWeather0
	}
	return *w, nil
}

func GetCurrentWeather() (WeatherMeasurement, error) {
	// Retrieve the server configuration
	config, found, err := database.GetServerConfiguration()
	if err != nil {
		return WeatherMeasurement{}, err
	}
	if !found {
		return WeatherMeasurement{}, fmt.Errorf("server configuration not found")
	}

	// Calculate the sunrise / sunset time
	sunRise, sunSet := sunrise.SunriseSunset(
		float64(config.Latitude), float64(config.Longitude),
		time.Now().Year(), time.Now().Month(), time.Now().Day(),
	)

	// Attempt to retrieve the weather records from the last 5 minutes (use cache if possible)
	cached, err := database.GetWeatherDataRecords(5)
	if err != nil {
		return WeatherMeasurement{}, err
	}
	// If there is already a cached version available, return the latest one
	if len(cached) > 0 {
		// Just update the time of the cached item
		cachedLatest := transformWeatherStruct(cached[0])
		cachedLatest.Sunrise = uint(sunRise.UnixMilli())
		cachedLatest.Sunset = uint(sunSet.UnixMilli())
		return cachedLatest, nil
	}
	// Otherwise, new data must be fetched and inserted into the database
	freshData, err := fetchWeather(
		float64(config.Latitude),
		float64(config.Longitude),
		config.OpenWeatherMapApiKey,
	)
	if err != nil {
		return WeatherMeasurement{}, err
	}

	newLabel := freshData.Weather[0].Main
	newDescription := freshData.Weather[0].Description

	// Insert the new record into the database
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
		Sunrise:            uint(sunRise.UnixMilli()),
		Sunset:             uint(sunSet.UnixMilli()),
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
