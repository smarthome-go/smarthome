package scheduler

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/lnquy/cron"
	"github.com/sirupsen/logrus"
)

// The main scheduler which will run all jobs
var scheduler *gocron.Scheduler

var log *logrus.Logger

func InitLogger(logger *logrus.Logger) {
	log = logger
}

// Generates a cron expression based on hour, minute, and a slice of days on which the action will run
func generateCronExpression(hour uint8, minute uint8, days []uint8) (string, error) {
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

// Initializes the scheduler
func Init() error {
	scheduler = gocron.NewScheduler(time.Local)
	scheduler.TagsUnique()
	scheduler.StartAsync()
	return nil
}
