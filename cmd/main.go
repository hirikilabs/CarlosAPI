package main

import (
	"carlosapi/pkg/color"
	"carlosapi/pkg/config"
	"carlosapi/pkg/controllers"
	"carlosapi/pkg/routes"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func main() {
	// get config
	conf := config.GetConfig()

	// init
	config.NoRecording()

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
	fmt.Fprintf(os.Stderr, color.Cyan + logo + color.Reset)
	log.Printf("📡 " + color.Green + "CarlosAPI version " + color.Purple + "%s" + color.Green + " listening on port " + color.Yellow + "%d" + color.Reset, conf.Version, conf.Port)

	// run sheduler on another thread
	go controllers.RunScheduling()

	addr := fmt.Sprintf("%s:%d", conf.Addr, conf.Port) 
	log.Fatal(http.ListenAndServe(addr, router))
}
