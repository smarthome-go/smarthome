package hardware

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/MikMuellerDev/smarthome/core/database"
)

type HardwareRequest struct {
	Switch string `json:"switch"`
	Power  bool   `json:"power"`
}

// Delivers a power job to a given hardware node
// Returns an error if the job fails to execute on the hardware
// However, the preferred method of communication is by using the API `SetPower()` this way, priorities and interrupts are scheduled automatically
func sendPowerRequest(node database.HardwareNode, switchName string, powerOn bool) error {
	requestBody, err := json.Marshal(HardwareRequest{
		Switch: switchName,
		Power:  powerOn,
	})
	if err != nil {
		log.Error("Could not parse node request: ", err.Error())
		return err
	}
	res, err := http.Post(fmt.Sprintf("%s/power?token=%s", node.Url, node.Token), "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		log.Error("Hardware node request failed: ", err.Error())
		return err
	}
	if res.StatusCode != 200 {
		// TODO: check firmware version of the power nodes, analyze errors
		log.Error(fmt.Sprintf("Received non 200 status code: %d", res.StatusCode))
		return errors.New("received non 200 status code while sending request to hardware node")
	}
	defer res.Body.Close()
	return nil
}

// More user-friendly API to directly address all hardware nodes
// However, the preferred method of communication is by using the API `ExecuteJob()` this way, priorities and interrupts are scheduled automatically
// This method is internally used by `ExecuteJob`
// Makes a database request at the beginning in order to obtain information about the available nodes
// Updates the power state in the database after the jobs have been sent to the hardware nodes
func setPowerOnAllNodes(switchName string, powerOn bool) error {
	var err error = nil
	// Retrieves available hardware nodes from the database
	nodes, err := database.GetHardwareNodes()
	if err != nil {
		log.Error("Failed to process power request: could not get nodes from database: ", err.Error())
		return err
	}
	for _, node := range nodes {
		errTemp := sendPowerRequest(node, switchName, powerOn)
		if errTemp != nil {
			err = errTemp
		} else {
			log.Debug("Successfully sent power request to: ", node.Name)
		}
	}
	if _, err := database.SetPowerState(switchName, powerOn); err != nil {
		log.Error("Failed to set power after addressing all nodes: updating database entry failed: ", err.Error())
		return err
	}
	return err
}
