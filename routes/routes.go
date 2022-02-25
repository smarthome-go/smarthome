package routes

import (
	"fmt"
	"net/http"

	"github.com/MikMuellerDev/smarthome/middleware"
	"github.com/gorilla/mux"
)

func NewRouter() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/", middleware.AuthRequired(test)).Methods("GET")

	// For JS and CSS components
	outFilepath := "./web/out/"
	staticPathPrefix := "/static"
	outFileserver := http.FileServer(http.Dir(outFilepath))
	r.PathPrefix(staticPathPrefix).Handler(http.StripPrefix(staticPathPrefix, outFileserver))

	// Other assets, such as PNG or JPEG
	assetsFilepath := "./web/assets/"
	assetsPathPrefix := "/assets"
	assetsFileserver := http.FileServer(http.Dir(assetsFilepath))
	r.PathPrefix(assetsPathPrefix).Handler(http.StripPrefix(assetsPathPrefix, assetsFileserver))

	fmt.Println("Initialized Router.")
	return r
}
