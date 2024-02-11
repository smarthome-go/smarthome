package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"unicode/utf8"

	"golang.org/x/exp/utf8string"

	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/homescript"
	"github.com/smarthome-go/smarthome/server/middleware"
)

type AddDeviceRequest struct {
	Type           database.DEVICE_TYPE `json:"type"`
	Id             string               `json:"id"`
	Name           string               `json:"name"`
	RoomId         string               `json:"roomId"`
	DriverVendorId string               `json:"driverVendorId"`
	DriverModelId  string               `json:"driverModelId"`
}

type ModifyDeviceRequest struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type ConfigureDeviceRequest struct {
	ID   string      `json:"id"`
	Data interface{} `json:"data"`
}

type DeleteDeviceRequest struct {
	Id string `json:"id"`
}

// Returns a list of available devices as JSON to the user, no authentication required
func GetAllDevices(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	devices, err := database.ListAllDevices()
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "database error", Error: "database failure"})
		return
	}
	if err := json.NewEncoder(w).Encode(devices); err != nil {
		log.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		Res(w, Response{Success: false, Message: "failed to get devices", Error: "could not encode content"})
		return
	}
}

// Only returns devices which the user has access to, authentication required
func GetUserDevices(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username, err := middleware.GetUserFromCurrentSession(w, r)
	if err != nil {
		return
	}
	devices, err := database.ListUserDevices(username)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "database error", Error: "database error"})
		return
	}
	if err := json.NewEncoder(w).Encode(devices); err != nil {
		log.Error(err.Error())
		Res(w, Response{Success: false, Message: "failed to get personal devices", Error: "could not encode content"})
	}
}

// Returns a list of available devices as JSON to the user, no authentication required
func GetAllDevicesRich(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	devices, err := homescript.ListAllDevices()
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "database error", Error: "database failure"})
		return
	}
	if err := json.NewEncoder(w).Encode(devices); err != nil {
		log.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		Res(w, Response{Success: false, Message: "failed to get devices", Error: "could not encode content"})
		return
	}
}

// Only returns devices which the user has access to, authentication required
func GetUserDevicesRich(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username, err := middleware.GetUserFromCurrentSession(w, r)
	if err != nil {
		return
	}
	devices, err := homescript.ListPersonalDevices(username)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "database error", Error: "database error"})
		return
	}
	if err := json.NewEncoder(w).Encode(devices); err != nil {
		log.Error(err.Error())
		Res(w, Response{Success: false, Message: "failed to get personal devices", Error: "could not encode content"})
	}
}

// Creates a device in the database
func CreateDevice(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request AddDeviceRequest
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}

	// Validate length and encoding
	if strings.Contains(request.Id, " ") || !utf8string.NewString(request.Id).IsASCII() {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "bad request", Error: "id should only include ASCII characters and must not have whitespaces"})
		return
	}
	if utf8.RuneCountInString(request.Id) > 20 || utf8.RuneCountInString(request.Name) > 30 {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "bad request", Error: "maximum lengths for id and name are 20 and 30"})
		return
	}

	// Validate that the device type is legal
	parsedType, valid := database.ParseDeviceType(string(request.Type))
	if !valid {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to create device", Error: fmt.Sprintf("illegal device type: `%s`", request.Type)})
		return
	}

	// Validate that no conflicts are present
	_, alreadyExists, err := database.GetDeviceById(request.Id)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to create device", Error: "database failure"})
		return
	}
	if alreadyExists {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to create device", Error: "id already exists"})
		return
	}

	// Validate that the room exists
	_, roomExists, err := database.GetRoomDataById(request.RoomId)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to create device", Error: "database failure"})
		return
	}
	if !roomExists {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to create device", Error: "invalid room id"})
		return
	}

	// Validate that there is a driver for this `vendorId` and `modelId`
	_, driverExists, err := database.GetDeviceDriver(request.DriverVendorId, request.DriverModelId)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to create device", Error: "database failure"})
		return
	}
	if !driverExists {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to create device", Error: fmt.Sprintf("this device model is not supported: no such driver for modelID `%s` of vendorID `%s`", request.DriverModelId, request.DriverVendorId)})
		return
	}

	driverFound, hmsErr, dbErr := homescript.CreateDevice(
		parsedType,
		request.Id,
		request.Name,
		request.RoomId,
		request.DriverVendorId,
		request.DriverModelId,
	)

	if dbErr != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to create device", Error: "database failure"})
		return
	}

	if hmsErr != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to create device", Error: hmsErr.Error()})
		return
	}

	if !driverFound {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "failed to create device", Error: "No such driver"})
		return
	}

	Res(w, Response{Success: true, Message: "successfully created device"})
}

func ModifyDevice(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request ModifyDeviceRequest
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}
	_, found, err := database.GetDeviceById(request.Id)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to modify device", Error: "database failure"})
		return
	}
	if !found {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to modify device", Error: "no device with id exists"})
		return
	}

	// Validate length of the name
	if len(request.Name) > 30 {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "bad request", Error: "maximum name length of 30 chars. was exceeded"})
		return
	}

	if err := database.ModifyDeviceName(request.Id, request.Name); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to modify device", Error: "database failure"})
		return
	}
	Res(w, Response{Success: true, Message: "successfully modified device"})
}

func ConfigureDevice(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request ConfigureDeviceRequest
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}

	found, validateErr, dbErr := homescript.ValidateDeviceConfigurationChange(request.ID, request.Data)
	if dbErr != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to modify device", Error: "database failure"})
		return
	}

	if validateErr != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "device configuration schema validation failed", Error: validateErr.Error()})
		return
	}

	if !found {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{
			Success: false,
			Message: "failed to configure device",
			Error:   fmt.Sprintf("the device `%s` does not exist", request.ID),
		})
		return
	}

	if err := homescript.StoreDeviceSingletonConfigUpdate(request.ID, request.Data); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to configure device", Error: "database failure"})
		return
	}

	Res(w, Response{Success: true, Message: "successfully configured device"})
}

func DeleteDevice(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request DeleteDeviceRequest
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}
	_, found, err := database.GetDeviceById(request.Id)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to delete device", Error: "database failure"})
		return
	}
	if !found {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to delete device", Error: "no device with id exists"})
		return
	}
	if err := database.DeleteDevice(request.Id); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to delete device", Error: "database failure"})
		return
	}
	Res(w, Response{Success: true, Message: "successfully deleted device"})
}
