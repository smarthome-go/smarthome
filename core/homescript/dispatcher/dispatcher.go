package dispatcher

import (
	"fmt"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/sirupsen/logrus"
	"github.com/smarthome-go/smarthome/core/homescript/types"
)

var logger *logrus.Logger

func InitLogger(log *logrus.Logger) {
	logger = log
}

//
// Dispatcher implementation.
//

type InstanceT struct {
	Hms               types.Manager
	Mqtt              MqttManager
	Registrations     []RegisterInfo
	MqttRegistrations map[string][]RegisterInfo
}

var Instance InstanceT

func InitInstance(hms types.Manager, mqtt MqttManager) {
	Instance = InstanceT{
		Hms:  hms,
		Mqtt: mqtt,
	}
}

func (i *InstanceT) Register(info RegisterInfo) error {
	switch trigger := info.Trigger.(type) {
	case CallBackTriggerMqtt:
		// TODO: maybe check that a program cannot register twice.
		i.Mqtt.Subscribe(trigger.Topics, i.mqttCallBack)
	default:
		panic(fmt.Sprintf("Unreachable: introduced a new trigger type (%v) without updating this code", info.Trigger))
	}

	return nil
}

func (i *InstanceT) CallBack(info RegisterInfo, meta any) {
	panic("TODO")
}

type MqttMessage struct {
	Topic   string
	Payload string
}

func (i *InstanceT) mqttCallBack(_ mqtt.Client, message mqtt.Message) {
	// Invoke all MQTT registrations for this topic.
	topic := message.Topic()
	payload := string(message.Payload())

	logger.Tracef("Mqtt Callback: topic: %s, payload: %s", topic, payload)

	for _, registration := range i.MqttRegistrations[topic] {
		i.CallBack(registration, MqttMessage{
			Topic:   topic,
			Payload: payload,
		})
	}
}
