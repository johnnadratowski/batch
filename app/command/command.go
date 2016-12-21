/*
The command package is used to namespace commands ran though the
command-line interface.  The Command value is what is used to determine
the top-level CLI commands.  This uses a subcommand structure.
This is used in conjunction with https://github.com/codegangsta/cli
*/
package command

import (
	"log"

	"github.com/codegangsta/cli"

	"os"
	"os/signal"
	"time"

	"github.com/johnnadratowski/batch/app/model"
)

var Commands []cli.Command = []cli.Command{
	{
		Name:  "worker",
		Usage: "Start background async batch request workers",
		Flags: []cli.Flag{
			cli.IntFlag{
				Name:  "workers, w",
				Usage: "The number of workers to start",
				Value: 1,
			},
		},
		Action: func(c *cli.Context) {
			catchSignal := make(chan os.Signal, 1)
			quit := make(chan bool, 1)
			finished := make(chan bool, 1)
			signal.Notify(catchSignal, os.Interrupt)
			go model.StartAsyncWorkers(c.Int("workers"), quit, finished)
			<-catchSignal
			log.Println("Caught interrupt signal. Waiting for workers to finish.")
			quit <- true

			select {
			case <-finished:
				log.Println("All workers finished. Shutdown successfully")
			case <-time.After(3 * time.Second):
				log.Println("All workers DID NOT finished. Forcefully shutting down")
			}
		},
	},
}
