package main

import (
	"carlosapi/pkg/routes"
	"carlosapi/pkg/config"
	"carlosapi/pkg/controllers"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
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
	const logo = `
___/\/\/\/\/\______/\/\______/\/\/\/\/\____/\/\__________/\/\/\/\______/\/\/\/\/\_
_/\/\____________/\/\/\/\____/\/\____/\/\__/\/\________/\/\____/\/\__/\/\_________
_/\/\__________/\/\____/\/\__/\/\/\/\/\____/\/\________/\/\____/\/\____/\/\/\/\___
_/\/\__________/\/\/\/\/\/\__/\/\__/\/\____/\/\________/\/\____/\/\__________/\/\_
___/\/\/\/\/\__/\/\____/\/\__/\/\____/\/\__/\/\/\/\/\____/\/\/\/\____/\/\/\/\/\___

          Cooperative Amateur RadioTelescope Listening Outer Space

`	
	fmt.Fprintf(os.Stderr, logo)
	log.Printf("📡 CarlosAPI version %s Listening on port %d", conf.Version, conf.Port)
	addr := fmt.Sprintf("%s:%d", conf.Addr, conf.Port) 
	log.Fatal(http.ListenAndServe(addr, router))
}
