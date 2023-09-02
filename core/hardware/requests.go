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
			"Can not check health of node: '%s' due to malformed base URL: `%s`",
			node.Name,
			err.Error(),
		))
		return err
	}
	urlTemp.Path = "/health"

	// New in node v.0.4.0: healthcheck requires authentication
	query := url.Values{}
	query.Add("token", node.Token)
	urlTemp.RawQuery = query.Encode()

	// Perform the health-check request
	res, err := client.Get(urlTemp.String())
	if err != nil {
		log.Error("Hardware node checking request failed: ", err.Error())
		return err
	}
	if res.StatusCode != 200 {
		log.Error(fmt.Sprintf("Hardware node checking request failed: received status code %d (%s)", res.StatusCode, res.Status))
		return fmt.Errorf("Hardware node checking request failed: received status code %d (%s)", res.StatusCode, res.Status)
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

func setPowerOnNodesWrapper(switchItem database.Switch, powerOn bool) error {
	if switchItem.TargetNode != nil {
		hardwareNode, found, err := database.GetHardwareNodeByUrl(*switchItem.TargetNode)
		if err != nil {
			log.Error(fmt.Sprintf("Hardware node `%s` could not be retrieved from the database", *switchItem.TargetNode))
			return err
		}

		if !found {
			errMsg := fmt.Sprintf("Hardware node `%s` of switch `%s` does not exist but is referenced", *switchItem.TargetNode, switchItem.Id)
			log.Error(errMsg)
			return errors.New(errMsg)
		}

		actionPerformed, err := setPowerOnNode(hardwareNode, powerOn, switchItem)
		if err != nil {
			return err
		}

		if !actionPerformed {
			log.Warn("No power action performed since switch target node is disabled or offline")
		}

		// Update the switch power-state in the database
		if _, err := database.SetPowerState(switchItem.Id, powerOn); err != nil {
			log.Error("Failed to set power after dispatching to all nodes: updating power-state database entry failed: ", err.Error())
			return err
		}

		return nil
	} else {
		return setPowerOnAllNodes(switchItem, powerOn, false)
	}
}

func setPowerOnNode(node database.HardwareNode, powerOn bool, switchItem database.Switch) (performedAction bool, err error) {
	if !node.Online && node.Enabled {
		// Check if the node is back online
		// a goroutine is used in order to keep up a fast response time
		go checkNodeOnlineWrapper(node)

		log.Warn(fmt.Sprintf("Skipping node: '%s' because it is currently marked as offline", node.Name))
		return false, nil
	}

	// If the node is not enabled, skip the request
	if !node.Enabled {
		log.Trace(fmt.Sprintf("Skipping power request to disabled node '%s'", node.Name))
		return false, nil
	}

	// Perform node request
	errTemp := sendPowerRequest(node, switchItem.Id, powerOn)
	if errTemp != nil {
		event.Error("Node Request Failed", fmt.Sprintf("Power request to node '%s' failed: %s", node.Name, errTemp.Error()))

		// If the request failed, check the node and mark it as offline
		if err := checkNodeOnline(node); err != nil {
			log.Error("Failed to check node online: ", err.Error())
		}
		return false, errTemp
	} else {
		// If the node was previously offline and is now online, run a healthcheck to update its state
		// Log the event and mark the node as `online`
		if !node.Online {
			if err := checkNodeOnline(node); err != nil {
				log.Error("Failed to check node online: ", err.Error())
				return false, err
			}
		}
		log.Debug("Successfully dispatched power request to: ", node.Name)
	}

	return true, nil
}

// More user-friendly API to directly address all hardware nodes
// However, the preferred method of communication is by using the API `ExecuteJob()` this way, priorities and interrupts are scheduled automatically
// This function is internally used by `ExecuteJob`
// Makes a database request at the beginning in order to obtain information about the available nodes
// Updates the power state in the database after the jobs have been sent to the hardware nodes
func setPowerOnAllNodes(switchItem database.Switch, powerOn bool, forcePowerSet bool) error {
	var err error
	// Retrieves available hardware nodes from the database
	nodes, err := database.GetHardwareNodes()
	if err != nil {
		log.Error("Failed to process power request: could not get nodes from database: ", err.Error())
		return err
	}

	if len(nodes) == 0 {
		msg := "There are no hardware nodes, power state unaffected"
		log.Warn(msg)
		return fmt.Errorf(msg)
	}

	nodesFailed := 0
	nodesPerformed := 0

	for _, node := range nodes {
		actionPerformed, err := setPowerOnNode(node, powerOn, switchItem)
		if err != nil {
			nodesFailed++
			continue
		}
		if actionPerformed {
			nodesPerformed++
		}
	}

	if forcePowerSet {
		// Update the switch power-state in the database
		if _, err := database.SetPowerState(switchItem.Id, powerOn); err != nil {
			log.Error("Failed to set power after dispatching to all nodes: updating power-state database entry failed: ", err.Error())
			return err
		}
	}

	// If there was an error, return early
	if err != nil {
		return err
	}

	// There are no nodes or all nodes are disabled
	if nodesPerformed == 0 {
		log.Warn("There are no nodes or all nodes are disabled, no power action performed")
		return fmt.Errorf("There are no nodes or all nodes are disabled, no action performed")
	}

	// All nodes failed, should not update the power state
	if nodesFailed == len(nodes) {
		return fmt.Errorf("All nodes failed: `%s`", err.Error())
	}

	// Update the switch power-state in the database
	if !forcePowerSet {
		if _, err := database.SetPowerState(switchItem.Id, powerOn); err != nil {
			log.Error("Failed to set power after dispatching to all nodes: updating power-state database entry failed: ", err.Error())
			return err
		}
	}

	return nil
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
