package dispatcher

import (
	"fmt"
	"log"
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/sirupsen/logrus"
)

const MqttKeepAlive time.Duration = time.Second * 60
const MqttPingTimeout time.Duration = time.Second
const MqttDisconnectTimeoutMillis uint = 250
const MqttQOS byte = 0x0

type MqttConfig struct {
	Host     string
	Username string
	Password string
}

type MqttManager struct {
	Subscriptions map[string]mqtt.MessageHandler
	Client        mqtt.Client
	Config        MqttConfig
}

func (m MqttManager) messageHandler(_ mqtt.Client, msg mqtt.Message) {
	fmt.Printf("TOPIC: %s\n", msg.Topic())
	fmt.Printf("MSG: %s\n", msg.Payload())
}

func (m *MqttManager) Init() error {
	opts := mqtt.NewClientOptions().
		AddBroker(m.Config.Host).
		SetClientID("homescript-smarthome").
		SetUsername(m.Config.Username).
		SetPassword(m.Config.Password)

	opts.SetKeepAlive(MqttKeepAlive)
	opts.SetPingTimeout(MqttPingTimeout)
	opts.SetDefaultPublishHandler(m.messageHandler)

	c := mqtt.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	m.Client = c

	return nil
}

func (m *MqttManager) Shutdown() error {
	for topic := range m.Subscriptions {
		if token := m.Client.Unsubscribe(topic); token.Wait() && token.Error() != nil {
			return token.Error()
		}
	}

	m.Client.Disconnect(MqttDisconnectTimeoutMillis)
	return nil
}

func (m *MqttManager) Subscribe(topics []string, callBack mqtt.MessageHandler) {
	if logger.GetLevel() <= logrus.TraceLevel {
		mqtt.DEBUG = log.New(os.Stderr, "", 0)
		mqtt.ERROR = log.New(os.Stderr, "", 0)
	}

	topicsArgs := make(map[string]byte)
	for _, topic := range topics {
		topicsArgs[topic] = MqttQOS
	}

	if token := m.Client.SubscribeMultiple(topicsArgs, callBack); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}
}

func (m *MqttManager) Publish(topic string, message string) error {
	token := m.Client.Publish(topic, MqttQOS, false, message)
	token.Wait()
	return token.Error()
}
