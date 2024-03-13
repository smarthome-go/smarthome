package dispatcher

import (
	"errors"
	"log"
	"os"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/sirupsen/logrus"
	"github.com/smarthome-go/smarthome/core/database"
)

const MqttKeepAlive time.Duration = time.Second * 60
const MqttPingTimeout time.Duration = time.Second
const MqttDisconnectTimeoutMillis uint = 250
const MqttQOS byte = 0x0

// Error messages.

const notInitializedErrMsg = "MQTT subsystem is not initialized"

// TODO: actually track subscriptions

type Subscription struct {
	Consumers uint
}

type Subscriptions struct {
	Set  map[string]Subscription
	Lock sync.RWMutex
}

type MqttManager struct {
	Subscriptions Subscriptions
	Client        mqtt.Client
	Config        database.MqttConfig
	Initialized   bool
}

var Manager MqttManager

func (m *MqttManager) messageHandler(_ mqtt.Client, msg mqtt.Message) {
	panic("Unreachable: fallback on default message handler")
}

func NewMqttManager(config database.MqttConfig) (m *MqttManager, e error) {
	manager := MqttManager{
		Subscriptions: Subscriptions{
			Set:  make(map[string]Subscription),
			Lock: sync.RWMutex{},
		},
		Client:      nil,
		Config:      config,
		Initialized: false,
	}

	logger.Debugf("Intializing MQTT subsystem for broker `%s@%s`...", manager.Config.Username, manager.Config.Host)

	if err := manager.init(); err != nil {
		return &manager, err
	}

	logger.Infof("Initialized MQTT subsystem for broker `%s@%s`", manager.Config.Username, manager.Config.Host)

	Manager = manager

	return &Manager, nil
}

func (m *MqttManager) init() error {
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
	m.Initialized = true

	if logger.GetLevel() <= logrus.TraceLevel {
		mqtt.DEBUG = log.New(os.Stderr, "", 0)
		mqtt.ERROR = log.New(os.Stderr, "", 0)
	}

	return nil
}

func (m *MqttManager) Shutdown() error {
	if !m.Initialized {
		return nil
	}

	for topic := range m.Subscriptions.Set {
		if token := m.Client.Unsubscribe(topic); token.Wait() && token.Error() != nil {
			return token.Error()
		}
	}

	m.Client.Disconnect(MqttDisconnectTimeoutMillis)
	return nil
}

func (m *MqttManager) Subscribe(topics []string, callBack mqtt.MessageHandler) error {
	if !m.Initialized {
		return errors.New(notInitializedErrMsg)
	}

	topicsArgs := make(map[string]byte)
	for _, topic := range topics {
		topicsArgs[topic] = MqttQOS
	}

	if token := m.Client.SubscribeMultiple(topicsArgs, callBack); token.Wait() && token.Error() != nil {
		logger.Errorf("Could not subscribe to MQTT topics `%v`: %s", topics, token.Error())
		return token.Error()
	}

	for _, topic := range topics {
		m.Subscriptions.Lock.Lock()
		old, exists := m.Subscriptions.Set[topic]
		if !exists {
			m.Subscriptions.Set[topic] = Subscription{
				Consumers: 1,
			}
			logger.Tracef("Created new consumer tracking for topic `%s`", topic)
		} else {
			old.Consumers++
			m.Subscriptions.Set[topic] = old
			logger.Tracef("New consumer for topic `%s`, new count: %d", topic, old.Consumers)
		}
		m.Subscriptions.Lock.Unlock()
	}

	return nil
}

func (m *MqttManager) Unsubscribe(topic string) error {
	if !m.Initialized {
		return errors.New(notInitializedErrMsg)
	}

	m.Subscriptions.Lock.Lock()
	defer m.Subscriptions.Lock.Unlock()

	old, exists := m.Subscriptions.Set[topic]
	if !exists {
		panic("Cannot unsubscribe from a topic without subscriptions")
	}
	if old.Consumers == 0 {
		panic("Cannot unsubscribe from a topic with 0 consumers")
	}

	old.Consumers--

	logger.Tracef("Deleted one subscription of topic `%s`, new consumer count: %d", topic, old.Consumers)

	if old.Consumers == 0 {
		if token := m.Client.Unsubscribe(topic); token.Wait() && token.Error() != nil {
			logger.Errorf("Could not unsubsribe from MQTT topic `%s`: %s", topic, token.Error())
			return token.Error()
		}
		delete(m.Subscriptions.Set, topic)
		logger.Tracef("Deleted all subscriptions of topic `%s`", topic)
	} else {
		m.Subscriptions.Set[topic] = old
	}

	return nil
}

func (m *MqttManager) Publish(topic string, message string) error {
	if !m.Initialized {
		return errors.New(notInitializedErrMsg)
	}

	token := m.Client.Publish(topic, MqttQOS, false, message)
	token.Wait()
	if token.Error() != nil {
		logger.Errorf("Could not publish to MQTT topic `%s`: %s", topic, token.Error())
	}
	return token.Error()
}
