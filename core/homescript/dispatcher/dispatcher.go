package dispatcher

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"math"
	"math/big"
	"slices"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/smarthome-go/homescript/v3/homescript/analyzer/ast"
	herrors "github.com/smarthome-go/homescript/v3/homescript/errors"
	"github.com/smarthome-go/homescript/v3/homescript/runtime"
	"github.com/smarthome-go/homescript/v3/homescript/runtime/value"
	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/device/driver"
	driverTypes "github.com/smarthome-go/smarthome/core/device/driver/types"
	dispatcherTypes "github.com/smarthome-go/smarthome/core/homescript/dispatcher/types"
	"github.com/smarthome-go/smarthome/core/homescript/types"
	"github.com/smarthome-go/smarthome/core/scheduler"
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
	Hms                  types.Manager
	Mqtt                 *MqttManager
	DoneRegistrations    dispatcherTypes.Registrations
	PendingRegistrations PendingQueue

	// Used to prevent runaway retries
	LastRegistrationErrorTime time.Time
}

const ErrorCooldown = time.Second

var Instance InstanceT

func InitInstance(hms types.Manager, mqtt *MqttManager) error {
	Instance = InstanceT{
		Hms:  hms,
		Mqtt: mqtt,
		DoneRegistrations: dispatcherTypes.Registrations{
			Lock:                   sync.RWMutex{},
			Set:                    make(map[dispatcherTypes.RegistrationID]dispatcherTypes.RegisterInfo),
			MqttRegistrations:      make(map[string][]dispatcherTypes.RegistrationID),
			SchedulerRegistrations: make(map[string]dispatcherTypes.RegistrationID),
		},
		PendingRegistrations:      NewQueue(),
		LastRegistrationErrorTime: time.Time{},
	}

	if err := Instance.Mqtt.init(); err != nil {
		return err
	}

	return nil
}

func (i *InstanceT) MQTTStatus() error {
	return i.Mqtt.Status()
}

func (i *InstanceT) Reload(mqttConfig database.MqttConfig) error {
	i.Mqtt.setConfig(mqttConfig)
	return i.Mqtt.Reload()
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
	i.DoneRegistrations.Lock.Lock()
	defer i.DoneRegistrations.Lock.Unlock()

	for {
		potential := i.generatePotentialID()
		_, invalid := i.DoneRegistrations.Set[potential]

		if !invalid {
			// nolint:exhaustruct
			i.DoneRegistrations.Set[potential] = dispatcherTypes.RegisterInfo{}
			return potential
		}
	}
}

func (i *InstanceT) FreeRegistrationID(id dispatcherTypes.RegistrationID) {
	i.DoneRegistrations.Lock.Lock()
	delete(i.DoneRegistrations.Set, id)
	i.DoneRegistrations.Lock.Unlock()
}

func (i *InstanceT) registerInternal(info dispatcherTypes.RegisterInfo) (dispatcherTypes.RegistrationID, error) {
	if info.Function == nil {
		panic("Function cannot be <nil>")
	}

	//
	// TODO: check that this registration does not exist already.
	//

	if info.Function.CallMode.Kind() == dispatcherTypes.CallModeKindAdaptive {
		i.DoneRegistrations.Lock.RLock()
		registrations := i.DoneRegistrations.Set
		i.DoneRegistrations.Lock.RUnlock()

		for id, infoIter := range registrations {
			if infoIter.Function == nil {
				panic("Function may not be <nil>")
			}

			if infoIter.Function.CallMode.Kind() == dispatcherTypes.CallModeKindAdaptive && infoIter.ProgramID == info.ProgramID {
				// Remove this old registration.
				if err := i.Unregister(id); err != nil {
					return 0, err
				}

				logger.Debugf("Unregistered old ADAPTIVE job with id %d\n", id)
			}
		}
	}

	id := i.AllocRegistrationID()

	switch trigger := info.Trigger.(type) {
	case dispatcherTypes.CallBackTriggerMqtt:
		// Filter out any empty topics.
		topics := make([]string, 0)
		for _, topic := range trigger.Topics {
			if topic == "" {
				continue
			}

			topics = append(topics, topic)
		}

		// TODO: maybe check that a program cannot register twice.
		if err := i.Mqtt.Subscribe(topics, i.mqttCallBack); err != nil {
			// Delete allocated ID again. TODO: make deletion on failure more robust -> refactor code
			i.DoneRegistrations.Lock.Lock()
			delete(i.DoneRegistrations.Set, id)
			i.DoneRegistrations.Lock.Unlock()

			return 0, errors.WithMessage(err, "Could not register via MQTT manager")
		}

		i.DoneRegistrations.Lock.Lock()
		i.DoneRegistrations.Set[id] = info

		for _, topic := range topics {
			old, found := i.DoneRegistrations.MqttRegistrations[topic]
			if !found {
				i.DoneRegistrations.MqttRegistrations[topic] = make([]dispatcherTypes.RegistrationID, 0)
			}

			i.DoneRegistrations.MqttRegistrations[topic] = append(old, id)
		}

		i.DoneRegistrations.Lock.Unlock()
	case dispatcherTypes.CallBackTriggerAtTime:
		// TODO: need job ID here, this is not unique.
		schedulerTag := fmt.Sprintf("dispatcher-%s-%s", info.ProgramID, info.Function.Ident)

		if err := scheduler.Manager.CreateNewScheduleInternal(
			trigger.Hour,
			trigger.Minute,
			schedulerTag,
			i.timeCallBack,
			info,
		); err != nil {
			// Delete allocated ID again.
			i.DoneRegistrations.Lock.Lock()
			delete(i.DoneRegistrations.Set, id)
			i.DoneRegistrations.Lock.Unlock()

			return 0, fmt.Errorf("Could not register time: %s", err.Error())
		}

		i.DoneRegistrations.Lock.Lock()
		i.DoneRegistrations.Set[id] = info
		i.DoneRegistrations.SchedulerRegistrations[schedulerTag] = id
		i.DoneRegistrations.Lock.Unlock()
	default:
		panic(fmt.Sprintf("Unreachable: introduced a new trigger type (%v) without updating this code", info.Trigger))
	}

	return id, nil
}

func (i *InstanceT) Unregister(id dispatcherTypes.RegistrationID) error {
	i.DoneRegistrations.Lock.Lock()

	_, valid := i.DoneRegistrations.Set[id]
	if !valid {
		i.DoneRegistrations.Lock.Unlock()
		return fmt.Errorf("Cannot unregister registration with ID %d: not registered", id)
	}

	delete(i.DoneRegistrations.Set, id)

	var unregisterErr error

	mqttRegistrations := i.DoneRegistrations.MqttRegistrations
	i.DoneRegistrations.Lock.Unlock()

	// Also delete all references in MQTT.
	for topic, ids := range mqttRegistrations {
		if slices.Contains(ids, id) {
			if err := i.Mqtt.Unsubscribe(topic); err != nil && unregisterErr == nil {
				unregisterErr = err
			}

			if len(ids) == 1 {
				// Remove entire topic from map.
				i.DoneRegistrations.Lock.Lock()
				delete(i.DoneRegistrations.MqttRegistrations, topic)
				i.DoneRegistrations.Lock.Unlock()
				continue
			}

			newSlice := make([]dispatcherTypes.RegistrationID, 0)
			for _, idToCheck := range ids {
				if idToCheck == id {
					continue
				}
				newSlice = append(newSlice, idToCheck)
			}

			i.DoneRegistrations.Lock.Lock()
			i.DoneRegistrations.MqttRegistrations[topic] = newSlice
			i.DoneRegistrations.Lock.Unlock()
		}
	}

	// Delete reference in scheduler if required.
	for tag, idToCheck := range i.DoneRegistrations.SchedulerRegistrations {
		fmt.Printf("REGISTRATION: ===== %d", idToCheck)
		if id == idToCheck {
			if err := scheduler.Manager.RemoveScheduleInternal(tag); err != nil && unregisterErr == nil {
				unregisterErr = err
			}
			delete(i.DoneRegistrations.SchedulerRegistrations, tag)
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
	jobID uint64,
	meta CallBackMeta,
) {
	job, found := i.Hms.GetJobById(jobID)
	if !found {
		logger.Errorf("Could not dispatch callback into HMS job with ID %d (callback mode attaching)", jobID)
		return
	}

	logger.Tracef("Attaching callback for function `%s` for job ID %d", info.Function.Ident, job.JobID)

	_ = job.VM.SpawnAsync(runtime.FunctionInvocation{
		Function:          info.Function.Ident,
		LiteralName:       info.Function.IdentIsLiteral,
		Args:              meta.Args,
		FunctionSignature: meta.FunctionSignature,
	},
		nil,
		nil,
		nil,
	)

	// TODO: WHAT to do with this core.
}

func (i *InstanceT) CallBack(info dispatcherTypes.RegisterInfo, meta CallBackMeta) {
	switch callMode := info.Function.CallMode.(type) {
	case dispatcherTypes.CallModeAllocating:
		go i.allocatingCall(
			info,
			meta,
			callMode.Context,
		)
	case dispatcherTypes.CallModeAdaptive:
		for _, job := range i.Hms.GetJobList() {
			if job.HmsID == info.ProgramID {
				i.AttachingCall(info, job.JobID, meta)
				return
			}
		}

		logger.Tracef("Could not perform attaching call for adaptive job `%s`\n", info.ProgramID)

		go i.allocatingCall(
			info,
			meta,
			callMode.AllocatingFallback.Context,
		)
	case dispatcherTypes.CallModeAttaching:
		go i.AttachingCall(info, callMode.HMSJobID, meta)
	default:
		panic("A new call mode was added without updating this code")
	}
}

func (i *InstanceT) allocatingCall(
	info dispatcherTypes.RegisterInfo,
	meta CallBackMeta,
	execContext types.ExecutionContext,
) {
	logger.Tracef("Performing allocating call for program `%s`...\n", info.ProgramID)
	cancelCtx, cancelFnc := context.WithCancel(context.Background())

	invocation := runtime.FunctionInvocation{
		Function:          info.Function.Ident,
		LiteralName:       info.Function.IdentIsLiteral,
		Args:              meta.Args,
		FunctionSignature: meta.FunctionSignature,
	}

	spew.Dump(meta.FunctionSignature)

	var buffer bytes.Buffer

	var err error
	var res types.HmsRes

	switch c := execContext.(type) {
	case types.ExecutionContextDriver:
		hmsRes, errTemp := driver.Manager.InvokeDriverFunc(
			driverTypes.DriverInvocationIDs{
				DeviceID: c.DeviceID,
				VendorID: c.DriverVendor,
				ModelID:  c.DriverModel,
			},
			driver.FunctionCall{
				Invocation: invocation,
			},
		)

		err = errTemp
		res = types.HmsRes{
			Errors:             hmsRes.Errors,
			Singletons:         nil,
			ReturnValue:        nil,
			CalledFunctionSpan: herrors.Span{},
		}
	default:
		if c.Username() == nil {
			panic("Can only dispatch to driver or user environment")
		}

		resTemp, errTemp := i.Hms.RunUserScript(
			info.ProgramID,
			*c.Username(),
			nil,
			types.Cancelation{
				Context:    cancelCtx,
				CancelFunc: cancelFnc,
			},
			&buffer,
			nil,
			nil,
		)

		res = resTemp
		err = errTemp
	}

	if err != nil {
		logger.Errorf("Failed to run dispatcher target HMS: %s\n", err.Error())
		cancelFnc()
		return
	}

	if res.Errors.ContainsError {
		// TODO: better way to handle errors: maybe system log or admin user notication?
		for _, error := range res.Errors.Diagnostics {
			fmt.Println(error.String())
		}

		panic("HMS crashed on invocation")
	}

	cancelFnc()
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

	i.DoneRegistrations.Lock.RLock()
	defer i.DoneRegistrations.Lock.RUnlock()

	for _, registrationID := range i.DoneRegistrations.MqttRegistrations[topic] {
		registration, found := i.DoneRegistrations.Set[registrationID]
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
				ReturnType: ast.NewNullType(herrors.Span{}),
			},
		})
	}
}

func (i *InstanceT) timeCallBack(registration dispatcherTypes.RegisterInfo) {
	trigger := registration.Trigger.(dispatcherTypes.CallBackTriggerAtTime)
	i.CallBack(registration, CallBackMeta{
		Args: []value.Value{
			*value.NewValueInt(int64(time.Since(trigger.RegisteredAt).Seconds())),
		},
		FunctionSignature: runtime.FunctionInvocationSignature{
			Params: []runtime.FunctionInvocationSignatureParam{
				{Ident: "elapsed", Type: ast.NewIntType(herrors.Span{})},
			},
			ReturnType: ast.NewNullType(herrors.Span{}),
		},
	})
}
