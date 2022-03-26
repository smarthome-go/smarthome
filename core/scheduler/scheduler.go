package scheduler

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/lnquy/cron"
	"github.com/sirupsen/logrus"
)

type Day uint8

const (
	Sunday Day = iota
	Monday
	TuesDay
	Wednesday
	Thursday
	Friday
	Saturday
)

// The main scheduler which will run all jobs
var scheduler *gocron.Scheduler

var log *logrus.Logger

func InitLogger(logger *logrus.Logger) {
	log = logger
}

// Generates a cron expression based on hour, minute, and a slice of days on which the action will run
func generateCronExpression(hour uint8, minute uint8, days []Day) (string, error) {
	output := [5]string{"", "", "*", "*", ""}
	output[0] = fmt.Sprintf("%d", minute)
	output[1] = fmt.Sprintf("%d", hour)
	if len(days) > 7 {
		log.Error("The maximum amount of days allowed are 7")
		return "", fmt.Errorf("Amount of days should not be greater than 7")
	}
	if len(days) == 7 {
		// Set the days to '*' when all days are included in the slice
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

// Initializes the scheduler
func Init() error {
	// Creates the scheduler
	scheduler = gocron.NewScheduler(time.Local)
	scheduler.TagsUnique()

	exprDesc, _ := cron.NewDescriptor()
	cronExpr, _ := generateCronExpression(15, 10, []Day{Saturday, Friday})
	desc, _ := exprDesc.ToDescription(cronExpr, cron.LocaleAll)
	fmt.Printf("The scheduler will run: %s\n", desc)

	scheduler.StartAsync()
	return nil
}
