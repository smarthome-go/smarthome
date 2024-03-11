package automation

import (
	"time"

	"github.com/nathan-osman/go-sunrise"
)

// Utils for determining the times for sunrise and sunset
// Will be used if the automation's mode is set to either 'sunset' or 'sunrise'

type SunTime struct {
	Hour   uint
	Minute uint
}

// Returns (sunrise, sunset) based on the provided coordinates which are stored in the server configuration
func CalculateSunRiseSet(lat float32, lon float32) (SunTime, SunTime) {
	sunRise, sunSet := sunrise.SunriseSunset(
		float64(lat), float64(lon),
		time.Now().Year(), time.Now().Month(), time.Now().Day(),
	)

	return SunTime{
			uint(sunRise.Local().Hour()), uint(sunRise.Local().Minute()),
		}, SunTime{
			uint(sunSet.Local().Hour()), uint(sunSet.Local().Minute()),
		}
}
