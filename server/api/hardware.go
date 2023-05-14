package api

import (
	"encoding/json"
	"net/http"
	"unicode/utf8"

	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/hardware"
)

// Contains all endpoints for managing hardware

/*
Section for default node hardware
(https://github.com/smarthome-go/node)
*/

type addhardwareNodeRequest struct {
	Url  string   `json:"url"`
	Data nodeData `json:"data"`
}

type modifyHardwareNodeRequest struct {
	Url  string   `json:"url"`
	Data nodeData `json:"data"`
}

type nodeData struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
	Token   string `json:"token"`
}

type deleteHardwareNodeRequest struct {
	Url string `json:"url"`
}

// Returns a list of configured hardware nodes and their state
func ListHardwareNodes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	nodes, err := database.GetHardwareNodes()
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "could not list hardware nodes", Error: "database failure"})
		return
	}
	if err := json.NewEncoder(w).Encode(nodes); err != nil {
		log.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		Res(w, Response{Success: false, Message: "could not list hardware nodes", Error: "could not encode content"})
		return
	}
}

// Returns a list of configured hardware nodes and their state (no privileged information included)
func ListHardwareNodesNoPriv(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	nodes, err := database.GetHardwareNodes()
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "could not list hardware nodes", Error: "database failure"})
		return
	}
	outputNodes := make([]database.HardwareNode, 0)
	for _, node := range nodes {
		outputNodes = append(outputNodes, database.HardwareNode{
			Name:    node.Name,
			Url:     node.Url,
			Token:   "redacted",
			Online:  node.Online,
			Enabled: node.Enabled,
		})
	}
	if err := json.NewEncoder(w).Encode(outputNodes); err != nil {
		log.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		Res(w, Response{Success: false, Message: "could not list hardware nodes", Error: "could not encode content"})
		return
	}
}

func ListHardwareNodesWithCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	// Run the healcheck first (will update the database entries if node states change)
	if err := hardware.RunNodeCheck(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		Res(w, Response{Success: false, Message: "could not list hardware nodes", Error: "healthcheck failed: backend failure"})
		return
	}
	// Return the hardware nodes from the database
	nodes, err := database.GetHardwareNodes()
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "could not list hardware nodes", Error: "database failure"})
		return
	}
	if err := json.NewEncoder(w).Encode(nodes); err != nil {
		log.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		Res(w, Response{Success: false, Message: "could not list hardware nodes", Error: "could not encode content"})
		return
	}
}

// Creates a hardware node
func CreateHardwareNode(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request addhardwareNodeRequest
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}
	// Validate length of the URL
	if utf8.RuneCountInString(request.Url) > 50 {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "bad request", Error: "the URL may not exceed 50 characters"})
		return
	}
	// Validate length of the name
	if utf8.RuneCountInString(request.Data.Name) > 30 {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "bad request", Error: "the name may not exceed 30 characters"})
		return
	}
	// Validate length of the token
	if utf8.RuneCountInString(request.Data.Token) > 100 {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "bad request", Error: "the token may not exceed 100 characters"})
		return
	}
	// Validate that no conflicts are present
	_, alreadyExists, err := database.GetHardwareNodeByUrl(request.Url)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to create hardware node", Error: "database failure"})
		return
	}
	if alreadyExists {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to create hardware node", Error: "a node with the same URL already exists"})
		return
	}
	// Create the node in the database
	if err := database.CreateHardwareNode(database.HardwareNode{
		Url:     request.Url,
		Name:    request.Data.Name,
		Token:   request.Data.Token,
		Enabled: request.Data.Enabled,
		Online:  false,
	}); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to create hardware node", Error: "database failure"})
		return
	}
	// Run a hardware health-check afterwards
	if err := hardware.RunNodeCheck(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		Res(w, Response{Success: false, Message: "could not add hardware node", Error: "healthcheck failed: backend failure"})
		return
	}
	Res(w, Response{Success: true, Message: "successfully created hardware node"})
}

// Modifies a hardware node
func ModifyHardwareNode(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request modifyHardwareNodeRequest
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}
	// Validate length of the name
	if utf8.RuneCountInString(request.Data.Name) > 30 {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "bad request", Error: "the name may not exceed 30 characters"})
		return
	}
	// Validate length of the token
	if utf8.RuneCountInString(request.Data.Token) > 100 {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "bad request", Error: "the token may not exceed 100 characters"})
		return
	}
	// Validate that the hardware node exists
	_, exists, err := database.GetHardwareNodeByUrl(request.Url)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to modify hardware node", Error: "database failure"})
		return
	}
	if !exists {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to modify hardware node", Error: "no such node exists"})
		return
	}
	// Modify the node in the database
	if err := database.ModifyHardwareNode(
		request.Url,
		request.Data.Enabled,
		request.Data.Name,
		request.Data.Token,
	); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to modify hardware node", Error: "database failure"})
		return
	}
	Res(w, Response{Success: true, Message: "successfully modified hardware node"})
}

// Deletes a hardware node
func DeleteHardwareNode(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request deleteHardwareNodeRequest
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		Res(w, Response{Success: false, Message: "bad request", Error: "invalid request body"})
		return
	}
	// Validate if the node exists
	_, exists, err := database.GetHardwareNodeByUrl(request.Url)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to delete hardware node", Error: "database failure"})
		return
	}
	if !exists {
		w.WriteHeader(http.StatusUnprocessableEntity)
		Res(w, Response{Success: false, Message: "failed to delete hardware node", Error: "no such node exists"})
		return
	}
	// Delete the node from the database
	if err := database.DeleteHardwareNode(request.Url); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		Res(w, Response{Success: false, Message: "failed to delete hardware node", Error: "database failure"})
		return
	}
	Res(w, Response{Success: true, Message: "successfully deleted hardware node"})
}
