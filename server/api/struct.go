package api

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// A generic response used in most routes
// swagger:response genericResponse
type Response struct {
	// Status of the request
	// in: bool
	Success bool `json:"success"`
	// A summary of the request's outcome
	// in: string
	Message string `json:"message"`
	// If the request was unsuccessfull, the error will describe what went wrong
	// in: string
	Error string `json:"error"`
	// A timestamp which indicates when the server was ready and sent the request
	// in: string
	Time string `json:"time"`
}
