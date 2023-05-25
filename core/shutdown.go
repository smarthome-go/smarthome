package core

import (
	"fmt"
	"sync"
	"time"

	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/event"
	"github.com/smarthome-go/smarthome/core/hardware"
	"github.com/smarthome-go/smarthome/core/homescript"
)

type shutdownJob struct {
	channel chan struct{}
	name    string
}

// Maximum time to wait until everyting is shutdown
const SHUTDOWN_TIMEOUT = time.Second * 20

// Maximum allowed runtime for each boot automations
const BOOT_AUTOMATION_MAX_RUNTIME = time.Second * 20

func waitForHomescripts(ch *chan struct{}) {
	// Signal that the HMS wait task is finished
	defer func() {
		*ch <- struct{}{}
	}()
	for len(homescript.HmsManager.GetJobList()) > 0 {
		time.Sleep(time.Millisecond * 500)

		hmsList := ""
		for idx, hms := range homescript.HmsManager.GetJobList() {
			if idx > 0 {
				hmsList += ", "
			}
			hmsList += "`" + hms.Executor.ScriptName + "`"
		}

		log.Trace(fmt.Sprintf("Waiting for %d Homescripts [%s] to finish execution...", len(homescript.HmsManager.GetJobList()), hmsList))
	}
}

func waitForPowerJobs(ch *chan struct{}) {
	// Signal that the power wait task is finished
	defer func() {
		*ch <- struct{}{}
	}()

	for hardware.GetPendingJobCount() > 0 {
		time.Sleep(time.Millisecond * 500)
		log.Trace(fmt.Sprintf("Waiting for %d power jobs to finish...", hardware.GetPendingJobCount()))
	}
}

func waitForJobsWithTimeout(tasks *[]shutdownJob, timeout time.Duration) error {
	start := time.Now()

	for len(*tasks) != 0 {
		if time.Since(start) > timeout {
			return fmt.Errorf("timeout of `%v` exceeded", timeout)
		}

		for taskIdx, job := range *tasks {
			select {
			case <-job.channel:
				// remove this channel from `tasks`
				copy((*tasks)[taskIdx:], (*tasks)[taskIdx+1:])
				(*tasks)[len(*tasks)-1] = shutdownJob{}
				*tasks = (*tasks)[:len(*tasks)-1]
			default:
				// Do nothing, wait
			}
		}
	}
	return nil
}

func RunBootAutomations(config database.ServerConfig) {
	if !config.AutomationEnabled {
		log.Debug("Not running boot automations, automation system disabled")
		return
	}

	automations, err := database.GetAutomations()
	if err != nil {
		log.Error("Could not run boot automations: ", err.Error())
	}

	for _, job := range automations {
		if job.Data.Trigger != database.TriggerOnBoot || !job.Data.Enabled {
			continue
		}

		go func(jobId uint) {
			maxRuntime := BOOT_AUTOMATION_MAX_RUNTIME
			homescript.AutomationRunnerFunc(jobId, homescript.AutomationContext{MaximumHMSRuntime: &maxRuntime})
		}(job.Id)
	}
}

func runShutdownAutomations(ch *chan struct{}, config database.ServerConfig) {
	// Signal that all shutdown automations have successfully completed
	defer func() {
		*ch <- struct{}{}
	}()

	if !config.AutomationEnabled {
		log.Debug("Not running shutdown automations, automation system disabled")
		return
	}

	automations, err := database.GetAutomations()
	if err != nil {
		log.Error("Could not run shutdown automations: ", err.Error())
	}

	var wg sync.WaitGroup

	for _, job := range automations {
		if job.Data.Trigger != database.TriggerOnShutdown || !job.Data.Enabled {
			continue
		}

		wg.Add(1)

		go func(jobId uint) {
			homescript.AutomationRunnerFunc(jobId, homescript.AutomationContext{})
			wg.Done()
		}(job.Id)
	}

	wg.Wait()
}

func Shutdown(config database.ServerConfig) error {
	var tasks = make([]shutdownJob, 0)
	var error error

	// Shutdown automations (it is not safe to do this concurrently with the background HMS jobs)
	if err := homescript.DeactivateAutomationSystem(config); err != nil {
		error = err
	}

	// Run any shutdown automations
	autCh := make(chan struct{})
	tasks = append(tasks, shutdownJob{
		channel: autCh,
		name:    "run shutdown automations",
	})
	go runShutdownAutomations(&autCh, config)

	// HMS jobs
	hmsCh := make(chan struct{})
	tasks = append(tasks, shutdownJob{
		channel: hmsCh,
		name:    "wait for Homescripts",
	})
	go waitForHomescripts(&hmsCh)

	// Power jobs
	pwrCh := make(chan struct{})
	tasks = append(tasks, shutdownJob{
		channel: pwrCh,
		name:    "wait for power jobs",
	})
	go waitForPowerJobs(&pwrCh)

	// Fait for all background jobs
	if err := waitForJobsWithTimeout(&tasks, SHUTDOWN_TIMEOUT); err != nil {
		if len(tasks) > 0 {
			jobText := ""
			for idx, job := range tasks {
				if idx > 0 {
					jobText += fmt.Sprintf(", `%s`", job.name)
				} else {
					jobText += fmt.Sprintf("`%s`", job.name)
				}
			}

			log.Info(fmt.Sprintf("Unfinished shutdown jobs: [%s]", jobText))
		}
		return err
	}

	log.Debug("All core background shutdown tasks have finished")
	event.Info("System Shutdown", "System shutdown completed")
	return error
}
