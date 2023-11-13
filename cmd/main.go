package main

import (
	"carlosapi/pkg/routes"
	"carlosapi/pkg/config"
	"carlosapi/pkg/controllers"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

func main() {
	// get config
	conf := config.GetConfig()

	// init
	config.NoRecording()

	// run sheduler on another thread
	go controllers.RunScheduling()
	
	// create HTTP routes
	router := mux.NewRouter()
	routes.RegisterRoutes(router)
	http.Handle("/", router)

	// launch server
	log.Printf("📡 CarlosAPI version %s Listening on port %d", conf.Version, conf.Port)
	addr := fmt.Sprintf("%s:%d", conf.Addr, conf.Port) 
	log.Fatal(http.ListenAndServe(addr, router))
}
