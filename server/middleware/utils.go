package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func Res(w http.ResponseWriter, res Response) {
	now := time.Now().Local()
	response := res
	response.Time = fmt.Sprint(now.UnixMilli())
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Error("Could not send response to client: ", err.Error())
		return
	}
}
