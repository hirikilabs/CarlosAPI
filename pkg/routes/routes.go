package routes

import(
	"carlosapi/pkg/controllers"
	"github.com/gorilla/mux"
)

func RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/", controllers.Root).Methods("GET")
	router.HandleFunc("/record", controllers.CreateRecording).Methods("POST")
	router.HandleFunc("/status", controllers.GetStatus).Methods("GET")
	router.HandleFunc("/status/{id}", controllers.GetStatusId).Methods("GET")
	router.HandleFunc("/clear", controllers.ClearDatabase).Methods("GET")
}
