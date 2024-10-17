package dispatcher

import (
	"errors"
	"fmt"
	"log"
	"os"
	"runtime"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/sirupsen/logrus"
	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/homescript/dispatcher/types"
)

const MqttKeepAlive time.Duration = time.Second * 60
const MqttConnectTimeout time.Duration = time.Second * 3
const MqttPingTimeout time.Duration = time.Second
const MqttDisconnectTimeoutMillis uint = 250
const MqttQOS byte = 0x2
const MqttHealthCheckTopic = "healthcheck"
const mqttAllowVerbose = false

// Error messages.

const notInitializedErrMsg = "MQTT subsystem is not initialized"

// TODO: actually track subscriptions

type Subscription struct {
	Consumers uint
	// This is required so that re-subscriptions work.
	CallbackFn mqtt.MessageHandler
}

type MqttManagerBody struct {
	Subscriptions map[string]Subscription
	Client        mqtt.Client
	Config        database.MqttConfig
	Initialized   bool
}

//
// BEGIN TRACING MUTEX.
//

const runtimeCallerFailedMsg = "Getting caller function at runtime failed"

type TracingRWMutex struct {
	Internal *sync.RWMutex
}

func (m TracingRWMutex) Lock() {
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		panic(runtimeCallerFailedMsg)
	}
	logger.Tracef("Called `Lock()` from %s:%d", file, line)
	m.Internal.Lock()
}

func (m TracingRWMutex) Unlock() {
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		panic(runtimeCallerFailedMsg)
	}
	logger.Tracef("Called `Unlock()` from %s:%d", file, line)
	m.Internal.Unlock()
}

func (m TracingRWMutex) RLock() {
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		panic(runtimeCallerFailedMsg)
	}
	logger.Tracef("Called `RLock()` from %s:%d", file, line)
	m.Internal.RLock()
}

func (m TracingRWMutex) RUnlock() {
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		panic(runtimeCallerFailedMsg)
	}
	logger.Tracef("Called `RUnlock()` from %s:%d", file, line)
	m.Internal.RUnlock()
}

//
// END TRACING MUTEX.
//

type MqttManager struct {
	Body struct {
		Lock    TracingRWMutex
		Content MqttManagerBody
	}

	ConnectionInProgressLock sync.Mutex

	// Is being called from the outside if the outside knows that some things, which could have caused the initial
	// error, changed.
	TriggerTryPendingRegistrations func() error

	ShutdownChan      chan struct{}
	ShutdownCompleted chan struct{}
}

var Manager MqttManager

func (m *MqttManager) messageHandler(_ mqtt.Client, _ mqtt.Message) {
	panic("Unreachable: fallback on default message handler, this callback function is overwritten")
}

func (m *MqttManager) connectionLostHandler(_ mqtt.Client, err error) {
	m.Body.Lock.Lock()
	m.Body.Content.Initialized = false
	m.Body.Lock.Unlock()
	logger.Errorf("MQTT connection lost: %s\n", err.Error())
}

func (m *MqttManager) connectionEstablishedHandler(client mqtt.Client) {
	m.Body.Lock.Lock()
	m.Body.Content.Client = client
	m.Body.Content.Initialized = true
	m.Body.Lock.Unlock()

	if err := m.reloadOnReconnect(); err != nil {
		logger.Errorf("Failed to reload MQTT dispatcher after connection was established: %s\n", err.Error())
	}

	if err := m.TriggerTryPendingRegistrations(); err != nil {
		logger.Errorf("Failed to trigger parent reload after connection was established: %s\n", err.Error())
	}

	logger.Info("MQTT connection established")
}

func InitModule() {
	if logger.GetLevel() == logrus.TraceLevel && mqttAllowVerbose {
		mqtt.DEBUG = log.New(os.Stderr, "", 0)
		mqtt.ERROR = log.New(os.Stderr, "", 0)
	}
}

func NewMqttManager(config database.MqttConfig, retryHook func() error) (m *MqttManager, e error) {
	m = &MqttManager{
		Body: struct {
			Lock    TracingRWMutex
			Content MqttManagerBody
		}{
			Lock: TracingRWMutex{
				Internal: &sync.RWMutex{},
			},
			Content: MqttManagerBody{
				Subscriptions: make(map[string]Subscription),
				Client:        nil,
				Config:        config,
				Initialized:   false,
			},
		},
		TriggerTryPendingRegistrations: retryHook,
		ShutdownChan:                   make(chan struct{}),
		ShutdownCompleted:              make(chan struct{}),
	}

	Manager = *m

	go m.MQTTKeepalive()

	return m, nil
}

func (m *MqttManager) setConfig(config database.MqttConfig) {
	m.Body.Lock.Lock()
	defer m.Body.Lock.Unlock()
	m.Body.Content.Config = config
}

func (m *MqttManager) init() error {
	m.ConnectionInProgressLock.Lock()
	defer m.ConnectionInProgressLock.Unlock()

	m.Body.Lock.Lock()
	defer m.Body.Lock.Unlock()
	mqttEnabled := m.Body.Content.Config.Enabled

	if !mqttEnabled {
		logger.Debugf("MQTT subsystem is disabled according to server config")
		return nil
	}

	logger.Debugf(
		"Initializing MQTT subsystem for broker `%s@%s` (timeout: %v)...",
		m.Body.Content.Config.Username,
		m.Body.Content.Config.Host,
		MqttConnectTimeout,
	)

	opts := mqtt.NewClientOptions().
		AddBroker(types.MakeBrokerURI(
			m.Body.Content.Config.Host,
			m.Body.Content.Config.Port,
		)).
		SetClientID("homescript-smarthome").
		SetUsername(m.Body.Content.Config.Username).
		SetPassword(m.Body.Content.Config.Password)

	opts.SetConnectTimeout(MqttConnectTimeout)
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

	m.Body.Content.Initialized = true

	logger.Infof("Initialized MQTT subsystem for broker `%s@%s`", m.Body.Content.Config.Username, m.Body.Content.Config.Host)

	return nil
}

func (m *MqttManager) MQTTKeepalive() {
	for {
		select {
		case <-m.ShutdownChan:
			logger.Debug("Shutting down MQTT keepalive...")

			m.Body.Lock.Lock()
			if m.Body.Content.Client != nil {
				m.Body.Content.Client.Disconnect(0)
			}
			m.Body.Lock.Unlock()

			m.ShutdownCompleted <- struct{}{}

			return
		default:
		}

		if err := m.Status(); err != nil {
			m.Body.Lock.RLock()
			enabled := m.Body.Content.Config.Enabled
			m.Body.Lock.RUnlock()

			if enabled {
				logger.Errorf("MQTT could not be initialized: %s", err.Error())
			}
		}

		time.Sleep(5 * time.Second)
	}
}

func (m *MqttManager) IsConnected() bool {
	return m.Body.Content.Initialized && m.Body.Content.Client != nil && m.Body.Content.Client.IsConnected()
}

func (m *MqttManager) Status() error {
	m.ConnectionInProgressLock.Lock()
	m.ConnectionInProgressLock.Unlock()

	m.Body.Lock.RLock()
	isConnected := m.IsConnected()
	m.Body.Lock.RUnlock()

	if !isConnected {
		err := m.init()
		if err != nil {
			return err
		}

		m.Body.Lock.RLock()
		isNotConnected := m.Body.Content.Client == nil || !m.Body.Content.Client.IsConnected()
		m.Body.Lock.RUnlock()

		if isNotConnected {
			return fmt.Errorf("Not connected to broker")
		}
	}

	m.Body.Lock.Lock()
	token := m.Body.Content.Client.Publish(MqttHealthCheckTopic, MqttQOS, false, "")
	token.Wait()
	m.Body.Lock.Unlock()

	if token.Error() != nil {
		return token.Error()
	}

	return nil
}

func (m *MqttManager) Reload() error {
	m.unsubscribeAllSubscriptionsNonTracing()

	m.Body.Lock.Lock()

	// Purge old connection.
	if m.Body.Content.Client != nil {
		m.Body.Content.Client.Disconnect(MqttDisconnectTimeoutMillis)
	}

	m.Body.Lock.Unlock()

	// Apply new config and connect.
	if err := m.init(); err != nil {
		// nolint:goconst
		logger.Errorf("Failed to reload MQTT manager: %s", err.Error())
		return err
	}

	m.Body.Lock.RLock()
	enabled := m.Body.Content.Config.Enabled
	m.Body.Lock.RUnlock()

	if enabled {
		if err := m.resubscribeNonTracing(); err != nil {
			return err
		}
	}

	return nil
}

func (m *MqttManager) reloadOnReconnect() error {
	m.Body.Lock.RLock()
	mqttEnabled := m.Body.Content.Config.Enabled
	m.Body.Lock.RUnlock()

	if !mqttEnabled {
		return nil
	}

	m.unsubscribeAllSubscriptionsNonTracing()
	return m.resubscribeNonTracing()
}

func (m *MqttManager) unsubscribeAllSubscriptionsNonTracing() {
	// Unsubsribe from old topics to all topics that were previously subscribed to.
	m.Body.Lock.Lock()
	subscriptions := m.Body.Content.Subscriptions
	m.Body.Lock.Unlock()

	for topicName := range subscriptions {
		err := m.unsubscribeWithoutTracing(topicName)
		if err != nil {
			logger.Warn("Failed to unsubscribe from old topic: ", err.Error())
		}
	}
}

func (m *MqttManager) resubscribeNonTracing() error {
	m.Body.Lock.Lock()
	subscriptions := m.Body.Content.Subscriptions
	m.Body.Lock.Unlock()

	for topic, subscription := range subscriptions {
		err := m.subscribeWithoutTracing([]string{topic}, subscription.CallbackFn)
		if err != nil {
			logger.Errorf("Failed to reload MQTT manager: %s", err.Error())
			return err
		}
	}

	return nil
}

func (m *MqttManager) Shutdown() error {
	m.Body.Lock.Lock()
	initialized := m.Body.Content.Initialized
	m.Body.Lock.Unlock()

	if !initialized {
		return nil
	}

	subscriptions := m.Body.Content.Subscriptions
	for topic := range subscriptions {
		m.Body.Lock.Lock()
		token := m.Body.Content.Client.Unsubscribe(topic)
		token.Wait()
		m.Body.Lock.Unlock()

		if token.Error() != nil {
			return token.Error()
		}
	}

	m.Body.Lock.Lock()
	m.Body.Content.Client.Disconnect(MqttDisconnectTimeoutMillis)
	m.Body.Lock.Unlock()

	return nil
}

func (m *MqttManager) subscribeWithoutTracing(topics []string, callBack mqtt.MessageHandler) error {
	m.Body.Lock.Lock()
	initialized := m.Body.Content.Initialized
	m.Body.Lock.Unlock()

	if !initialized {
		return errors.New(notInitializedErrMsg)
	}

	topicsArgs := make(map[string]byte)
	for _, topic := range topics {
		topicsArgs[topic] = MqttQOS
	}

	m.Body.Lock.Lock()
	token := m.Body.Content.Client.SubscribeMultiple(topicsArgs, callBack)
	token.Wait()
	m.Body.Lock.Unlock()

	if token.Error() != nil {
		logger.Errorf("Could not subscribe to MQTT topics `%v`: %s", topics, token.Error())
		return token.Error()
	}

	return nil
}

func (m *MqttManager) Subscribe(topics []string, callBack mqtt.MessageHandler) error {
	if err := m.subscribeWithoutTracing(topics, callBack); err != nil {
		return err
	}

	m.Body.Lock.Lock()
	defer m.Body.Lock.Unlock()

	for _, topic := range topics {
		old, exists := m.Body.Content.Subscriptions[topic]
		if !exists {
			m.Body.Content.Subscriptions[topic] = Subscription{
				Consumers:  1,
				CallbackFn: callBack,
			}
			logger.Tracef("Created new consumer tracking for topic `%s`", topic)
		} else {
			old.Consumers++
			m.Body.Content.Subscriptions[topic] = old
			logger.Tracef("New consumer for topic `%s`, new count: %d", topic, old.Consumers)
		}
	}

	return nil
}

func (m *MqttManager) unsubscribeWithoutTracing(topic string) error {
	m.Body.Lock.Lock()
	defer m.Body.Lock.Unlock()
	if token := m.Body.Content.Client.Unsubscribe(topic); token.Wait() && token.Error() != nil {
		logger.Errorf("Could not unsubscribe from MQTT topic `%s`: %s", topic, token.Error())
		return token.Error()
	}

	return nil
}

func (m *MqttManager) Unsubscribe(topic string) error {
	m.Body.Lock.RLock()
	initialized := m.Body.Content.Initialized
	m.Body.Lock.RUnlock()

	if !initialized {
		return errors.New(notInitializedErrMsg)
	}

	m.Body.Lock.RLock()
	old, exists := m.Body.Content.Subscriptions[topic]
	m.Body.Lock.RUnlock()
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

		m.Body.Lock.Lock()
		delete(m.Body.Content.Subscriptions, topic)
		m.Body.Lock.Unlock()
		logger.Tracef("Deleted all subscriptions of topic `%s`", topic)
	} else {
		m.Body.Lock.Lock()
		m.Body.Content.Subscriptions[topic] = old
		m.Body.Lock.Unlock()
	}

	return nil
}

func (m *MqttManager) Publish(topic string, message string) error {
	m.Body.Lock.Lock()
	defer m.Body.Lock.Unlock()

	if !m.Body.Content.Initialized {
		return errors.New(notInitializedErrMsg)
	}

	token := m.Body.Content.Client.Publish(topic, MqttQOS, false, message)
	token.Wait()
	if token.Error() != nil {
		logger.Errorf("Could not publish to MQTT topic `%s`: %s", topic, token.Error())
	}
	return token.Error()
}
