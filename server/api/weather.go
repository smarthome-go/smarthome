package api

import (
	"encoding/json"
	"net/http"
	"unicode/utf8"

	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/services/weather"
)

type newApiKeyRequest struct {
	Key string `json:"key"`
}

func GetWeather(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	res, err := weather.GetCurrentWeather()
	if err != nil {
		if err.Error() == "invalid api key" {
			w.WriteHeader(http.StatusPaymentRequired)
			Res(w, Response{Success: false, Message: "could not get weather data", Error: "this server is not configured to use a valid OWM API key"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		Res(w, Response{Success: false, Message: "could not get weather data", Error: "internal server error"})
		return
	}
	if err := json.NewEncoder(w).Encode(res); err != nil {
		log.Error(err.Error())
		Res(w, Response{Success: false, Message: "could not get weather data", Error: "could not encode content"})
	}
}

// Used as a fallback when the normal weather fails (due to network conditions)
func GetCachedWeather(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	res, err := database.GetWeatherDataRecords(30)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		Res(w, Response{Success: false, Message: "could not get cached weather data", Error: "internal server error"})
		return
	}
	// Select the youngest entry of them all
	if len(res) == 0 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "could not get cached weather data", Error: "no cached data available"})
		return
	}
	if err := json.NewEncoder(w).Encode(res[0]); err != nil {
		log.Error(err.Error())
		Res(w, Response{Success: false, Message: "could not get cached weather data", Error: "could not encode content"})
	}
}

func UpdateOpenWeatherMapApiKey(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request newApiKeyRequest
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}
	// Check if the key is too long for the database
	if utf8.RuneCountInString(request.Key) > 64 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "will not update API key", Error: "the maximum key length is 64 characters"})
		return
	}
	if err := database.UpdateOpenWeatherMapApiKey(request.Key); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "could not update API key", Error: "database failure"})
		return
	}
	Res(w, Response{Success: true, Message: "successfully updated the OWM API-key"})
}

func TestOpenWeatherMapApiKey(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
}

// Is used to flush the weather cache manually
// Deletes all records, regardless of their age
func PurgeWeatherCache(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := database.PurgeWeatherData(); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to purge weather cache", Error: "database failure"})
		return
	}
	Res(w, Response{Success: true, Message: "successfully purged weather cache"})
}
