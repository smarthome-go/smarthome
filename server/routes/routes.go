package routes

import (
	"net/http"

	"github.com/MikMuellerDev/smarthome/server/middleware"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

func InitLogger(logger *logrus.Logger) {
	log = logger
}

func NewRouter() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/", middleware.AuthRequired(indexGetHandler)).Methods("GET")
	r.HandleFunc("/dash", middleware.AuthRequired(dashGetHandler)).Methods("GET")
	r.HandleFunc("/login", loginGetHandler).Methods("GET")

	//// API ////
	r.HandleFunc("/api/login", loginPostHandler).Methods("POST")

	/// Api / Power ///
	r.HandleFunc("/api/power", powerPostHandler).Methods("POST")

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

	r.NotFoundHandler = http.HandlerFunc(notFoundHandler)

	log.Trace("Initialized Router.")
	return r
}
