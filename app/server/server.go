package server

import (
	"os"
	"strings"
	"github.com/Unified/pmn/lib/config"
	"github.com/Unified/batch/app/model"
	"github.com/Shopify/sarama"
	"log"
)

// Initializes the applications configuration
func InitializeConfig() {
	conf := map[string]string{
		"env": config.EnvDefault("ENV", "dev"),

		// Webserver
		"host":             config.EnvDefault("HOST", "localhost"),
		"port":             config.EnvDefault("PORT", "8087"),
		"read_timeout":     config.EnvDefault("READ_TIMEOUT", "300"),
		"write_timeout":    config.EnvDefault("WRITE_TIMEOUT", "300"),
		"max_header_bytes": config.EnvDefault("MAX_HEADER_BYTES", "0"),

		// Batch options
		"max_batch_requests":     config.EnvDefault("MAX_BATCH_REQUESTS", "100"),
		"max_batch_requests_async":     config.EnvDefault("MAX_BATCH_REQUESTS_ASYNC", "10000"),

		// Zookeeper/Kafka options
		"zookeeper": config.EnvDefault("ZOOKEEPER", "localhost:2181"),
		"topic": config.EnvDefault("TOPIC", "batch_async"),

		// Redis
		"redis_host":     config.EnvDefault("REDIS_HOST", "localhost"),
		"redis_port":     config.EnvDefault("REDIS_PORT", "6379"),
		"redis_db":       config.EnvDefault("REDIS_DB", "0"),
		"redis_password": config.EnvDefault("REDIS_PASSWORD", ""),
		"async_expire": config.EnvDefault("ASYNC_EXPIRE", "60"),

		// Workers
		"workers":     config.EnvDefault("WORKERS", "0"),
		"worker_sleep":     config.EnvDefault("WORKER_SLEEP", "500"),
		"head_offsets": config.EnvDefault("HEAD_OFFSETS", "-2"),
		"reset_offsets": config.EnvDefault("RESET_OFFSETS", "false"),
		"consumer_group": config.EnvDefault("CONSUMER_GROUP", "batch_async"),
	}

	for _, envVar := range os.Environ() {
		parts := strings.SplitN(envVar, "=", 2)
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

		parts := strings.SplitN(key, "_", 2)
		model.HostMap[parts[0]] = val
	}

	sarama.Logger = log.New(os.Stderr, "[Sarama] ", log.LstdFlags)
}
