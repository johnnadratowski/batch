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

## Folder Structure

* conf - Contains configuration files
* Godeps - The dependency files saved by the godep tool
* app - Contains all of the golang code used in the main application
    * command - Contains the command-line commands
    * context - Contains all of the different Context objects
    * controller - Contains all of the controller objects
    * middleware - Contains all of the middleware objects
    * model - Contains all of the model objects
    * route - Contains the routes for the application
* test - contains all application tests
    * common - contains common utilites for integration tests
    * integration - contains all integration tests
* main.go - Contains the main entry point to the application
* Dockerfile - used to bootstrap a Docker instance for deploying Batch


## Code Documentation

We will be using Go's built in documentation functionality to do our internal developer documentation.

* All public members should have a docstring associated to them.
* All packages should have a doc.go file that documents the package, unless the packages is one file

There is also additional developer documentation in the docs/dev folder

### Godoc

Go doc generates an HTML website with your documentation in a searchable, nice format using the godoc tool.  It also allows for seaching the documentation from the command line

# Logging

Engineers should use the logging mechanism provided by Go's "log" library.  It is not necessary to instantiate a new logger, and for the most part, the engineer should use the methods in the "log" package.

This will log out the messages to stderr.  That should then we piped to a log aggregation utility.


# Commands

If the Batch executable is ran with command-line arguments, it goes into CLI mode.  This allows for the engineer to create management commands for the Batch.  These commands can do things such as:

* Aiding in debugging
* Aiding in logging
* Work directly with the database
* Pinging dependant services

We use a similar structure to the go tools or to git in our command line tool.  It will use sub commands for different command types.  

Run it with `--help` to see what commands are available.


# Testing

* We are using a test first approach wherever possible.  Please write your tests for your functionality first.
* We are using BDD style testing with ginkgo and gomega
* Batch implements commands to help run the tests
* You can see all of the test commands by running
    * `./batch test --help`
* You can also run the tests by using the `ginkgo` command directly:
    * `ginkgo -r -p --randomizeAllSpecs --randomizeSuites --failOnPending --cover --trace --race --progress`
* You can watch and re-run tests using `./batch test all --watch`
* [This](http://onsi.github.io/ginkgo/#running-tests) Shows you how to run the test suite manually

## Unit Tests

// TODO


## Integration Tests

// TODO

## Code Coverage

* You can generate code coverage using go tooling by running:
    * ``ginkgo -coverpkg=`go list ./app/... ./lib/... | grep -v /batch/app/command | tr '\n' ','` -r ./test ./app``
* You can merge the cover profiles by using the [gocovmerge](https://github.com/wadey/gocovmerge) tool:
    * `gocovmerge **/*.coverprofile > /tmp/total_coverage`
* You can view the code coverage html heatmap by running:
    * `go tool cover -html=/tmp/total_coverage`
* Running this command will run all of the tests, and generate code coverage, opening it automatically in your browser:
    * ``rm -rf **/*.coverprofile; NEO4J_TEST_RUN=false ginkgo -coverpkg=`go list ./app/... ./lib/... | grep -v /batch/app/command | tr '\n' ','` -r ./test ./app ./lib; gocovmerge **/*.coverprofile > /tmp/total_coverage; go tool cover -html=/tmp/total_coverage``
    * It’s important to clear out the coverprofiles every time
* If you have entr, you can run this command to watch go files and re-run all of the tests on change, producing code coverage output:
    * ``find . -name "*.go" | entr -r zsh -c "rm -rf **/*.coverprofile; NEO4J_TEST_RUN=false ginkgo -coverpkg=`go list ./app/... ./lib/... | grep -v /batch/app/command | tr '\n' ','` -r ./test ./app ./lib; gocovmerge **/*.coverprofile > /tmp/total_coverage; go tool cover -html=/tmp/total_coverage"``


## Benchmarks

TODO


## Ginkgo

We are using [Ginkgo (w/Gomega)](https://github.com/onsi/ginkgo) to do BDD testing.

To generate a test suite for a package navigate to that package's folder and run: 
```
ginkgo bootstrap
```

A test suite is the entry point for running the associated test modules in its folder, this is also where you should do global test setup & teardown. For example, in the model folder we have a test suite file, [model_test_suite.go](https://github.com/Unified/batch/blob/master/app/model/model_suite_test.go) that contains BeforeSuite & AfterSuite blocks. The BeforeSuite block opens a new database connection to the Neo4j database, while the AfterSuite block clears the database & closes the connection.

To run all the tests, navigate to the top-level app folder and run:
```
ginkgo -r
```

This will recursively search for all test suites & run their associated test modules. It is also recommended that you run the tests with the -v flag for verbose output. To run an individual test suite, navigate to that test suite's folder and run:
```
ginkgo
```

To run specific tests you have the option of either marking those tests as focused, skipping them, or marking them as pending. You can mark these at either the Describe, Context, or It block levels as shown below:

###### Parallel Tests

Ginkgo is set up to be able to run the tests in parallel.  In order to do this, you may pass the -p flag like so:
```
ginkgo -r -p .
```

That command, ran from the top level, will run all of the tests in the codebase in parallel.  You do not need to worry about data pollution from tests in the database, as each parallel process will run it's own instance of neo4j.


###### Focused Test - Only run this test block.
```
FIt('description', func() {
   Expect('actual').To('expectation')
})
```

###### Skipped Describe Container - Skip all tests within this Describe block.
```
XDescribe('function', func() {
   'Nested Describes, Contexts, & It blocks here'
})
```

###### Pending Context Container - Mark everything within this Context block as pending for work to be done later.
```
PContext('description', func() {
   'Nested Contexts & It blocks here'
})
```

For further information on Ginkgo, please read the [documentation](https://github.com/onsi/ginkgo).


# Docker

Docker is used for deployment and development containment.  It is necessary for elasticbeanstalk.

1. Install Docker
2. Set up environment as laid out in the Quick Start Section
3. cd $GOPATH/src/github.com/Unified/batch
3. Build Docker Image
    * docker build -t batch .
4. Run Docker Image
    * docker run --expose 8080 -t -e PORT=8080 -e HOST=0.0.0.0 -p 8080:8080 batch

You can get a shell into an image by running `docker run --entrypoint="/bin/bash" -t -i batch`


# Vagrant

There is a Vagrantfile in /conf that you can use for vagrant machines.

# Golang

We have chosen Go for this project for the following reasons:

* The safety provided by the type system
* The concurrency primitives
* The single-binary output
* The simple syntax and structure

## Important Libraries

* [Go-CQ](https://github.com/go-cq/cq): We use Go-CQ to interface with the Neo4J graph DB
* [Ginkgo (w/Gomega)](https://github.com/onsi/ginkgo): We use Ginkgo and Gomega for our unit testing
* [Gocraft/Web](https://github.com/gocraft/web): Simple HTTP Route Muxer w/Middleware
* [codegangsta/cli](https://github.com/codegangsta/cli): Used to create the command line interface
* [StatsD](https://github.com/quipo/statsd): Client for StatsD
* [uuid](https://github.com/pborman/uuid): Library for generating UUIDs
* [Validator](gopkg.in/validator.v2): Library for validating models




## Dependency Management

We use [Godep](https://github.com/tools/godep) to manage the dependencies for the Batch project.  Godep allows us to freeze the dependencies for the project and manage the dependencies in the GOPATH.


## Learning Golang

The following links provide a good starting point for learning Go:

* [Golang Playground](http://play.golang.org/)
* [Effective Go](https://golang.org/doc/effective_go.html)
* [How To Write Go Code](https://golang.org/doc/code.html)
* [Using Go With Docker](http://blog.golang.org/docker)


## IDEs

* Visual Studio Code has really good support
* IntelliJ products have a great golang plugin

#### Sublime Text
Sublime Text has a good plugin for go called [GoSublime](https://github.com/DisposaBoy/GoSublime).  In order to get it to work, I had to set the env setting for GOPATH, GOBIN, GOROOT, and PATH in the GoSublime plugin user settings.

For working in Sublime Text, I would also recommend:

* [Sublime Package Control](https://sublime.wbond.net/installation).
* [SidebarEnhancements](https://github.com/titoBouzout/SideBarEnhancements) 


#### Atom
Atom has several packages for Go, including one created by the Atom team named [language-go](https://atom.io/packages/language-go). As with Sublime Text, some packages require you to set the env settings.

For working in Atom, I would recommend the following packages:
* [Ginko & Gomega Snippets](https://atom.io/packages/ginkgo-and-gomega-snippets) - Code snippets for BDD testing in Golang with Ginkgo & Gomega.
* [GoCode](https://atom.io/packages/gocode) - More Golang autocompletions.
* [GoFormat](https://atom.io/packages/go-format) - Automatically runs go fmt on Go source files when saved.
* [GoPlayground](https://atom.io/packages/go-playground) - [Go Playground](http://play.golang.org) functionality in the Atom editor.
* [GoPlus](https://atom.io/packages/go-plus) - Improved Go Experience In Atom, several features, highly recommended!

Go also has [quite a few](https://code.google.com/p/go-wiki/wiki/IDEsAndTextEditorPlugins) different IDEs and IDE integrations.


