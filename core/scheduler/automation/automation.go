package automation

import (
	"fmt"
	"strings"
	"time"

	"github.com/MikMuellerDev/smarthome/core/database"
	"github.com/go-co-op/gocron"
	"github.com/lnquy/cron"
	"github.com/sirupsen/logrus"
)

// The main scheduler which will run all automation jobs
var scheduler *gocron.Scheduler

var log *logrus.Logger

func InitLogger(logger *logrus.Logger) {
	log = logger
}

// Generates a cron expression based on hour, minute, and a slice of days on which the action will run
func GenerateCronExpression(hour uint8, minute uint8, days []uint8) (string, error) {
	output := [5]string{"", "", "*", "*", ""}
	output[0] = fmt.Sprintf("%d", minute)
	output[1] = fmt.Sprintf("%d", hour)
	if len(days) > 7 {
		log.Error("The maximum amount of days allowed are 7")
		return "", fmt.Errorf("Amount of days should not be greater than 7")
	}
	if len(days) == 7 {
		// Set the days to '*' when all days are included in the slice, does not check for duplicate days
		output[4] = "*"
		return strings.Join(output[:], " "), nil
	}
	for index, day := range days {
		output[4] += fmt.Sprintf("%d", day)
		if index < len(days)-1 {
			output[4] += ","
		}
	}
	return strings.Join(output[:], " "), nil
}

// Generates a human-readable string from a given cron expression
func generateHumanReadableCronExpression(expr string) (string, error) {
	descriptor, err := cron.NewDescriptor()
	if err != nil {
		log.Error("Failed to parse cron expression into human readable format: ", err.Error())
		return "", err
	}
	output, err := descriptor.ToDescription(expr, cron.Locale_en)
	if err != nil {
		log.Error("Failed to parse cron expression into human readable format: ", err.Error())
		return "", err
	}
	return output, nil
}

// Validates a given cron expression, returns false if the given cron expression is invalid
func IsValidCronExpression(expr string) bool {
	descriptor, err := cron.NewDescriptor()
	if err != nil {
		return false
	}
	if _, err = descriptor.ToDescription(expr, cron.Locale_en); err != nil {
		return false
	}
	return true
}

// Activates all jobs saved in the database, will be used when the server was restarted
// If a non-critical error occurs, for example the failure to setup a job, it will be returned
// This function will not cancel directly if an error occurs in order to preserve the automation system's uptime
func startSavedAutomations() error {
	automations, err := database.GetAutomations()
	if err != nil {
		log.Error("Failed to activate automation system: database failure whilst starting saved automations: ", err.Error())
		return err // This is a critical error which can not be recovered
	}
	for _, automation := range automations {
		if !IsValidCronExpression(automation.CronExpression) {
			log.Error(fmt.Sprintf("Could not activate automation '%d': invalid cron expression", automation.Id))
			continue // non-critical error
		}
		automationJob := scheduler.Cron(automation.CronExpression)
		automationJob.Tag(fmt.Sprintf("%d", automation.Id))
		automationJob.Do(automationRunnerFunc, automation.Id)
		log.Debug(fmt.Sprintf("Successfully activated automation '%d' of user '%s'", automation.Id, automation.Owner))
	}
	log.Debug(fmt.Sprintf("Activated saved automations: registered %d total automation jobs", len(automations)))
	return nil
}

// Initializes the scheduler
func Init() error {
	scheduler = gocron.NewScheduler(time.Local)
	scheduler.TagsUnique()
	if err := startSavedAutomations(); err != nil {
		log.Error("Failed to activate automation system: could not activate persistent jobs: ", err.Error())
	}
	scheduler.StartAsync()
	log.Info("Successfully activated automation scheduler system")
	return nil
}
