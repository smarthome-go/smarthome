package hardware

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
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
	// Client has timeout of a second too
	client := http.Client{Timeout: time.Second}
	res, err := client.Get(fmt.Sprintf("%s/health", node.Url))
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
		if node.Online {
			log.Warn(fmt.Sprintf("Node `%s` failed to respond and is now offline", node.Name))
			go event.Error("Node Offline",
				fmt.Sprintf("Node %s went offline. Users will have to deal with increased wait times. It is advised to address this issue as soon as possible", node.Name))
		}
		if errDB := database.SetNodeOnline(node.Url, false); errDB != nil {
			log.Error("Failed to update power state of node: ", errDB.Error())
			return errDB
		}
		return nil
	}
	if !node.Online {
		log.Info(fmt.Sprintf("Node `%s` is now back online", node.Name))
		go event.Info("Node Online", fmt.Sprintf("Node %s is back online.", node.Name))
	}
	if errDB := database.SetNodeOnline(node.Url, true); errDB != nil {
		log.Error("Failed to update power state of node: ", errDB.Error())
		return errDB
	}
	return nil
}

// Delivers a power job to a given hardware node
// Returns an error if the job fails to execute on the hardware
// However, the preferred method of communication is by using the API `SetPower()` this way, priorities and interrupts are scheduled automatically
// A check if  a node is online again can be still executed afterwards
func sendPowerRequest(node database.HardwareNode, switchName string, powerOn bool) error {
	if !node.Enabled {
		log.Trace("Skipping power request to disabled node")
		return nil
	}
	requestBody, err := json.Marshal(PowerRequest{
		Switch: switchName,
		Power:  powerOn,
	})
	if err != nil {
		log.Error("Could not parse node request: ", err.Error())
		return err
	}
	// Create a client with a more realistic timeout of 1 second
	client := http.Client{Timeout: time.Second}
	res, err := client.Post(fmt.Sprintf("%s/power?token=%s", node.Url, node.Token), "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		log.Error("Hardware node request failed: ", err.Error())
		return err
	}
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
	defer res.Body.Close()
	return nil
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
			if errTemp := checkNodeOnline(node); errTemp != nil {
				log.Debug(fmt.Sprintf("Node %s is still offline", node.Name))
			}
			log.Warn(fmt.Sprintf("Skipping node: '%s' because it is currently marked as offline", node.Name))
			continue
		}
		errTemp := sendPowerRequest(node, switchName, powerOn)
		if errTemp != nil {
			// Log the error
			event.Error("Node Request Failed", fmt.Sprintf("Power request to node '%s' failed: %s", node.Name, errTemp.Error()))
			// If the request failed, check the node and mark it as offline
			if err := checkNodeOnline(node); err != nil {
				log.Error("Failed to check node online: ", err.Error())
			}
			err = errTemp
		} else {
			if !node.Online {
				// If the node was previously offline and is now online
				if err := checkNodeOnline(node); err != nil {
					log.Error("Failed to check node online: ", err.Error())
				}
			}
			log.Debug("Successfully sent power request to: ", node.Name)
		}
	}
	if _, err := database.SetPowerState(switchName, powerOn); err != nil {
		log.Error("Failed to set power after addressing all nodes: updating database entry failed: ", err.Error())
		return err
	}
	return err
}

// Runs a health-check on all nodes of the system
// Used in system-level healthcheck
func RunNodeCheck() error {
	nodes, err := database.GetHardwareNodes()
	if err != nil {
		log.Error("Failed to check nodes: ", err.Error())
	}
	for _, node := range nodes {
		if err := checkNodeOnline(node); err != nil {
			log.Error("Failed to check node: checkNodeOnline failed: ", err.Error())
			return nil
		}
	}
	return nil
}
