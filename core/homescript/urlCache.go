package homescript

import (
	"net/url"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/smarthome-go/smarthome/core/database"
)

// Sets up a scheduler which triggers the flushing of the HMS URL cache every 12 hours
func StartUrlCacheGC() error {
	scheduler := gocron.NewScheduler(time.Local)
	if _, err := scheduler.Every(12).Hours().Do(flushUrlCacheScheduleRunner); err != nil {
		return err
	}
	scheduler.StartAsync()
	log.Debug("Successfully started Homescript URL cache GC")
	return nil
}

// Runner function used in `StartUrlCacheGC` which handles errors through logging
// => Target function executed by the schedule
func flushUrlCacheScheduleRunner() {
	log.Trace("Flushing Homescript URL cache records which are older than 12 hours...")
	if err := database.FlushHomescriptUrlCache(); err != nil {
		log.Error("Failed to flush Homescript URL cache records: ", err.Error())
	}
	log.Debug("Successfully flushed Homescript URL cache records which are older than 12 hours")
}

// Adds or updates an item in the URL cache
func insertCacheEntry(urlToInsert url.URL) error {
	return database.AddHomescriptUrlCacheEntry(urlToInsert.String())
}
