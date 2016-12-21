package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/johnnadratowski/batch/app/command"
	"github.com/johnnadratowski/batch/app/model"
	"github.com/johnnadratowski/batch/app/route"
	"github.com/johnnadratowski/batch/app/server"
	"github.com/codegangsta/cli"
)

var HOST string = ""
var PORT string = ""
var WORKERS int = 10

func main() {
	server.InitializeConfig()

	server.ConfigureServer()

	if len(os.Args) == 1 {
		log.Println("Starting Batch Server")

		listen := fmt.Sprintf("%s:%s", HOST, PORT)

		server := &http.Server{
			Addr:           listen,
			Handler:        route.Router(),
		}

		quit := make(chan bool, 1)
		finished := make(chan bool, 1)
		numWorkers := WORKERS
		if numWorkers > 0 {
			go model.StartAsyncWorkers(numWorkers, quit, finished)
		}

		err := server.ListenAndServe()
		log.Printf("Exiting Batch Server: %s", err)

		if numWorkers > 0 {
			log.Println("Quitting workers")
			quit <- true

			select {
			case <-finished:
				log.Println("All workers finished. Shutdown successfully")
			case <-time.After(3 * time.Second):
				log.Println("All workers DID NOT finished. Forcefully shutting down")
			}
		}
	} else {
		app := cli.NewApp()
		app.Name = "Batch"
		app.Usage = "Makes batch calls"
		app.EnableBashCompletion = true // TODO: Figure out how to get this to work
		app.Commands = command.Commands
		app.Run(os.Args)
	}
}
