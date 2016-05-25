#!/bin/sh
# vim: syntax=sh tabstop=8 expandtab shiftwidth=4 softtabstop=4
# Setup go for local development 

#### LINUX ONLY FOR NOW!!!

if [ ! -d ./.git ]; then
    echo "You must push your project from the root, so the hooks run properly. Exiting."
    exit 1
fi

if ! command -v go >/dev/null 2>&1; then
  echo "INSTALLING GO"
  cd /tmp/
  wget -q https://storage.googleapis.com/golang/go1.4.2.linux-amd64.tar.gz
  tar -C /usr/local -xzf go1.4.2.linux-amd64.tar.gz
  ln -s /usr/local/go/bin/go /usr/bin/go
  mkdir -p ~/go/src/github.com/Unified
  ln -s `pwd` ~/go/src/github.com/Unified/batch
  echo "export GOPATH=~/go/" >> ~/.bashrc
  echo "alias watchTest='find ./app ./lib ./test -name \"*.go\" -print | entr ginkgo -r -p --randomizeAllSpecs --randomizeSuites --failOnPending --cover --trace --race --progress '" >> ~/.bashrc
  echo "alias watchServer=\"find ./app ./lib ./test -name \"*.go\" -print | entr -r sh -c \"go build; ./batch\"" >> ~/.bashrc
fi

if ! command -v godep >/dev/null 2>&1; then
  GOPATH=~/go/ go get github.com/tools/godep
  GOPATH=~/go/ godep restore
fi

if ! command -v ginkgo >/dev/null 2>&1; then
  GOPATH=~/go/ go get github.com/onsi/ginkgo/ginkgo
  GOPATH=~/go/ go get github.com/onsi/gomega
fi

