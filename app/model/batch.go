package model

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/Shopify/sarama"
	"github.com/Unified/pmn/lib/config"
	"github.com/Unified/pmn/lib/errors"
	"github.com/pborman/uuid"
	"time"
)

// Contains the mapping for internal services - like "pmn": "http://pmn-load-balancer:80/"
var HostMap map[string]string

// Interface for the http method "Do", useful for mocking requests/responses
type BatchClient interface {
	Do(req *http.Request) (resp *http.Response, err error)
}

// The response for a single item in a batch
type BatchResponseItem struct {
	Code    int               `json:"code"`
	Body    interface{}       `json:"body"`
	Headers map[string]string `json:"headers"`
}

// The list of responses for all of the batch item requests
type BatchResponse []BatchResponseItem

// A single batch item request
type BatchItem struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Body    interface{}       `json:"body"`
	Headers map[string]string `json:"headers"`
}

// Get the URL to hit for an internal request batch item
func (batchItem BatchItem) InternalURL() (string, *errors.JsonError) {
	parts := strings.SplitN(batchItem.URL, "://", 2)
	domain, found := HostMap[parts[0]]
	if !found {
		log.Printf("An error occurred getting the batch URL for %s. Service unrecognized.", batchItem.URL)
		return "", errors.New("Unrecognized service: %s", 400, parts[0])
	}

	if !strings.HasSuffix(domain, "/") {
		domain += "/"
	}
	return domain + parts[1], nil
}

// Create a request for this internal request batch item
func (batchItem BatchItem) NewInternalRequest(identityID string) (*http.Request, *errors.JsonError) {
	data, _ := json.Marshal(batchItem.Body)
	url, jsonErr := batchItem.InternalURL()
	if jsonErr != nil {
		return nil, jsonErr
	}

	request, err := http.NewRequest(strings.ToUpper(batchItem.Method), url, bytes.NewBuffer(data))
	if err != nil {
		log.Printf("An error occurred making the new internal batch request: %s", err)
		return nil, errors.New("An Internal Server error occurred making the request", 500)
	}

	for header, val := range batchItem.Headers {
		request.Header.Add(header, val)
	}
	request.Header.Add("X-Unified-IdentityID", identityID)
	return request, nil
}

// Create a request for this external request batch item
func (batchItem BatchItem) NewExternalRequest(identityID string) (*http.Request, *errors.JsonError) {
	data, _ := json.Marshal(batchItem.Body)

	request, err := http.NewRequest(strings.ToUpper(batchItem.Method), batchItem.URL, bytes.NewBuffer(data))
	if err != nil {
		log.Printf("An error occurred making the new external batch request: %s", err)
		return nil, errors.New("An Internal Server error occurred making the request", 500)
	}

	for header, val := range batchItem.Headers {
		request.Header.Add(header, val)
	}

	return request, nil
}

// Create a request for this request batch item
func (batchItem BatchItem) NewRequest(identityID string) (*http.Request, *errors.JsonError) {
	var request *http.Request
	var jsonErr *errors.JsonError
	if strings.HasPrefix(batchItem.URL, "http") {
		// Represents a request to an external system
		request, jsonErr = batchItem.NewExternalRequest(identityID)
		if jsonErr != nil {
			log.Printf("An error occurred creating request: %s %+v", jsonErr, batchItem)
			return request, jsonErr
		}
	} else {
		// Represents a request to an internal system
		request, jsonErr = batchItem.NewInternalRequest(identityID)
		if jsonErr != nil {
			log.Printf("An error occurred creating request: %s %+v", jsonErr, batchItem)
			return request, jsonErr
		}
	}
	return request, jsonErr
}

// Make a request for this batch item
func (batchItem BatchItem) Do(request *http.Request) (BatchResponseItem, error) {
	client := GetRequestClient()
	response, err := client.Do(request)
	if err != nil {
		log.Printf("An error occurred calling the new batch request: %s", err)
		return BatchResponseItem{}, errors.New("An Internal Server error occurred making the request", 500)
	}

	responseItem := BatchResponseItem{
		Code: response.StatusCode,
	}

	err = json.NewDecoder(response.Body).Decode(&responseItem.Body)
	if err != nil {
		log.Printf("An error occurred reading the batch item response: %s", err)
		return BatchResponseItem{}, errors.New("An Internal Server error occurred making the request", 500)
	}

	return responseItem, nil
}

// Request a single item from the BatchItems.  Meant to be used asynchronously using a channel.
func (batchItem BatchItem) RequestItemAsync(response chan interface{}, identityID string) {
	request, jsonErr := batchItem.NewRequest(identityID)
	if jsonErr != nil {
		response <- jsonErr
		return
	}

	responseItem, err := batchItem.Do(request)
	if err != nil {
		log.Printf("An error occurred making request: %s %+v", err, batchItem)
		response <- err
		return
	}

	response <- responseItem
}

// Request a single item from the BatchItems.  Meant to be used asynchronously using a channel.
func (batchItem BatchItem) RequestItem(identityID string) (BatchResponseItem, *errors.JsonError) {
	request, jsonErr := batchItem.NewRequest(identityID)
	if jsonErr != nil {
		return BatchResponseItem{}, jsonErr
	}

	responseItem, err := batchItem.Do(request)
	if err != nil {
		log.Printf("An error occurred making request: %s %+v", err, batchItem)
		return responseItem, errors.New("An internal server error occurred", 500)
	}

	return responseItem, nil
}

type BatchItems []BatchItem

// Create an error message for the batch item
func (batchItems BatchItems) MakeError(code int, jsonErr *errors.JsonError) BatchResponseItem {
	return BatchResponseItem{
		Code: code,
		Body: jsonErr,
	}
}

// Runs all of the jobs in this list of batch items
func (batchItems BatchItems) RunBatch(identityID string) BatchResponse {

	batchResponseChans := make([]chan interface{}, len(batchItems))
	for idx, batchItem := range batchItems {
		batchResponseChans[idx] = make(chan interface{})
		go batchItem.RequestItemAsync(batchResponseChans[idx], identityID)
	}

	batchResponse := make(BatchResponse, len(batchItems))
	for idx, batchResponseChan := range batchResponseChans {
		itemResponse := <-batchResponseChan

		switch itemResponse := itemResponse.(type) {
		case *errors.JsonError:
			batchResponse[idx] = batchItems.MakeError(500, itemResponse)

		case BatchResponseItem:
			batchResponse[idx] = itemResponse

		}
	}

	return batchResponse
}

// Runs all of the jobs in this list of batch items
func (batchItems BatchItems) RunBatchAsync(identityID string) (string, *errors.JsonError) {

	producer, err := GetAsyncBatchProducer()
	if err != nil {
		return "", errors.New("An internal server error occurred.", 500)
	}
	defer producer.Close()

	requestID := uuid.New()
	for idx, batchItem := range batchItems {
		asyncItem := AsyncBatchItem{
			RequestID:  requestID,
			Index:      int64(idx),
			Item:       batchItem,
			IdentityID: identityID,
		}
		output, _ := json.Marshal(asyncItem)
		message := &sarama.ProducerMessage{
			Topic: config.Get("topic"),
			Key:   sarama.ByteEncoder(identityID + batchItem.URL),
			Value: sarama.ByteEncoder(output),
		}

		partition, offset, err := producer.SendMessage(message)
		if err != nil {
			log.Printf("An error occurred sending message to Kafka. (error: %s)", err)
			return "", errors.New("An internal server error occurred.", 500)
		} else {
			log.Printf("Items successfully sent: [request id: %s] [parition: %d] (offset: %d)", requestID, partition, offset)
		}
	}

	redis := GetAsyncJobRedis()
	defer redis.Close()

	pushCmd := redis.LPush(requestID, make([]string, len(batchItems))...)
	pushResult, err := pushCmd.Result()
	if err != nil {
		log.Printf("An error occurred saving new request to Redis: [request id: %s] [Result: %d] (error: %s)", requestID, pushResult, err)
		return "", errors.New("An internal server error occurred.", 500)
	} else {
		log.Printf("New async batch request successfully sent to redis: [request id: %s] [Num Items: %d]", requestID, pushResult)
	}

	expireCmd := redis.Expire(requestID, time.Duration(config.GetInt("async_expire"))*time.Minute)
	expireResult, err := expireCmd.Result()
	if err != nil {
		log.Printf("An error occurred saving new request to Redis: [request id: %s] [Expire set?: %t] (error: %s)", requestID, expireResult, err)
	} else {
		log.Printf("Expiration on new async batch request successfully sent to redis: [request id: %s] [Expire set?: %t]", requestID, expireResult)
	}

	return requestID, nil
}
