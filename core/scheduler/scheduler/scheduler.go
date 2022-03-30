package scheduler

import (
	"time"

	"github.com/go-co-op/gocron"
	"github.com/sirupsen/logrus"
)

// This scheduler is executed only once, then disabled the job it should run
var scheduler *gocron.Scheduler

var log *logrus.Logger

func InitLogger(logger *logrus.Logger) {
	log = logger
}

// - [ ] Add function Receives a homescript string as an input
// - [ ] The homescript string is generated by the API layer
// - [ ] Store jobs in the database
// - [ ] Start jobs which have not been executed from the database
// - [ ] Make a scheduler runner
// - [ ] Make already set up scheduler editable
// - [ ] Add database schema
func Init() error {
	scheduler = gocron.NewScheduler(time.Local)
	scheduler.TagsUnique()
	scheduler.StartAsync()
	return nil
}