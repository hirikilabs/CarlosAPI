package routes

import (
	"carlosapi/pkg/controllers"
	"net/http"

	"github.com/gorilla/mux"
)

func RegisterRoutes(router *mux.Router) {
	// API
	router.HandleFunc("/", controllers.Root).Methods("GET")
	router.HandleFunc("/record", controllers.CreateRecording).Methods("POST")
	router.HandleFunc("/status", controllers.GetStatus).Methods("GET")
	router.HandleFunc("/status/{id}", controllers.GetStatusId).Methods("GET")
	router.HandleFunc("/clear", controllers.ClearDatabase).Methods("GET")
	router.HandleFunc("/download/{id}", controllers.DownloadId).Methods("GET")
	// Web
	router.PathPrefix("/static").Handler(http.FileServer(http.Dir("./static/")))
	router.HandleFunc("/request", controllers.MakeRequest).Methods("GET")
}
