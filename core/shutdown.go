package core

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/smarthome-go/smarthome/core/automation"
	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/event"
	hardware "github.com/smarthome-go/smarthome/core/hardware_deprecated"
	"github.com/smarthome-go/smarthome/core/homescript"
	"github.com/smarthome-go/smarthome/core/homescript/dispatcher"
	"github.com/smarthome-go/smarthome/core/homescript/types"
)

type shutdownJobName string

const (
	shutdownJobHMS = "homescript"
)

type shutdownJob struct {
	channel chan struct{}
	name    shutdownJobName
}

// Maximum time to wait until everything is shutdown
const SHUTDOWN_TIMEOUT = time.Second * 20

// Maximum allowed runtime for automations
const BOOT_AUTOMATION_MAX_RUNTIME = time.Second * 20

func waitForHomescripts(ch *chan struct{}) {
	// Record the start time, if there are still scripts after the half of the shutdown timeout,
	// then kill all scripts.
	start := time.Now()

	// Signal that the HMS wait task is finished
	defer func() {
		*ch <- struct{}{}
	}()

	sentKill := false

	for len(homescript.HmsManager.GetJobList()) > 0 {
		time.Sleep(time.Millisecond * 500)

		jobLen := len(homescript.HmsManager.GetJobList())
		if !sentKill && (time.Since(start) >= (SHUTDOWN_TIMEOUT-homescript.KillEventMaxRuntime)/2) && jobLen > 0 {
			log.Infof("Killing remaining %d Homescripts...", jobLen)

			for _, job := range homescript.HmsManager.GetJobList() {
				go homescript.HmsManager.Kill(job.JobID)
			}

			sentKill = true
			continue
		}

		hmsList := ""
		for idx, hms := range homescript.HmsManager.GetJobList() {
			if idx > 0 {
				hmsList += ", "
			}
			hmsList += fmt.Sprintf("`%s`", hms.HmsID)
		}

		waitForWhatText := "finish execution"
		if sentKill {
			waitForWhatText = "respond to termination"
		}

		log.Trace(fmt.Sprintf("Waiting for %d Homescripts [%s] to %s...", len(homescript.HmsManager.GetJobList()), hmsList, waitForWhatText))
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
			automation.AutomationRunnerFunc(
				jobId,
				types.NewExecutionContextAutomation(
					types.NewExecutionContextUser(
						job.Data.HomescriptId,
						job.Owner,
						nil,
					),
					types.ExecutionContextAutomationInner{
						NotificationContext: nil,
						MaximumHMSRuntime:   &maxRuntime,
					},
				),
			)
		}(job.Id)
	}
}

func runShutdownAutomations(ch *chan struct{}, config database.ServerConfig) {
	// Signal that all shutdown automations have successfully completed.
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
			automation.AutomationRunnerFunc(
				jobId,
				types.NewExecutionContextAutomation(
					types.NewExecutionContextUser(
						job.Data.HomescriptId,
						job.Owner,
						nil,
					),
					types.ExecutionContextAutomationInner{
						NotificationContext: nil,
						MaximumHMSRuntime:   nil,
					},
				),
			)
			wg.Done()
		}(job.Id)
	}

	wg.Wait()
}

func Shutdown(terminateProcess bool) error {
	log.Info("System shutting down...")

	config, found, err := database.GetServerConfiguration()
	if err != nil {
		return err
	}

	if !found {
		return errors.New("Could not shutdown: not server configuration found")
	}

	return ShutdownWithConfig(config, terminateProcess)
}

func shutdownMQTT() {
	log.Debug("Initiating MQTT shutdown...")
	dispatcher.Manager.ShutdownChan <- struct{}{}

	timeout := time.Second * 10
	start := time.Now()

outer:
	for {
		select {
		case <-dispatcher.Manager.ShutdownCompleted:
			break outer
		default:
			if time.Since(start) > timeout {
				log.Warn("MQTT shutdown timeout exceeded")
				break outer
			}
		}
	}

	log.Debug("MQTT shutdown complete")
}

func ShutdownWithConfig(config database.ServerConfig, terminateProcess bool) error {
	var tasks = make([]shutdownJob, 0)
	var error error

	if err := database.SetLockDownModeEnabled(false); err != nil {
		return err
	}

	// Shutdown MQTT keepalive.
	// BUG: this destroys everything.
	shutdownMQTT()

	// Shutdown automations (it is not safe to do this concurrently with the background HMS jobs)
	if err := automation.Manager.DeactivateAutomationSystem(config); err != nil {
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
	if error != nil {
		return error
	}

	event.Info("System Shutdown", "System shutdown completed")
	log.Info("Shutdown completed")

	if err := database.SetLockDownModeEnabled(config.LockDownMode); err != nil {
		return err
	}

	if terminateProcess {
		os.Exit(0)
	}

	return nil
}
