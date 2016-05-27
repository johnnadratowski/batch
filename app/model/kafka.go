package model

import (
	"github.com/Shopify/sarama"
	"github.com/wvanbergen/kazoo-go"
	"github.com/Unified/pmn/lib/config"
	"time"
	"log"
)

// NewAsyncBatchProducer gets a new producer instance for producing batch item messages to kafka
// Can be overridden for mocking kafka
var NewAsyncBatchProducer = func(zookeeperConn string) (sarama.SyncProducer, error) {
	zookeeper, err := kazoo.NewKazooFromConnectionString(zookeeperConn, kazoo.NewConfig())
	if err != nil {
		log.Fatalln("An error occurred connecting to zookeeper", err)
		return nil, err
	}

	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 3
	config.Producer.Retry.Backoff = 100 * time.Millisecond
	// TODO: This is a reminder to try profiling against batching these vs sending them 1-by-1
	//config.Producer.Flush.Messages = 10000
	//config.Producer.Flush.MaxMessages = 10000

	brokerList, err := zookeeper.BrokerList()
	if err != nil {
		log.Fatalln("An error occurred getting broker list from zookeeper", err)
		return nil, err
	}

	producer, err := sarama.NewSyncProducer(brokerList, config)
	if err != nil {
		log.Fatalln("Failed to start Sarama producer:", err)
		return nil, err
	}

	return producer, nil
}
