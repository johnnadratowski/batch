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

// Starts a bunch of background worker tasks to process asynchronous batch items from kafka/redis
func StartAsyncWorkers(numWorkers int, quit chan bool, finished chan bool) {
	quitWorkers := make([]chan bool, numWorkers)
	finishedWorkers := make([]chan bool, numWorkers)
	for i := 0; i < numWorkers; i++ {
		quitWorkers[i] = make(chan bool, 1)
		finishedWorkers[i] = make(chan bool, 1)
		go StartAsyncWorker(i, quitWorkers[i], finishedWorkers[i])
	}

	<-quit
	go func() {
		for idx, quitWorker := range quitWorkers {
			log.Printf("Quitting worker: %d", idx)
			quitWorker <- true
		}
	}()

	for idx, finishedWorker := range finishedWorkers {
		select {
		case <-finishedWorker:
			log.Println("Worker %d finished successfully.", idx)
		case <-time.After(1 * time.Second):
			log.Println("Worker %d DID NOT finish.", idx)
		}
	}

	finished <- true
}

// Starts a new background worker task to process asynchronous batch items from kafka/redis
func StartAsyncWorker(workerNum int, quit chan bool, finished chan bool) {

	consumer, err := GetAsyncBatchConsumer()
	if err != nil {
		log.Printf("An error occurred connecting to consumer: %s", err)
	}

	redis := GetAsyncJobRedis()

	for {

		select {
		case <-quit:
			log.Println("Interrupt detected on worker: ", workerNum)
			go func(workerNum int) {
				log.Println("Closing Consumer for worker: ", workerNum)
				if err := consumer.Close(); err != nil {
					sarama.Logger.Printf("Error closing the task consumer for worker %d: %s", workerNum, err)
				}
			}(workerNum)

			finished <- true
			return
		case <-time.After(time.Duration(config.GetInt("worker_sleep")) * time.Millisecond):
			consumeMessages(consumer, redis)
		}

	}
}

func consumeMessages(consumer *consumergroup.ConsumerGroup, redis *redis.Client) {

	for message := range consumer.Messages() {
		var batchItem AsyncBatchItem
		err := json.Unmarshal(message.Value, &batchItem)
		if err != nil {
			log.Printf("This shouldn't happen. We put this JSON into the Kafka queue and it should always be properly formatted. [key: %s] [offset: %d] [partition: %d] [topid: %s] [value: %s] (error: %s)",
				message.Key,
				message.Offset,
				message.Partition,
				message.Topic,
				message.Value,
				err)
			continue
		}

		redisCheckCmd := redis.LIndex(batchItem.RequestID, batchItem.Index)
		checkResult, err := redisCheckCmd.Result()
		if err != nil {
			log.Printf("An error occurred getting Redis info. [request ID: %s] [result: %s] (error: %s)", batchItem.RequestID, checkResult, err)
			continue
		}

		if checkResult != "" {
			log.Printf("Batch Item already processed: [request id: %s] [request index: %d] [key: %s] [offset: %d] [partition: %d] [topid: %s] [value: %s]",
				batchItem.RequestID,
				batchItem.Index,
				message.Key,
				message.Offset,
				message.Partition,
				message.Topic,
				message.Value)
			continue
		}

		response, jsonErr := batchItem.Item.RequestItem(batchItem.IdentityID)
		if jsonErr != nil {
			log.Printf("An error occurred requesting batch item: [request id: %s] [key: %s] [offset: %d] [partition: %d] [topid: %s] [value: %s]",
				batchItem.RequestID,
				message.Key,
				message.Offset,
				message.Partition,
				message.Topic,
				message.Value)
			response = BatchResponseItem{
				Code: 500,
				Body: jsonErr,
			}
		}

		responseJson, _ := json.Marshal(response)
		redisPutCmd := redis.LSet(batchItem.RequestID, batchItem.Index, string(responseJson))
		putResult, err := redisPutCmd.Result()
		if err != nil {
			log.Printf("An error occurred putting batch item response into Redis: [request id: %s] [request index: %s] [key: %s] [offset: %d] [partition: %d] [topid: %s] [value: %s] (error: %s)",
				batchItem.RequestID,
				batchItem.Index,
				message.Key,
				message.Offset,
				message.Partition,
				message.Topic,
				message.Value,
				err)
		} else {
			log.Printf("Successfully processed batch item: [request id: %s] [request index: %s] [key: %s] [offset: %d] [partition: %d] [topid: %s] [value: %s] (result: %s)",
				batchItem.RequestID,
				batchItem.Index,
				message.Key,
				message.Offset,
				message.Partition,
				message.Topic,
				message.Value,
				putResult)
		}

		consumer.CommitUpto(message)
	}
}
