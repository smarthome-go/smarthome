package weather

import (
	"fmt"

	owm "github.com/briandowns/openweathermap"
	"github.com/smarthome-go/smarthome/core/database"
)

func GetWeather() (owm.CurrentWeatherData, error) {
	// Retrieve the server configuration
	config, found, err := database.GetServerConfiguration()
	if err != nil {
		return owm.CurrentWeatherData{}, err
	}
	if !found {
		return owm.CurrentWeatherData{}, fmt.Errorf("server configuration not found")
	}
	fmt.Println(config.OpenWeatherMapApiKey)
	// Get the current weather
	w, err := owm.NewCurrent("C", "en", config.OpenWeatherMapApiKey)
	if err != nil {
		return owm.CurrentWeatherData{}, err
	}
	if err := w.CurrentByCoordinates(&owm.Coordinates{
		Longitude: float64(config.Longitude),
		Latitude:  float64(config.Latitude),
	}); err != nil {
		return owm.CurrentWeatherData{}, err
	}
	return *w, nil
}
