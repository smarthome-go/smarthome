package api

import (
	"encoding/json"
	"net/http"
	"unicode/utf8"

	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/services/weather"
)

type newApiKeyRequest struct {
	Key string `json:"string"`
}

func GetWeather(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	res, err := weather.GetWeather()
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
