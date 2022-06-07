package api

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// A generic return value for indicating the result of a request
type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error"`
	Time    string `json:"time"`
}
