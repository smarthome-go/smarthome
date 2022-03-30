package automation

import (
	"errors"
	"fmt"
	"time"

	"github.com/MikMuellerDev/smarthome/core/database"
	"github.com/nathan-osman/go-sunrise"
)

// Utils for determining the times for sunrise and sunset
// Will be used if the automation's mode is set to either 'sunset' or 'sunrise'

type SunTime struct {
	Hour   uint
	Minute uint
}

// Returns (sunrise, sunset) based on the provided coordinates
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

// Given a jobId and if it should use sunrise or sunset, this function modifies the running time of a given automation
func updateJobTime(id uint, useSunRise bool) error {
	// Obtain the server's configuration in order to determine the latitude and longitude
	config, found, err := database.GetServerConfiguration()
	if err != nil || !found {
		log.Error("Failed to update job launch time: could not obtain the server's configuration")
		return errors.New("could not update launch time: failed to obtain server config")
	}
	// Retrieve the current job in order to get its current cron expression (for the days)
	job, found, err := database.GetAutomationById(id)
	if err != nil || !found {
		return errors.New("could not update launch time: invalid id supplied")
	}
	// Calculate both times
	sunRise, sunSet := CalculateSunRiseSet(config.Latitude, config.Longitude)
	var finalTime SunTime // Will be set according to `useSunRise`
	if useSunRise {
		finalTime = sunRise
	} else {
		finalTime = sunSet
	}
	// Get the days from the cron expression
	days, err := GetDaysFromCronExpression(job.CronExpression)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to get days from cron expression '%s': Error: %s", job.CronExpression, err))
		return err
	}
	cronExpression, err := GenerateCronExpression(uint8(finalTime.Hour), uint8(finalTime.Minute), days)
	if err != nil {
		return err
	}
	if err := ModifyAutomationById(id, database.AutomationWithoutIdAndUsername{
		Name:           job.Name,
		Description:    job.Description,
		CronExpression: cronExpression,
		HomescriptId:   job.HomescriptId,
		Enabled:        job.Enabled,
		TimingMode:     job.TimingMode,
	}); err != nil {
		log.Error("Failed to update launch time of automation: could not modify automation: ", err.Error())
		return err
	}
	log.Trace(fmt.Sprintf("Successfully updated the launch time of automation %d", id))
	return nil
}
