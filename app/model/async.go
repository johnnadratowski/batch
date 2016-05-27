package model

import (
	"log"
	"net/http"
	"time"
	"gopkg.in/redis.v2"
	"github.com/Unified/pmn/lib/config"
	"github.com/Shopify/sarama"
	"github.com/wvanbergen/kafka/consumergroup"
	"encoding/json"
)

// Struct used to write to kafka an asynchronous batch item request
type AsyncBatchItem struct {
	RequestID string `json:"requestId"`
	Index int64 `json:"idx"`
	Item BatchItem `json:"item"`
	IdentityID string `json:"identityId"`
}

// Get a redis instance for the async jobs
var GetAsyncJobRedis = func() (*redis.Client) {
	return NewRedisClient(config.Get("redis_host"),
		config.Get("redis_port"),
		config.Get("redis_password"),
		config.Get("redis_db"))
}

// Get the kafka producer for asynchronous batch requests
var GetAsyncBatchProducer = func() (sarama.SyncProducer, error) {
	return NewAsyncBatchProducer(config.Get("zookeeper"))
}


// Get the kafka consumer for asynchronous batch requests
var GetAsyncBatchConsumer = func() (*consumergroup.ConsumerGroup, error) {
	return NewAsyncBatchConsumer(config.Get("zookeeper"),
		config.Get("consumer_group"),
		config.Get("topic"),
		config.GetInt64("head_offsets"),
		config.IsTrue("reset_offsets", false))
}

// Get the client to use for the http requests
var GetRequestClient = func() (BatchClient) {
	return &http.Client{}
}

// Starts a new background worker task to process asynchronous batch items from kafka/redis
func StartAsyncWorker() {
	for {
		time.Sleep(time.Duration(config.GetInt("worker_sleep")) * time.Millisecond)

		consumer, err := GetAsyncBatchConsumer()
		if err != nil {
			log.Printf("An error occurred connecting to consumer: %s", err)
			continue
		}
		defer consumer.Close()

		redis := GetAsyncJobRedis()

		for message := range consumer.Messages() {
			var batchItem AsyncBatchItem
			err = json.Unmarshal(message.Value, &batchItem)
			if err != nil {
				log.Printf("This shouldn't happen. We put this JSON into the Kafka queue and it should always be properly formatted. [key: %s] [offset: %d] [partition: %d] [topid: %s] [value: %s]",
					message.Key,
					message.Offset,
					message.Partition,
					message.Topic,
					message.Value)
				continue
			}

			redisCmd := redis.LIndex(batchItem.RequestID, batchItem.Index)
			result, err := redisCmd.Result()
			if err != nil {
				log.Printf("An error occurred getting Redis info for request ID %s. [error: %s]", batchItem.RequestID, err)
				continue
			}

			if result != "" {
				log.Printf("Batch Item already processed: [key: %s] [offset: %d] [partition: %d] [topid: %s] [value: %s]",
					message.Key,
					message.Offset,
					message.Partition,
					message.Topic,
					message.Value)
				continue
			}

			batchItem.Item.RequestItem(batchItem.IdentityID)
		}
	}
}