package dispatcher

import (
	"errors"
	"fmt"
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
const MqttQOS byte = 0x2
const MqttHealthCheckTopic = "healthcheck"

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

type MqttManagerBody struct {
	Subscriptions Subscriptions
	Client        mqtt.Client
	Config        database.MqttConfig
	Initialized   bool
}

type MqttManager struct {
	// Lock sync.Mutex
	Body MqttManagerBody

	// This is set by the parent (the dispatcher) to be triggered once
	// there is an event worthy of triggering all pending registrations to be retried.
	TriggerTryPendingRegistrations func() error
}

var Manager MqttManager

func (m *MqttManager) messageHandler(_ mqtt.Client, msg mqtt.Message) {
	panic("Unreachable: fallback on default message handler")
}

func (m *MqttManager) connectionLostHandler(_ mqtt.Client, err error) {
	m.Body.Initialized = false
	logger.Errorf("MQTT connection lost: %s\n", err.Error())
}

func (m *MqttManager) connectionEstablishedHandler(client mqtt.Client) {
	m.Body.Client = client
	m.Body.Initialized = true

	if err := m.reloadOnReconnect(); err != nil {
		logger.Errorf("Failed to reload MQTT dispatcher after connection was established: %s\n", err.Error())
	}

	if err := m.TriggerTryPendingRegistrations(); err != nil {
		logger.Errorf("Failed to trigger parent reload after connection was established: %s\n", err.Error())
	}

	logger.Debug("MQTT connection established")
}

func InitModule() {
	if logger.GetLevel() <= logrus.TraceLevel {
		mqtt.DEBUG = log.New(os.Stderr, "", 0)
		mqtt.ERROR = log.New(os.Stderr, "", 0)
	}
}

func NewMqttManager(config database.MqttConfig, retryHook func() error) (m *MqttManager, e error) {
	Manager = MqttManager{
		Body: MqttManagerBody{
			Subscriptions: Subscriptions{
				Set:  make(map[string]Subscription),
				Lock: sync.RWMutex{},
			},
			Client:      nil,
			Config:      config,
			Initialized: false,
		},
		TriggerTryPendingRegistrations: retryHook,
	}

	if err := Manager.init(); err != nil {
		return &Manager, err
	}

	return &Manager, nil
}

func (m *MqttManager) setConfig(config database.MqttConfig) {
	m.Body.Config = config
}

func (m *MqttManager) init() error {
	if !m.Body.Config.Enabled {
		logger.Debugf("MQTT subsystem is disabled according to server config")
		return nil
	}

	logger.Debugf("Initializing MQTT subsystem for broker `%s@%s`...", m.Body.Config.Username, m.Body.Config.Host)

	opts := mqtt.NewClientOptions().
		AddBroker(types.MakeBrokerURI(m.Body.Config.Host, m.Body.Config.Port)).
		SetClientID("homescript-smarthome").
		SetUsername(m.Body.Config.Username).
		SetPassword(m.Body.Config.Password)

	opts.SetKeepAlive(MqttKeepAlive)
	opts.SetPingTimeout(MqttPingTimeout)
	opts.SetDefaultPublishHandler(m.messageHandler)
	opts.SetAutoReconnect(true)
	opts.SetConnectionLostHandler(m.connectionLostHandler)
	opts.SetOnConnectHandler(m.connectionEstablishedHandler)

	client := mqtt.NewClient(opts)

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	start := time.Now()

	// TODO: configure timeout here
	for !client.IsConnected() && time.Since(start) < time.Second*10 {
		logger.Trace("[MQTT manager] Waiting for connection...")
		time.Sleep(time.Second)
	}

	if token := client.Publish(MqttHealthCheckTopic, MqttQOS, false, ""); token.Error() != nil {
		return token.Error()
	}

	m.Body.Initialized = true

	logger.Infof("Initialized MQTT subsystem for broker `%s@%s`", m.Body.Config.Username, m.Body.Config.Host)

	return nil
}

func (m *MqttManager) IsConnected() bool {
	return m.Body.Initialized && m.Body.Client != nil && m.Body.Client.IsConnected()
}

func (m *MqttManager) Status() error {
	if !m.IsConnected() {
		if err := m.init(); err != nil {
			return err
		}

		if m.Body.Client == nil || !m.Body.Client.IsConnected() {
			return fmt.Errorf("Not connected to broker")
		}
	}

	if token := m.Body.Client.Publish(MqttHealthCheckTopic, MqttQOS, false, ""); token.Error() != nil {
		return token.Error()
	}

	return nil
}

func (m *MqttManager) Reload() error {
	m.unsubscribeAllSubscriptionsNonTracing()

	// Purge old connection.
	if m.Body.Client != nil {
		m.Body.Client.Disconnect(MqttDisconnectTimeoutMillis)
	}

	// Apply new config and connect.
	if err := m.init(); err != nil {
		logger.Errorf("Failed to reload MQTT manager: %s", err.Error())
		return err
	}

	enabled := m.Body.Config.Enabled

	if enabled {
		if err := m.resubscribeNonTracing(); err != nil {
			return err
		}
	}

	return nil
}

func (m *MqttManager) reloadOnReconnect() error {
	if !m.Body.Config.Enabled {
		return nil
	}

	m.unsubscribeAllSubscriptionsNonTracing()
	return m.resubscribeNonTracing()
}

func (m *MqttManager) unsubscribeAllSubscriptionsNonTracing() {
	// Unsubsribe from old old topics to all topics that were previously subscribed to.
	m.Body.Subscriptions.Lock.Lock()
	defer m.Body.Subscriptions.Lock.Unlock()

	for topicName := range m.Body.Subscriptions.Set {
		m.Body.Subscriptions.Lock.Unlock()
		err := m.unsubscribeWithoutTracing(topicName)
		m.Body.Subscriptions.Lock.Lock()
		if err != nil {
			logger.Warn("Failed to unsubscribe from old topic", err.Error())
		}
	}
}

func (m *MqttManager) resubscribeNonTracing() error {
	m.Body.Subscriptions.Lock.Lock()
	defer m.Body.Subscriptions.Lock.Unlock()

	for topicName, subscription := range m.Body.Subscriptions.Set {
		err := m.subscribeWithoutTracing([]string{topicName}, subscription.CallbackFn)
		if err != nil {
			logger.Errorf("Failed to reload MQTT manager: %s", err.Error())
			return err
		}
	}

	return nil
}

func (m *MqttManager) Shutdown() error {
	if !m.Body.Initialized {
		return nil
	}

	for topic := range m.Body.Subscriptions.Set {
		if token := m.Body.Client.Unsubscribe(topic); token.Wait() && token.Error() != nil {
			return token.Error()
		}
	}

	m.Body.Client.Disconnect(MqttDisconnectTimeoutMillis)
	return nil
}

func (m *MqttManager) subscribeWithoutTracing(topics []string, callBack mqtt.MessageHandler) error {
	if !m.Body.Initialized {
		return errors.New(notInitializedErrMsg)
	}

	topicsArgs := make(map[string]byte)
	for _, topic := range topics {
		topicsArgs[topic] = MqttQOS
	}

	if token := m.Body.Client.SubscribeMultiple(topicsArgs, callBack); token.Wait() && token.Error() != nil {
		logger.Errorf("Could not subscribe to MQTT topics `%v`: %s", topics, token.Error())
		return token.Error()
	}

	return nil
}

func (m *MqttManager) Subscribe(topics []string, callBack mqtt.MessageHandler) error {
	if err := m.subscribeWithoutTracing(topics, callBack); err != nil {
		return err
	}

	m.Body.Subscriptions.Lock.Lock()
	defer m.Body.Subscriptions.Lock.Unlock()

	for _, topic := range topics {
		old, exists := m.Body.Subscriptions.Set[topic]
		if !exists {
			m.Body.Subscriptions.Set[topic] = Subscription{
				Consumers:  1,
				CallbackFn: callBack,
			}
			logger.Tracef("Created new consumer tracking for topic `%s`", topic)
		} else {
			old.Consumers++
			m.Body.Subscriptions.Set[topic] = old
			logger.Tracef("New consumer for topic `%s`, new count: %d", topic, old.Consumers)
		}
	}

	return nil
}

func (m *MqttManager) unsubscribeWithoutTracing(topic string) error {
	if token := m.Body.Client.Unsubscribe(topic); token.Wait() && token.Error() != nil {
		logger.Errorf("Could not unsubscribe from MQTT topic `%s`: %s", topic, token.Error())
		return token.Error()
	}

	return nil
}

func (m *MqttManager) Unsubscribe(topic string) error {
	m.Body.Subscriptions.Lock.Lock()
	defer m.Body.Subscriptions.Lock.Unlock()

	if !m.Body.Initialized {
		return errors.New(notInitializedErrMsg)
	}

	old, exists := m.Body.Subscriptions.Set[topic]
	if !exists {
		logger.Errorf("Cannot unsubscribe from a topic (%s) without subscriptions", topic)
		return nil
	}
	if old.Consumers == 0 {
		panic("Cannot unsubscribe from a topic with 0 consumers")
	}

	old.Consumers--

	logger.Tracef("Deleted one subscription of topic `%s`, new consumer count: %d", topic, old.Consumers)

	if old.Consumers == 0 {
		err := m.unsubscribeWithoutTracing(topic)
		if err != nil {
			return err
		}

		delete(m.Body.Subscriptions.Set, topic)
		logger.Tracef("Deleted all subscriptions of topic `%s`", topic)
	} else {
		m.Body.Subscriptions.Set[topic] = old
	}

	return nil
}

func (m *MqttManager) Publish(topic string, message string) error {
	if !m.Body.Initialized {
		return errors.New(notInitializedErrMsg)
	}

	token := m.Body.Client.Publish(topic, MqttQOS, false, message)
	token.Wait()
	if token.Error() != nil {
		logger.Errorf("Could not publish to MQTT topic `%s`: %s", topic, token.Error())
	}
	return token.Error()
}
