package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/homescript"
	"github.com/smarthome-go/smarthome/server/middleware"
)

type CreateDeviceDriverRequest struct {
	Data           database.DeviceDriverData `json:"data"`
	HomescriptCode string                    `json:"code"`
}

type CreateDriverResponse struct {
	HmsId string `json:"hmsId"`
}

func ListDeviceDrivers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	drivers, err := database.ListDeviceDrivers()
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to list device drivers", Error: "database failure"})
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
	var request CreateDeviceDriverRequest
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}

	// Check if the combination of vendor ID and model ID is already used
	_, found, err := database.GetDeviceDriver(request.Data.VendorId, request.Data.ModelId)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to add device driver", Error: "database failure"})
		return
	}

	if found {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to add device driver", Error: fmt.Sprintf("there is already a device driver for model `%s` of vendor id `%s`", request.Data.ModelId, request.Data.VendorId)})
		return
	}

	if strings.Contains(request.Data.VendorId, " ") || utf8.RuneCountInString(request.Data.VendorId) > database.DEVICE_DRIVER_MODVEN_ID_LEN {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to add device driver", Error: fmt.Sprintf("the vendor id: '%s' must not exceed %d characters and must not include any whitespaces", request.Data.VendorId)})
		return
	}

	if strings.Contains(request.Data.ModelId, " ") || utf8.RuneCountInString(request.Data.ModelId) > database.DEVICE_DRIVER_MODVEN_ID_LEN {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to add device driver", Error: fmt.Sprintf("the model id: '%s' must not exceed %d characters and must not include any whitespaces", request.Data.ModelId)})
		return
	}

	if utf8.RuneCountInString(request.Data.Version) > database.DEVICE_DRIVER_VERSION_LEN {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to add device driver", Error: fmt.Sprintf("the version: '%s' must not exceed %d characters", request.Data.Version, database.DEVICE_DRIVER_VERSION_LEN)})
		return
	}

	homescriptId, err := database.CreateNewDeviceDriver(request.Data, request.HomescriptCode)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to create new device driver", Error: "database failure"})
		return
	}
	if err := json.NewEncoder(w).Encode(CreateDriverResponse{HmsId: homescriptId}); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to create new device driver", Error: "could not marshal or send result"})
		return
	}
}

// Deletes a Homescript by its Id, checks if it exists and if the user has permission to delete it
func DeleteDeviceDriver(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username, err := middleware.GetUserFromCurrentSession(w, r)
	if err != nil {
		return
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request HomescriptIdRequest
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}
	_, exists, err := database.GetUserHomescriptById(request.Id, username)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to delete Homescript: could not validate existence", Error: "database failure"})
		return
	}
	if !exists {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to delete Homescript", Error: "not found / permission denied: no data is associated to this id"})
		return
	}
	hasDependentAutomations, err := homescript.HasDependentAutomations(request.Id)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to delete Homescript: could not validate deletion safety", Error: "database failure"})
		return
	}
	if hasDependentAutomations {
		w.WriteHeader(http.StatusConflict)
		Res(w, Response{Success: false, Message: "can not delete Homescript: safety violation", Error: "script is used in one or more automations"})
		return
	}
	if err := database.DeleteHomescriptById(request.Id, username); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to delete Homescript", Error: "database failure"})
		return
	}
	Res(w, Response{Success: true, Message: "successfully deleted Homescript"})
}

// Modifies the metadata of a given Homescript
// func ModifyDeviceDriver(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Set("Content-Type", "application/json")
// 	username, err := middleware.GetUserFromCurrentSession(w, r)
// 	if err != nil {
// 		return
// 	}
// 	decoder := json.NewDecoder(r.Body)
// 	decoder.DisallowUnknownFields()
// 	var request CreateHomescriptRequest
// 	if err := decoder.Decode(&request); err != nil {
// 		w.WriteHeader(http.StatusBadRequest)
// 		Res(w, Response{Success: false, Message: "bad request", Error: "invalid request body"})
// 		return
// 	}
// 	_, exists, err := database.GetUserHomescriptById(request.Id, username)
// 	if err != nil {
// 		w.WriteHeader(http.StatusServiceUnavailable)
// 		Res(w, Response{Success: false, Message: "failed to modify device driver: could not validate existence", Error: "database failure"})
// 		return
// 	}
// 	if !exists {
// 		w.WriteHeader(http.StatusUnprocessableEntity)
// 		Res(w, Response{Success: false, Message: "failed to modify device driver", Error: "not found / permission denied: no data is associated to this id"})
// 		return
// 	}
// 	newHmsData := database.HomescriptData{
// 		Name:                request.Name,
// 		Description:         request.Description,
// 		QuickActionsEnabled: request.QuickActionsEnabled,
// 		SchedulerEnabled:    request.SchedulerEnabled,
// 		IsWidget:            request.IsWidget,
// 		Code:                request.Code,
// 		MDIcon:              request.MDIcon,
// 		Workspace:           request.Workspace,
// 	}
// 	if err := database.ModifyHomescriptById(
// 		request.Id,
// 		username,
// 		newHmsData,
// 	); err != nil {
// 		w.WriteHeader(http.StatusServiceUnavailable)
// 		Res(w, Response{Success: false, Message: "failed to modify Homescript", Error: "database failure"})
// 		return
// 	}
// 	Res(w, Response{Success: true, Message: "successfully modified Homescript"})
// }
