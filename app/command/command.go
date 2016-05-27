/*
The command package is used to namespace commands ran though the
command-line interface.  The Command value is what is used to determine
the top-level CLI commands.  This uses a subcommand structure.
This is used in conjunction with https://github.com/codegangsta/cli
*/
package command

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/codegangsta/cli"

	"github.com/Unified/pmn/lib/config"
	commandLib "github.com/Unified/pmn/lib/commands"
	"github.com/Unified/batch/app/model"
	"os"
	"os/signal"
	"time"
)

var Commands []cli.Command = []cli.Command{
	{
		Name:  "ping",
		Usage: "Commands for checking connections",
		Subcommands: []cli.Command{
			{
				Name:  "neo",
				Usage: "Ping Neo4J",
				Action: func(c *cli.Context) {
					neoLocation := fmt.Sprintf("%s://%s:%s", config.Get("neo4j_protocol"), config.Get("neo4j_host"), config.Get("neo4j_port"))
					log.Printf("Pinging Neo @ %s", neoLocation)

					req, err := http.NewRequest("GET", neoLocation+"/db/data/", nil)
					if err != nil {
						log.Fatal("Error occurred pinging: ", err.Error())
					}

					resp, err := (&http.Client{}).Do(req)
					if err != nil {
						log.Fatal("Error occurred pinging: ", err.Error())
					}

					body, err := ioutil.ReadAll(resp.Body)
					if err != nil {
						log.Fatal("Error reading response from entity manager: ", err.Error())
					}

					log.Print("Response Code: ", resp.StatusCode, "\nResponse Body: ", string(body))
				},
			},
		},
	},
	{
		Name:  "test",
		Usage: "Commands for running tests",
		Subcommands: []cli.Command{
			{
				Name:  "run",
				Usage: "Run the tests at the given path(s)",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "coverage, c",
						Usage: "Generate code coverage for this run. Set to xml, html, coverprofile, or none",
					},
					cli.StringFlag{
						Name:  "nodes, n",
						Usage: "If running this in parallel, the number of nodes to run",
					},
					cli.BoolFlag{
						Name:  "watch, w",
						Usage: "Watch .go files and re-test on change.",
					},
					cli.BoolFlag{
						Name:  "noColor, N",
						Usage: "Do not output color from the test output",
					},
				},
				Action: commandLib.RunTestCommand,
			},
			{
				Name: "coverprofile",
				Flags: []cli.Flag{
					cli.BoolFlag{
						Name:  "list, l",
						Usage: "List all of the coverprofile files in the project",
					},
					cli.BoolFlag{
						Name:  "packages, p",
						Usage: "List all of the packages to cover in the tests",
					},
					cli.BoolFlag{
						Name:  "merge, m",
						Usage: "Merge all of the coverprofile files into a single file",
					},
					cli.BoolFlag{
						Name:  "clear, c",
						Usage: "Remove all of the coverprofile files in the project",
					},
				},
				Usage:  "Utilities for coverprofile files - the files golang generates for code coverage",
				Action: commandLib.CoverprofileCommand,
			},
		},
	},
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
