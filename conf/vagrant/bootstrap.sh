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
######### KAFKA/ZOOKEEPER
###############################################

if ! command -v java >/dev/null 2>&1; then
  echo "INSTALL JAVA/SCALA"
  apt-get -y install scala
fi

if [ ! -e /srv/unified/kafka/bin/zookeeper-server-start.sh ]; then
    mkdir -p /srv/unified/kafka
    cd /srv/unified/kafka
    wget -q http://apache.claz.org/kafka/0.10.0.0/kafka_2.11-0.10.0.0.tgz
    tar xzf kafka_2.11-0.10.0.0.tgz
    cd kafka_2.11-0.10.0.0/
    echo "listeners=PLAINTEXT://192.168.50.7:9092" >> config/server.properties
    echo "advertised.listeners=PLAINTEXT://192.168.50.7:9092" >> config/server.properties
fi

./bin/zookeeper-server-start.sh -daemon config/zookeeper.properties
./bin/kafka-server-start.sh -daemon config/server.properties


###############################################
######### REDIS
###############################################

if ! command -v redis-server >/dev/null 2>&1; then
  echo "INSTALL REDIS"
  apt-get -y install redis-server
  echo "bind 192.168.50.7" >> /etc/redis/redis.conf
fi

/etc/init.d/redis-server restart

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
