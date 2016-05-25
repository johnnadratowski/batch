#!/bin/bash

##############################################
######### SETUP
##############################################

if ! command -v supervisorctl >/dev/null 2>&1; then
  echo "INSTALL SUPERVISOR"
  apt-get update
  apt-get -y install supervisor
fi

if ! command -v git >/dev/null 2>&1; then
  echo "INSTALL GIT"
  apt-get -y install git
fi

if ! command -v hg >/dev/null 2>&1; then
  echo "INSTALL MERCURIAL"
  apt-get -y install mercurial
fi

if [ -h "/etc/supervisor/conf.d/batch.conf" ]; then
    echo "REMOVING SYMLINK BATCH SUPERVISOR"
    rm /etc/supervisor/conf.d/batch.conf
fi

if [ -e "/srv/unified/batch/conf/vagrant/supervisor.conf-local" ]; then
    echo "CREATING SYMLINK BATCH SUPERVISOR TO LOCAL SUPERVISOR CONF"
    ln -s /srv/unified/batch/conf/vagrant/supervisor.conf-local /etc/supervisor/conf.d/batch.conf
else
    echo "CREATING SYMLINK BATCH SUPERVISOR TO SHARED SUPERVISOR CONF"
    ln -s /srv/unified/batch/conf/vagrant/supervisor.conf /etc/supervisor/conf.d/batch.conf
fi

###############################################
######### BUILD STEPS
###############################################

if ! command -v go >/dev/null 2>&1; then
  echo "INSTALLING GO"
  cd /tmp/
  wget -q https://storage.googleapis.com/golang/go1.4.2.linux-amd64.tar.gz
  tar -C /usr/local -xzf go1.4.2.linux-amd64.tar.gz
  ln -s /usr/local/go/bin/go /usr/bin/go
  mkdir -p /go/src/github.com/Unified
  ln -s /srv/unified/batch/ /go/src/github.com/Unified/batch
  GOPATH=/go/ go get github.com/tools/godep
  GOPATH=/go/ go get github.com/onsi/ginkgo/ginkgo
  GOPATH=/go/ go get github.com/onsi/gomega

  cd /go/src/github.com/Unified/batch
  GOPATH=/go/ /go/bin/godep restore

  chown -R vagrant:vagrant /go/
  chown -R vagrant:vagrant /srv/

  echo "cd /go/src/github.com/Unified/batch" >> /root/.bashrc
fi

echo "BUILDING BATCH"
cd /go/src/github.com/Unified/batch
GOPATH=/go/ /go/bin/godep go build

###############################################
######### RUN
###############################################

echo "RUNNING BATCH"
supervisorctl update
supervisorctl restart batch
