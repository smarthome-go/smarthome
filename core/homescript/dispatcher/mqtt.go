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
	"github.com/smarthome-go/smarthome/core/homescript/dispatcher/types"
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
	// This is required so that re-subscriptions work.
	CallbackFn mqtt.MessageHandler
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

	if err := manager.init(); err != nil {
		return &manager, err
	}

	Manager = manager

	return &Manager, nil
}

func (m *MqttManager) setConfig(config database.MqttConfig) {
	m.Config = config
}

func (m *MqttManager) init() error {
	if !m.Config.Enabled {
		logger.Debugf("MQTT subsystem is disabled according to server config")
		return nil
	}

	logger.Debugf("Initializing MQTT subsystem for broker `%s@%s`...", m.Config.Username, m.Config.Host)

	opts := mqtt.NewClientOptions().
		AddBroker(types.MakeBrokerURI(m.Config.Host, m.Config.Port)).
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

	logger.Infof("Initialized MQTT subsystem for broker `%s@%s`", m.Config.Username, m.Config.Host)
	return nil
}

func (m *MqttManager) Reload() error {
	// Unsubsribe from old old topics to all topics that were previously subscribed to.
	m.Subscriptions.Lock.Lock()
	defer m.Subscriptions.Lock.Unlock()

	for topicName := range m.Subscriptions.Set {
		if err := m.unsubscribeWithoutTracing(topicName); err != nil {
			logger.Warn("Failed to unsubscribe from old topic", err.Error())
		}
	}

	// Purge old connection.
	if m.Client != nil {
		m.Client.Disconnect(MqttDisconnectTimeoutMillis)
	}

	// Apply new config and connect.
	if err := m.init(); err != nil {
		logger.Errorf("Failed to reload MQTT manager: %s", err.Error())
		return err
	}

	if m.Config.Enabled {
		for topicName, subscription := range m.Subscriptions.Set {
			if err := m.subscribeWithoutTracing([]string{topicName}, subscription.CallbackFn); err != nil {
				logger.Errorf("Failed to reload MQTT manager: %s", err.Error())
				return err
			}
		}
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

func (m *MqttManager) subscribeWithoutTracing(topics []string, callBack mqtt.MessageHandler) error {
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

	return nil
}

func (m *MqttManager) Subscribe(topics []string, callBack mqtt.MessageHandler) error {
	if err := m.subscribeWithoutTracing(topics, callBack); err != nil {
		return err
	}

	for _, topic := range topics {
		m.Subscriptions.Lock.Lock()
		old, exists := m.Subscriptions.Set[topic]
		if !exists {
			m.Subscriptions.Set[topic] = Subscription{
				Consumers:  1,
				CallbackFn: callBack,
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

func (m *MqttManager) unsubscribeWithoutTracing(topic string) error {
	if token := m.Client.Unsubscribe(topic); token.Wait() && token.Error() != nil {
		logger.Errorf("Could not unsubscribe from MQTT topic `%s`: %s", topic, token.Error())
		return token.Error()
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
		if err := m.unsubscribeWithoutTracing(topic); err != nil {
			return err
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
