package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Unified/batch/app/command"
	"github.com/Unified/batch/app/model"
	"github.com/Unified/batch/app/route"
	"github.com/Unified/pmn/lib/config"
	"github.com/codegangsta/cli"
	"github.com/Unified/batch/app/server"
)

func main() {
	server.InitializeConfig()

	server.ConfigureServer()

	if len(os.Args) == 1 {
		log.Println("Starting Batch Server")

		config.Log()

		listen := fmt.Sprintf("%s:%s", config.Get("host"), config.Get("port"))

		server := &http.Server{
			Addr:           listen,
			ReadTimeout:    time.Duration(config.GetInt("read_timeout")) * time.Second,
			WriteTimeout:   time.Duration(config.GetInt("write_timeout")) * time.Second,
			MaxHeaderBytes: config.GetInt("max_header_bytes"),
			Handler:        route.Router(),
		}

		quit := make(chan bool, 1)
		finished := make(chan bool, 1)
		numWorkers := config.GetInt("workers")
		if numWorkers > 0 {
			model.StartAsyncWorkers(numWorkers, quit, finished)
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
