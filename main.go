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
	"strings"
	"github.com/Shopify/sarama"
)

// Initializes the applications configuration
func InitializeConfig() {
	conf := map[string]string{
		"env": config.EnvDefault("ENV", "dev"),

		// Webserver
		"host":             config.EnvDefault("HOST", "localhost"),
		"port":             config.EnvDefault("PORT", "8080"),
		"read_timeout":     config.EnvDefault("READ_TIMEOUT", "300"),
		"write_timeout":    config.EnvDefault("WRITE_TIMEOUT", "300"),
		"max_header_bytes": config.EnvDefault("MAX_HEADER_BYTES", "0"),

		// Batch options
		"max_batch_requests":     config.EnvDefault("MAX_BATCH_REQUESTS", "100"),
		"max_batch_requests_async":     config.EnvDefault("MAX_BATCH_REQUESTS_ASYNC", "10000"),

		// Zookeeper/Kafka options
		"zookeeper": config.EnvDefault("ZOOKEEPER", "localhost:2181"),
		"topic": config.EnvDefault("TOPIC", "batch_asyc"),

		// Redis
		"redis_host":     config.EnvDefault("REDIS_HOST", "localhost"),
		"redis_port":     config.EnvDefault("REDIS_PORT", "6379"),
		"redis_db":       config.EnvDefault("REDIS_DB", "0"),
		"redis_password": config.EnvDefault("REDIS_PASSWORD", ""),
		"async_expire": config.EnvDefault("ASYNC_EXPIRE", "60"),

		// Workers
		"workers":     config.EnvDefault("WORKERS", "0"),
		"worker_sleep":     config.EnvDefault("WORKER_SLEEP", "500"),
		"head_offsets": config.EnvDefault("HEAD_OFFSETS", "-1"),
		"reset_offsets": config.EnvDefault("RESET_OFFSETS", "false"),
		"consumer_group": config.EnvDefault("CONSUMER_GROUP", "batch_async"),
	}

	for _, envVar := range os.Environ() {
		parts := strings.SplitN(envVar, "=", 1)
		envKey := parts[0]
		envVal := parts[1]

		if strings.HasSuffix(envKey, "_BATCH_HOST") {
			conf[strings.ToLower(envKey)] = envVal
		}
	}

	config.Initialize(conf)
}

// ConfigureServer sets extra server configuration
func ConfigureServer() {
	model.HostMap = map[string]string{}
	for key, val := range config.All() {
		if !strings.HasSuffix(key, "_batch_host") {
			continue
		}

		parts := strings.SplitN(key, "_", 1)
		model.HostMap[parts[0]] = val
	}

	sarama.Logger = log.New(os.Stderr, "[Sarama] ", log.LstdFlags)
}

func main() {
	InitializeConfig()

	if len(os.Args) == 1 {
		log.Println("Starting Batch Server")

		config.Log()

		listen := fmt.Sprintf("%s:%s", config.Get("host"), config.Get("port"))

		ConfigureServer()

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
