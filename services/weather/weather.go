package weather

import (
	"fmt"

	owm "github.com/briandowns/openweathermap"
	"github.com/smarthome-go/smarthome/core/database"
)

func GetWeather() error {
	// Retrieve the server configuration
	config, found, err := database.GetServerConfiguration()
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("server configuration not found")
	}
	// Get the current weather
	w, err := owm.NewCurrent("C", "en", "")
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	w.CurrentByCoordinates(&owm.Coordinates{
		Longitude: float64(config.Longitude),
		Latitude:  float64(config.Latitude),
	})

	fmt.Println(w)

	return nil
}
