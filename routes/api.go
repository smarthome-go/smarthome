package routes

import (
	"fmt"
	"net/http"
)

func test(w http.ResponseWriter, h *http.Request) {
	fmt.Fprint(w, "It works.")
}
