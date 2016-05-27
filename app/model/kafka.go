package model

import (
	"github.com/Shopify/sarama"
	"github.com/wvanbergen/kafka/consumergroup"
	"github.com/wvanbergen/kazoo-go"
	"log"
	"strings"
	"time"
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

// creates the task consumer for the kafka queues
func NewAsyncBatchConsumer(zookeeperConn, consumerGroup, topic string, headOffset int64, resetOffsets bool) (*consumergroup.ConsumerGroup, error) {

	kafkaConfig := consumergroup.NewConfig()
	kafkaConfig.Offsets.Initial = headOffset
	kafkaConfig.Offsets.ProcessingTimeout = 10 * time.Second
	kafkaConfig.Offsets.ResetOffsets = resetOffsets

	var zookeeperNodes []string
	zookeeperNodes, kafkaConfig.Zookeeper.Chroot = kazoo.ParseConnectionString(zookeeperConn)

	consumer, err := consumergroup.JoinConsumerGroup(consumerGroup, strings.Split(topic, ","), zookeeperNodes, kafkaConfig)
	if err != nil {
		log.Println("Error joining task consumergroup: ", err)
		return nil, err
	}

	go func() {
		for err := range consumer.Errors() {
			log.Println("An error was received from the task consumer: ", err)
		}
	}()

	return consumer, nil
}
