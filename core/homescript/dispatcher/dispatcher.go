package dispatcher

import (
	"crypto/rand"
	"fmt"
	"math"
	"math/big"
	"slices"
	"sync"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/smarthome-go/homescript/v3/homescript/analyzer/ast"
	herrors "github.com/smarthome-go/homescript/v3/homescript/errors"
	"github.com/smarthome-go/homescript/v3/homescript/runtime"
	"github.com/smarthome-go/homescript/v3/homescript/runtime/value"
	"github.com/smarthome-go/smarthome/core/database"
	dispatcherTypes "github.com/smarthome-go/smarthome/core/homescript/dispatcher/types"
	"github.com/smarthome-go/smarthome/core/homescript/types"
)

const registrationIDNumDigits = 16

var logger *logrus.Logger

func InitLogger(log *logrus.Logger) {
	logger = log
}

//
// Dispatcher implementation.
//

type InstanceT struct {
	Hms           types.Manager
	Mqtt          *MqttManager
	Registrations dispatcherTypes.Registrations
}

var Instance InstanceT

func InitInstance(hms types.Manager, mqtt *MqttManager) {
	Instance = InstanceT{
		Hms:  hms,
		Mqtt: mqtt,
		Registrations: dispatcherTypes.Registrations{
			Lock:              sync.RWMutex{},
			Set:               make(map[dispatcherTypes.RegistrationID]dispatcherTypes.RegisterInfo),
			MqttRegistrations: make(map[string][]dispatcherTypes.RegistrationID),
		},
	}
}

// TODO: same as for the HMS manager, join these two functions.
func (i *InstanceT) generatePotentialID() dispatcherTypes.RegistrationID {
	maxLimit := int64(int(math.Pow10(registrationIDNumDigits)) - 1)
	lowLimit := uint64(math.Pow10(registrationIDNumDigits - 1))

	randomNum, err := rand.Int(rand.Reader, big.NewInt(maxLimit))
	if err != nil {
		panic(err.Error())
	}

	randomInt := dispatcherTypes.RegistrationID(randomNum.Int64())

	if uint64(randomInt) <= lowLimit {
		randomInt += dispatcherTypes.RegistrationID(lowLimit)
	}

	return randomInt
}

func (i *InstanceT) AllocRegistrationID() dispatcherTypes.RegistrationID {
	for {
		potential := i.generatePotentialID()
		i.Registrations.Lock.Lock()
		_, invalid := i.Registrations.Set[potential]

		if !invalid {
			// nolint:exhaustruct
			i.Registrations.Set[potential] = dispatcherTypes.RegisterInfo{}
			i.Registrations.Lock.Unlock()
			return potential
		}
	}
}

func (i *InstanceT) FreeRegistrationID(id dispatcherTypes.RegistrationID) {
	i.Registrations.Lock.Lock()
	delete(i.Registrations.Set, id)
	i.Registrations.Lock.Unlock()
}

func (i *InstanceT) Register(info dispatcherTypes.RegisterInfo) (dispatcherTypes.RegistrationID, error) {
	if info.Function == nil {
		panic("Function cannot be <nil>")
	}

	id := i.AllocRegistrationID()

	switch trigger := info.Trigger.(type) {
	case dispatcherTypes.CallBackTriggerMqtt:
		// TODO: maybe check that a program cannot register twice.
		if err := i.Mqtt.Subscribe(trigger.Topics, i.mqttCallBack); err != nil {
			return 0, errors.WithMessage(err, "Could not register via MQTT manager")
		}

		i.Registrations.Lock.Lock()
		i.Registrations.Set[id] = info

		for _, topic := range trigger.Topics {
			old, found := i.Registrations.MqttRegistrations[topic]
			if !found {
				i.Registrations.MqttRegistrations[topic] = make([]dispatcherTypes.RegistrationID, 0)
			}

			i.Registrations.MqttRegistrations[topic] = append(old, id)
		}

		i.Registrations.Lock.Unlock()
	default:
		panic(fmt.Sprintf("Unreachable: introduced a new trigger type (%v) without updating this code", info.Trigger))
	}

	return id, nil
}

func (i *InstanceT) Unregister(id dispatcherTypes.RegistrationID) error {
	i.Registrations.Lock.Lock()
	defer i.Registrations.Lock.Unlock()

	_, valid := i.Registrations.Set[id]
	if !valid {
		return fmt.Errorf("Cannot unregister registration with ID %d: not registered", id)
	}

	delete(i.Registrations.Set, id)

	var unregisterErr error

	// Also delete all references in MQTT.
	for topic, ids := range i.Registrations.MqttRegistrations {
		if slices.Contains[[]dispatcherTypes.RegistrationID](ids, id) {
			if err := i.Mqtt.Unsubscribe(topic); err != nil && unregisterErr == nil {
				unregisterErr = err
			}

			if len(ids) == 1 {
				// Remove entire topic from map.
				delete(i.Registrations.MqttRegistrations, topic)
				continue
			}

			newSlice := make([]dispatcherTypes.RegistrationID, 0)
			for _, idToCheck := range ids {
				if idToCheck == id {
					continue
				}
				newSlice = append(newSlice, idToCheck)
			}

			i.Registrations.MqttRegistrations[topic] = newSlice
		}
	}

	logger.Debugf("dispatcher: Unregistered ID %d", id)

	return unregisterErr
}

type CallBackMeta struct {
	Args              []value.Value
	FunctionSignature runtime.FunctionInvocationSignature
}

func (i *InstanceT) AttachingCall(
	info dispatcherTypes.RegisterInfo,
	call dispatcherTypes.CallModeAttaching,
	meta CallBackMeta,
) {
	job, found := i.Hms.GetJobById(call.HMSJobID)
	if !found {
		logger.Errorf("Could not dispatch callback into HMS job with ID %d (callback mode attaching)", call.HMSJobID)
		return
	}

	logger.Tracef("Attaching callback for function `%s` for job ID %d", info.Function.Ident, job.JobID)

	_ = job.VM.SpawnAsync(runtime.FunctionInvocation{
		Function:          info.Function.Ident,
		LiteralName:       info.Function.IdentIsLiteral,
		Args:              meta.Args,
		FunctionSignature: meta.FunctionSignature,
	}, nil)

	// TODO: WHAT to do with this core.
}

func (i *InstanceT) CallBack(info dispatcherTypes.RegisterInfo, meta CallBackMeta) {
	switch callMode := info.Function.CallMode.(type) {
	case dispatcherTypes.CallModeAllocating:
		panic("not supported")
	case dispatcherTypes.CallModeAdaptive:
		panic("not supported")
	case dispatcherTypes.CallModeAttaching:
		i.AttachingCall(info, callMode, meta)
	default:
		panic("A new call mode was added without updating this code")
	}
}

type MqttMessage struct {
	Topic   string
	Payload string
}

func (i *InstanceT) mqttCallBack(_ mqtt.Client, message mqtt.Message) {
	// Invoke all MQTT registrations for this topic.
	topic := message.Topic()
	payload := string(message.Payload())

	logger.Tracef("Mqtt Callback: topic: `%s`, payload: `%s`", topic, payload)

	i.Registrations.Lock.RLock()
	defer i.Registrations.Lock.RUnlock()
	for _, registrationID := range i.Registrations.MqttRegistrations[topic] {
		registration, found := i.Registrations.Set[registrationID]
		if !found {
			panic(fmt.Sprintf("Registered MQTT ID not found: %d", registrationID))
		}

		i.CallBack(registration, CallBackMeta{
			Args: []value.Value{
				*value.NewValueString(string(message.Payload())),
				*value.NewValueString(message.Topic()),
			},
			FunctionSignature: runtime.FunctionInvocationSignature{
				Params: []runtime.FunctionInvocationSignatureParam{
					{
						Ident: "topic",
						Type:  ast.NewStringType(herrors.Span{}),
					},
					{
						Ident: "payload",
						Type:  ast.NewStringType(herrors.Span{}),
					}},
				ReturnType: nil,
			},
		})
	}
}
