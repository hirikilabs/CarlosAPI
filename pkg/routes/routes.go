package routes

import (
	"carlosapi/pkg/controllers"
	"net/http"

	"github.com/gorilla/mux"
)

func RegisterRoutes(router *mux.Router) {
	// API
	router.HandleFunc("/api", controllers.ApiRoot).Methods("GET")
	router.HandleFunc("/api/record", controllers.ApiCreateRecording).Methods("POST")
	router.HandleFunc("/api/status", controllers.ApiGetStatus).Methods("GET")
	router.HandleFunc("/api/status/{id}", controllers.ApiGetStatusId).Methods("GET")
	router.HandleFunc("/api/clear", controllers.ApiClearDatabase).Methods("GET")
	router.HandleFunc("/api/download/{id}", controllers.ApiDownloadId).Methods("GET")
	// Web
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))
	router.HandleFunc("/", controllers.WebRoot).Methods("GET")
	router.HandleFunc("/request", controllers.WebMakeRequest).Methods("GET")
	router.HandleFunc("/record", controllers.WebCreateRecording).Methods("POST")
	router.HandleFunc("/info/{id}", controllers.WebInfo).Methods("GET")
}
