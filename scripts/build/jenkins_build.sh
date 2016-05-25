#!/usr/bin/env bash

# ======================================================================================================================
# CONFIGURATION
# ======================================================================================================================

# The script will terminate after the first line that fails (returns nonzero exit code)
set -e

# Export the Env. Var. needed for go
export GOPATH=$WORKSPACE/go
export PATH=$PATH:$GOPATH/bin

declare UNIFIED_GO_PATH=$GOPATH/src/github.com/Unified
declare BATCH_PATH=${UNIFIED_GO_PATH}/batch

# ======================================================================================================================
# MAIN
# ======================================================================================================================

cd "${BATCH_PATH}"

# Install the required go tools
go get github.com/tools/godep
go install github.com/tools/godep
go get -v github.com/onsi/ginkgo/ginkgo
go get -v github.com/onsi/gomega
go get github.com/axw/gocov/gocov
go get github.com/AlekSi/gocov-xml
go get github.com/golang/lint/golint

# Install godep.json packages into the GOPATH
godep restore

# Build the go project
godep go build

# Install dependencies packages
cd app
go get -v -t ./...


cd "${BATCH_PATH}"

# Unfocus any focused test to enable randomized execution
ginkgo unfocus

# Needed for our jenkins box to support relatives path below
shopt -s globstar

# Run the unit tests
ginkgo -nodes=${GINKGO_PARALLEL_NODES} -stream -r --randomizeAllSpecs --randomizeSuites -coverpkg=`go list ./app/... ./lib/... | grep -v /batch/app/command | tr '\n' ','`  --cover --trace --race --progress ./app ./lib ./test

# Export code coverage report
gocov convert app/controller/*.coverprofile test/**/*.coverprofile lib/**/*.coverprofile | gocov-xml > coverage.xml

# Export the lint checks report
golint ./... > lint.txt