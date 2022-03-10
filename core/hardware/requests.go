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
	Content string `json:"content"`
	Switch  string `json:"switch"`
	TurnOn  bool   `json:"turnOn"`
	Token   string `json:"token"`
}

var hwConfig HardwareConfig

func InitConfig(hwConf HardwareConfig) {
	hwConfig = hwConf
}

// Delivers a power job to a given hardware node
// Returns an error if the job fails to execute on the hardware
// However, the preferred method of communication is by using the API `SetPower()` this way, priorities and interrupts are scheduled automatically
func sendPowerRequest(node Node, switchName string, turnOn bool) error {
	// TODO: make hardware node software better, best would be non-python
	requestBody, err := json.Marshal(HardwareRequest{
		Content: "power",
		Switch:  switchName,
		TurnOn:  turnOn,
		Token:   node.Token,
	})
	if err != nil {
		log.Error("Could not parse node request: ", err.Error())
		return err
	}
	res, err := http.Post(node.Url, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		log.Error("Hardware node request failed: ", err.Error())
		return err
	}
	if res.StatusCode != 200 {
		log.Error(fmt.Sprintf("Received non 200 status code: %d", res.StatusCode))
		return errors.New("received non 200 status code while sending request to hardware node")
	}
	defer res.Body.Close()
	return nil
}

// More user-friendly API to directly address all hardware nodes
// However, the preferred method of communication is by using the API `ExecuteJob()` this way, priorities and interrupts are scheduled automatically
// This method is internally used by `ExecuteJob`
func setPowerOnAllNodes(switchName string, turnOn bool) error {
	var err error = nil
	for _, node := range hwConfig.Nodes {
		errTemp := sendPowerRequest(node, switchName, turnOn)
		if errTemp != nil {
			err = errTemp
		} else {
			log.Debug("Successfully sent power request to: ", node.Name)
		}
	}
	database.SetPowerState(switchName, turnOn)
	return err
}
