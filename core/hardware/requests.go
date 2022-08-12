package hardware

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/event"
)

type PowerRequest struct {
	Switch string `json:"switch"`
	Power  bool   `json:"power"`
}

// Checks if a node is online and updates the database entry accordingly
func checkNodeOnlineRequest(node database.HardwareNode) error {
	// Client has timeout of 5 seconds in order to keep wait time short
	client := http.Client{Timeout: time.Second * 5}

	// Build URL
	urlTemp, err := url.Parse(node.Url)
	if err != nil {
		log.Warn(fmt.Sprintf(
			"Can not check health of node: '%s' due to malformed base URL: %s",
			node.Name,
			err.Error(),
		))
		return err
	}
	urlTemp.Path = "/health"

	// Perform the health-check request
	res, err := client.Get(urlTemp.String())
	if err != nil {
		log.Error("Hardware node checking request failed: ", err.Error())
		return err
	}
	if res.StatusCode != 200 {
		log.Error("Hardware node checking request failed: non 200 status code")
		return errors.New("checking node failed: non 200 status code")
	}
	return nil
}

// Runs the check request and updated the database entry accordingly
func checkNodeOnline(node database.HardwareNode) error {
	if err := checkNodeOnlineRequest(node); err != nil {
		// `node.Online` is checked to be `true` because it is still the value before the healthcheck
		if node.Online {
			log.Warn(fmt.Sprintf("Node '%s' failed to respond and is now offline", node.Name))
			go event.Error("Node Offline",
				fmt.Sprintf("Node '%s' went offline. It is recommended to address this issue as soon as possible", node.Name),
			)
		}
		if errDB := database.SetNodeOnline(node.Url, false); errDB != nil {
			log.Error("Failed to update power state of node: ", errDB.Error())
			return errDB
		}
		return nil
	}
	// `!node.Online` is checked because it is still the value before the healthcheck
	if !node.Online {
		log.Info(fmt.Sprintf("Node '%s' is back online", node.Name))
		go event.Info("Node back online", fmt.Sprintf("Node '%s' is back online.", node.Name))
	}
	if errDB := database.SetNodeOnline(node.Url, true); errDB != nil {
		log.Error("Failed to update power state of node: ", errDB.Error())
		return errDB
	}
	return nil
}

// Dispatches a power job to a given hardware node
// Returns an error if the job fails to execute on the hardware
// However, the preferred method of communication is by using the API `SetPower()` this way, priorities and interrupts are scheduled automatically
// A check if  a node is online again can be still executed afterwards
func sendPowerRequest(node database.HardwareNode, switchName string, powerOn bool) error {
	// Create the request body for the request
	requestBody, err := json.Marshal(PowerRequest{
		Switch: switchName,
		Power:  powerOn,
	})
	if err != nil {
		log.Error("Could not encode node request body: ", err.Error())
		return err
	}

	// Creates a client with a timeout of 2 seconds
	// TODO: make timeout editable / decide to use better timeout
	client := http.Client{Timeout: time.Second * 2}

	// Build a URL using the node parameters
	urlTemp, err := url.Parse(node.Url)
	if err != nil {
		log.Warn(fmt.Sprintf("Can not send power request to node '%s' due to malformed URL: %s", node.Name, err.Error()))
	}
	urlTemp.Path = "/power"

	// Build the token query
	query := url.Values{}
	query.Add("token", node.Token)
	urlTemp.RawQuery = query.Encode()

	// Perform the request
	res, err := client.Post(
		urlTemp.String(),
		"application/json",
		bytes.NewBuffer(requestBody),
	)
	if err != nil {
		log.Error("Hardware node request failed: ", err.Error())
		return err
	}

	// Evaluate non-200 outcome
	if res.StatusCode != 200 {
		// TODO: check firmware version of the hardware nodes at startup / in the healthcheck
		switch res.StatusCode {
		case 400:
			log.Error(fmt.Sprintf("Power request to node '%s' failed with code '400/bad-request': smarthome has sent a request that the node could not process", node.Name))
		case 401:
			log.Error(fmt.Sprintf("Power request to node '%s' failed with code '401/unauthorized': token configuration is likely invalid", node.Name))
		case 422:
			log.Error(fmt.Sprintf("Power request to node '%s' failed with code 422/unprocessable-entity: the requested switch is not configured on the node", node.Name))
		case 423:
			log.Error(fmt.Sprintf("Power request to node '%s' failed with code 423/locked: node is currently in use by another service", node.Name))
		case 500:
			log.Error(fmt.Sprintf("Power request to node '%s' failed with code 500/internal-server-error: undefined error which could not be matched", node.Name))
		case 503:
			log.Error(fmt.Sprintf("Power request to node '%s' failed with code 503/service-unavailable: node is currently in maintenance mode", node.Name))
		default:
			log.Error(fmt.Sprintf("Power request to node '%s' failed with unknown status code: %s", node.Name, res.Status))
		}
		return errors.New("set power failed: non 200 status code")
	}
	return nil
}

// A wrapper function which calls `checkNodeOnline`
// Does not return errors, just prints them
// Is used for determining whether a node which was previously offline is back online
// Is used via a `go` call for lower latency
func checkNodeOnlineWrapper(node database.HardwareNode) {
	if err := checkNodeOnline(node); err != nil {
		log.Trace(fmt.Sprintf("Node: '%s' is still offline", node.Name))
	}
}

// More user-friendly API to directly address all hardware nodes
// However, the preferred method of communication is by using the API `ExecuteJob()` this way, priorities and interrupts are scheduled automatically
// This function is internally used by `ExecuteJob`
// Makes a database request at the beginning in order to obtain information about the available nodes
// Updates the power state in the database after the jobs have been sent to the hardware nodes
func setPowerOnAllNodes(switchName string, powerOn bool) error {
	var err error
	// Retrieves available hardware nodes from the database
	nodes, err := database.GetHardwareNodes()
	if err != nil {
		log.Error("Failed to process power request: could not get nodes from database: ", err.Error())
		return err
	}
	for _, node := range nodes {
		if !node.Online && node.Enabled {
			// Check if the node is back online
			// a goroutine is used in order to keep up a fast response time
			go checkNodeOnlineWrapper(node)

			log.Warn(fmt.Sprintf("Skipping node: '%s' because it is currently marked as offline", node.Name))
			continue
		}

		// If the node is not enabled, skip the request
		if !node.Enabled {
			log.Trace(fmt.Sprintf("Skipping power request to disabled node '%s'", node.Name))
			continue
		}

		// Perform node request
		errTemp := sendPowerRequest(node, switchName, powerOn)
		if errTemp != nil {
			event.Error("Node Request Failed", fmt.Sprintf("Power request to node '%s' failed: %s", node.Name, errTemp.Error()))

			// If the request failed, check the node and mark it as offline
			if err := checkNodeOnline(node); err != nil {
				log.Error("Failed to check node online: ", err.Error())
			}
			err = errTemp
		} else {
			// If the node was previously offline and is now online, run a healthcheck to update its state
			// Log the event and mark the node as `online`
			if !node.Online {
				if err := checkNodeOnline(node); err != nil {
					log.Error("Failed to check node online: ", err.Error())
					return err
				}
			}
			log.Debug("Successfully dispatched power request to: ", node.Name)
		}
	}

	// Update the switch power-state in the database
	if _, err := database.SetPowerState(switchName, powerOn); err != nil {
		log.Error("Failed to set power after dispatching to all nodes: updating power-state database entry failed: ", err.Error())
		return err
	}
	return err
}

// Runs a health-check on all nodes of the system
// Used in system-level healthcheck
func RunNodeCheck() error {
	log.Debug("Running hardware node health check...")
	nodes, err := database.GetHardwareNodes()
	if err != nil {
		log.Error("Failed to check all hardware nodes: ", err.Error())
	}
	for _, node := range nodes {
		if err := checkNodeOnline(node); err != nil {
			log.Error(fmt.Sprintf("Healthcheck of node '%s' failed: %s", node.Name, err.Error()))
			return nil
		}
	}
	log.Debug("Hardware node healtheck finished")
	return nil
}
