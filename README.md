# Batch

**NOTE: This was an experiment and is not 100% production ready.  If someone is interested in me finishing this project, just open a ticket.**

Microservice implementing batch commands for queries to other microservices.  You can do synchronous batches and asynchronous batches.  For asynchronous batches, you can poll to see if its done.  Utilizes Redis and Kafka to handle the asynchronous batching.

# Quick Start


```sh
# INSTALL HOOKS 
./conf/setup/setup-hooks.sh

# BUILDING PROJECT
go build

# RUNNING PROJECT
./batch
```

# Configuration

** NOTE: STILL HAVE TO SET UP CONFIG MECHANISM**
Configuration is done solely through env variables.  You can easily see what configuration values the system supports by looking at app/config.go.  However, whenever a new configuration value is added, it should be documented here, with it's default value.

```sh
ENV=dev # The current environment type

# Server Configs
HOST=localhost # Hostname to bind webserver to
PORT=8087 # Port to bind the webserver to
READ_TIMEOUT=300 # maximum duration before timing out read of the request (in seconds)
WRITE_TIMEOUT=300 # maximum duration before timing out write of the response (in seconds)
MAX_HEADER_BYTES=0 # maximum size of request headers, 1 MB if 0

# Batch Configs
MAX_BATCH_REQUESTS=100 # Max number of requests a user can make in a single call
MAX_BATCH_ASYNC_REQUESTS=10000 # Max number of requests a user can make in a single call

# Batch Host Configs
# These are dynamically read by the application. They use a naming scheme to determine the host identifier.  You can add as many of these as you want and batch will be able to communicate with those services
# This takes the format shown below. Example: PMN_BATCH_HOST=http://pmn.loadbalancer.unified.com:80
(SERVICE_ID)_BATCH_HOST=http://(host):(port)

# Zookeeper/Kafka Configs
ZOOKEEPER=localhost:2181 # The connection string to the zookeeper node(s)
TOPIC=batch_async # The kafka topic to use for async calls

# Redis
REDIS_HOST=localhost # The host that Redis is running on
REDIS_PORT=6379 # The port that Redis is running on
REDIS_DB=0 # The Redis db to connect to
REDIS_PASSWORD= # The password to use to connect to Redis
ASYNC_EXPIRE=60 # Expiration time for new async request, in minutes

# Workers
WORKERS=0 # The number of async workers to start with the webserver
WORKER_SLEEP=500 # Number of milliseconds to sleep between worker processing
HEAD_OFFSETS=-2 # Set to the offset to start at. Defaults to the oldest offset for the consumer group. Set to -1 to start at the newest offset for the group
RESET_OFFSETS=false # Set this to true to reset the offsets for the consumer group
CONSUMER_GROUP=batch_async # the consumer group to use for the worker
```
