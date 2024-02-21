package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/homescript"
)

type DeviceDriverRequest struct {
	VendorID string `json:"vendorId"`
	ModelID  string `json:"modelId"`
}

type ConfigureDriverRequest struct {
	Driver database.DriverTuple `json:"driver"`
	Data   interface{}          `json:"data"`
}

type DeviceDriverAddRequest struct {
	VendorId string `json:"vendorId"`
	ModelId  string `json:"modelId"`
	Name     string `json:"name"`
	Version  string `json:"version"`
	// If this is `nil`, the backend automatically generates code for the driver.
	HomescriptCode *string `json:"homescriptCode"`
}

func ListDeviceDrivers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	drivers, err := homescript.ListDriversWithStoredConfig()
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to list device drivers", Error: "database error"})
		return
	}
	if err := json.NewEncoder(w).Encode(drivers); err != nil {
		log.Error(err.Error())
		Res(w, Response{Success: false, Message: "failed to list device drivers", Error: "could not encode response"})
	}
}

func CreateDeviceDriver(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request DeviceDriverAddRequest
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}

	// Check if the combination of vendor ID and model ID is already used
	_, alreadyExists, dbErr := database.GetDeviceDriver(request.VendorId, request.ModelId)
	if dbErr != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to add device driver", Error: "database failure"})
		return
	}
	if alreadyExists {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{
			Success: false,
			Message: "failed to add device driver",
			Error:   fmt.Sprintf("there is already a device driver for model `%s` of vendor id `%s`", request.ModelId, request.VendorId),
		})
		return
	}

	// Check that the vendor and model ids are not too long
	if strings.Contains(request.VendorId, " ") || utf8.RuneCountInString(request.VendorId) > database.DEVICE_DRIVER_MODVEN_ID_LEN {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{
			Success: false,
			Message: "failed to add device driver",
			Error:   fmt.Sprintf("the vendor id: '%s' must not exceed %d characters and must not include any whitespaces", request.VendorId, database.DEVICE_DRIVER_MODVEN_ID_LEN),
		})
		return
	}
	if strings.Contains(request.ModelId, " ") || utf8.RuneCountInString(request.ModelId) > database.DEVICE_DRIVER_MODVEN_ID_LEN {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{
			Success: false,
			Message: "failed to add device driver",
			Error:   fmt.Sprintf("the model id: '%s' must not exceed %d characters and must not include any whitespaces", request.ModelId, database.DEVICE_DRIVER_MODVEN_ID_LEN),
		})
		return
	}

	// Check that the version is not too long
	if utf8.RuneCountInString(request.Version) > database.DEVICE_DRIVER_VERSION_LEN {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{
			Success: false,
			Message: "failed to add device driver", Error: fmt.Sprintf("the version: '%s' must not exceed %d characters", request.Version, database.DEVICE_DRIVER_VERSION_LEN),
		})
		return
	}

	_, dbErr = homescript.ParseDriverVersion(request.Version)
	if dbErr != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{
			Success: false,
			Message: "failed to add device driver", Error: fmt.Sprintf("the version: '%s' is not valid: %s", request.Version, dbErr.Error()),
		})
		return
	}

	hmsErr, dbErr := homescript.CreateDriver(
		request.VendorId,
		request.ModelId,
		request.Name,
		request.Version,
		request.HomescriptCode,
	)

	if dbErr != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to create new device driver", Error: "database failure"})
		return
	}

	if hmsErr != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "schema validation failed", Error: hmsErr.Error()})
		return
	}

	Res(w, Response{Success: true, Message: "successfully created device driver"})
}

func ModifyDeviceDriver(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request database.DeviceDriver
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}

	// Check that the new version is not too long
	if utf8.RuneCountInString(request.Version) > database.DEVICE_DRIVER_VERSION_LEN {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{
			Success: false,
			Message: "failed to modify device driver", Error: fmt.Sprintf("the version: '%s' must not exceed %d characters", request.Version, database.DEVICE_DRIVER_VERSION_LEN),
		})
		return
	}

	_, err := homescript.ParseDriverVersion(request.Version)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{
			Success: false,
			Message: "failed to modify device driver", Error: fmt.Sprintf("the version: '%s' is not valid: %s", request.Version, err.Error()),
		})
		return
	}

	_, wasFound, err := database.GetDeviceDriver(request.VendorId, request.ModelId)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to modify device driver", Error: "database failure"})
		return
	}

	if !wasFound {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{
			Success: false,
			Message: "failed to modify device driver",
			Error:   fmt.Sprintf("the device driver `%s:%s` does not exist", request.ModelId, request.VendorId),
		})
		return
	}

	if err := database.ModifyDeviceDriver(request); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to modify device driver", Error: "database failure"})
		return
	}

	Res(w, Response{Success: true, Message: "successfully modified device driver"})
}

func ConfigureDeviceDriver(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields() // TODO: is this the right way?

	var request ConfigureDriverRequest

	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}

	found, validationErr, dbErr := homescript.ValidateDriverConfigurationChange(
		request.Driver.VendorID,
		request.Driver.ModelID,
		request.Data,
	)

	if dbErr != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to configure device driver", Error: "database failure"})
		return
	}

	if validationErr != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "driver schema validation failed", Error: validationErr.Error()})
		return
	}

	if !found {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{
			Success: false,
			Message: "failed to configure device driver",
			Error:   fmt.Sprintf("the device driver `%s:%s` does not exist", request.Driver.ModelID, request.Driver.VendorID),
		})
		return
	}

	if dbErr = homescript.StoreDriverSingletonConfigUpdate(
		request.Driver.VendorID,
		request.Driver.ModelID,
		request.Data,
	); dbErr != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to configure device driver", Error: "database failure"})
		return
	}

	Res(w, Response{Success: true, Message: "successfully configured device driver"})
}

// Deletes a Homescript by its Id, checks if it exists and if the user has permission to delete it
func DeleteDeviceDriver(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request DeviceDriverRequest
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}

	_, exists, err := database.GetDeviceDriver(request.VendorID, request.ModelID)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to delete driver: could not validate existence", Error: "database failure"})
		return
	}

	if !exists {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to delete driver", Error: "not found / permission denied: no data is associated to this vendor + model ID"})
		return
	}

	hasDependentDevices, err := database.DriverHasDependentDevices(request.VendorID, request.ModelID)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to delete driver: could not validate deletion safety", Error: "database failure"})
		return
	}
	if hasDependentDevices {
		w.WriteHeader(http.StatusConflict)
		Res(w, Response{Success: false, Message: "cannot delete driver: safety violation", Error: "driver controls one or more devices"})
		return
	}

	if err := database.DeleteDeviceDriver(request.VendorID, request.ModelID); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to delete driver", Error: "database failure"})
		return
	}

	Res(w, Response{Success: true, Message: "successfully deleted driver"})
}
