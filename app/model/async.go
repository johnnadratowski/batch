package model

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/Shopify/sarama"
	"github.com/wvanbergen/kafka/consumergroup"
	"gopkg.in/redis.v2"
)

var REDIS_HOST string = "default"
var REDIS_PORT string = "default"
var REDIS_PW string = "default"
var REDIS_DB int = 0
var ZOOKEEPER string = "default"
var CONSUMERGROUP string = "default"
var TOPIC string = "default"
var HEAD_OFFSETS int64 = 0
var RESET_OFFSETS bool = false
var WORKER_SLEEP int = 100
var ASYNC_EXPIRE int = 1000

// Struct used to write to kafka an asynchronous batch item request
type AsyncBatchItem struct {
	RequestID  string    `json:"requestId"`
	Index      int64     `json:"idx"`
	Item       BatchItem `json:"item"`
	IdentityID string    `json:"identityId"`
}

// Get a redis instance for the async jobs
var GetAsyncJobRedis = func() *redis.Client {
	return NewRedisClient(REDIS_HOST,
		REDIS_PORT,
		REDIS_PW,
		strconv.Itoa(REDIS_DB))
}

// Get the kafka producer for asynchronous batch requests
var GetAsyncBatchProducer = func() (sarama.SyncProducer, error) {
	return NewAsyncBatchProducer(ZOOKEEPER)
}

// Get the kafka consumer for asynchronous batch requests
var GetAsyncBatchConsumer = func() (*consumergroup.ConsumerGroup, error) {
	return NewAsyncBatchConsumer(ZOOKEEPER,
		CONSUMERGROUP,
		TOPIC,
		HEAD_OFFSETS,
		RESET_OFFSETS)
}

// Get the client to use for the http requests
var GetRequestClient = func() BatchClient {
	return &http.Client{}
}

// Starts a bunch of background worker tasks to process asynchronous batch items from kafka/redis
func StartAsyncWorkers(numWorkers int, quit chan bool, finished chan bool) {
	quitWorkers := make([]chan bool, numWorkers)
	finishedWorkers := make([]chan bool, numWorkers)
	for i := 0; i < numWorkers; i++ {
		log.Printf("Starting worker: %d", i)
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
			log.Printf("Worker %d finished successfully.", idx)
		case <-time.After(1 * time.Second):
			log.Printf("Worker %d DID NOT finish.", idx)
		}
	}

	finished <- true
}

// Starts a new background worker task to process asynchronous batch items from kafka/redis
func StartAsyncWorker(workerNum int, quit chan bool, finished chan bool) {
	sleepDuration := time.Duration(WORKER_SLEEP) * time.Millisecond

	// Keep worker alive if it dies
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Worker %d died. Restarting in 1 seconds. (error: %s)", workerNum, r)
			time.Sleep(sleepDuration)
			StartAsyncWorker(workerNum, quit, finished)
		}
	}()

	redis := GetAsyncJobRedis()
	consumer, err := GetAsyncBatchConsumer()
	if err != nil {
		log.Printf("An error occurred connecting to consumer: %s", err)
	}

	log.Printf("Worker started: %d", workerNum)

	for {
		time.Sleep(sleepDuration)

		select {
		case <-quit:
			log.Println("Interrupt detected. Closing Consumer on worker: ", workerNum)
			if err := consumer.Close(); err != nil {
				sarama.Logger.Printf("Error closing the consumer for worker %d: %s", workerNum, err)
			}
			finished <- true
			return
		case message := <-consumer.Messages():
			processMessage(message, consumer, redis)
			consumer.CommitUpto(message)
		case <-time.After(sleepDuration):
		}
	}
}

// Process a single consumer message
func processMessage(message *sarama.ConsumerMessage, consumer *consumergroup.ConsumerGroup, redis *redis.Client) {

	log.Printf("Got message: [key: %s] [offset: %d] [partition: %d] [topid: %s] [value: %s]", message.Key, message.Offset, message.Partition, message.Topic, message.Value)

	var batchItem AsyncBatchItem
	err := json.Unmarshal(message.Value, &batchItem)
	if err != nil {
		log.Printf("This shouldn't happen. We put this JSON into the Kafka queue and it should always be properly formatted. [key: %s] [offset: %d] [partition: %d] [topid: %s] [value: %s] (error: %s)", message.Key, message.Offset, message.Partition, message.Topic, message.Value, err)
		return
	}

	redisCheckCmd := redis.LIndex(batchItem.RequestID, batchItem.Index)
	checkResult, err := redisCheckCmd.Result()
	if err != nil {
		log.Printf("An error occurred getting Redis info. [request ID: %s] [result: %s] (error: %s)", batchItem.RequestID, checkResult, err)
		return
	}

	if checkResult != "" {
		log.Printf("Batch Item already processed: [request id: %s] [request index: %d] [key: %s] [offset: %d] [partition: %d] [topid: %s] [value: %s]", batchItem.RequestID, batchItem.Index, message.Key, message.Offset, message.Partition, message.Topic, message.Value)
		return
	}

	response, err := batchItem.Item.RequestItem(batchItem.IdentityID)
	if err != nil {
		log.Printf("An error occurred requesting batch item: [request id: %s] [key: %s] [offset: %d] [partition: %d] [topid: %s] [value: %s]", batchItem.RequestID, message.Key, message.Offset, message.Partition, message.Topic, message.Value)
		response = BatchResponseItem{
			Code: 500,
			Body: err,
		}
	}

	responseJson, _ := json.Marshal(response)
	redisPutCmd := redis.LSet(batchItem.RequestID, batchItem.Index, string(responseJson))
	putResult, err := redisPutCmd.Result()
	if err != nil {
		log.Printf("An error occurred putting batch item response into Redis: [request id: %s] [request index: %d] [key: %s] [offset: %d] [partition: %d] [topid: %s] [value: %s] (error: %s)", batchItem.RequestID, batchItem.Index, message.Key, message.Offset, message.Partition, message.Topic, message.Value, err)
	} else {
		log.Printf("Successfully processed batch item: [request id: %s] [request index: %d] [key: %s] [offset: %d] [partition: %d] [topid: %s] [value: %s] (result: %s)", batchItem.RequestID, batchItem.Index, message.Key, message.Offset, message.Partition, message.Topic, message.Value, putResult)
	}

}

// Get a response for an async request
func RetrieveAsyncResponse(requestID string) (BatchResponse, error) {
	redis := GetAsyncJobRedis()

	existsCmd := redis.Exists(requestID)
	existsResult, err := existsCmd.Result()
	if err != nil {
		log.Printf("An error occurred attempting to check if async batch request still exists: [request id: %s] (error: %s)", requestID, err)
	} else {
		log.Printf("Successfully retreived async job exists from redis. [request id: %s] (result: %t)", requestID, existsResult)
		if !existsResult {
			return BatchResponse{}, fmt.Errorf("The async batch request can not be found.  It may have expired.")
		}
	}

	getCmd := redis.LRange(requestID, 0, -1)
	getResult, err := getCmd.Result()
	if err != nil {
		log.Printf("An error occurred attempting to get response data for async batch request: [request id: %s] (error: %s)", requestID, err)
	} else {
		log.Printf("Successfully retreived async response data from redis. [request id: %s] (result: %s)", requestID, getResult)
	}

	// Check responses first to see if it's finished
	for _, response := range getResult {
		if response == "" {
			return BatchResponse{}, nil
		}
	}

	batchResponse := make(BatchResponse, len(getResult))
	for idx, response := range getResult {
		var responseItem BatchResponseItem
		_ = json.Unmarshal([]byte(response), &responseItem)
		batchResponse[idx] = responseItem
	}

	return batchResponse, nil

}
