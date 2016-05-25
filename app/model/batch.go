package model

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/Unified/pmn/lib/errors"
)

var HostMap map[string]string

type BatchResponseItem struct {
	Code    int               `json:"code"`
	Body    interface{}       `json:"body"`
	Headers map[string]string `json:"headers"`
}

type BatchResponse []BatchResponseItem

type BatchItem struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Body    interface{}       `json:"body"`
	Headers map[string]string `json:"headers"`
}

// Get the URL to hit for the batch item
func (batchItem BatchItem) FullURL() (string, *errors.JsonError) {
	parts := strings.SplitN(batchItem.URL, "://", 1)
	domain, found := HostMap[parts[0]]
	if ! found {
		log.Printf("An error occurred getting the batch URL for %s. Service unrecognized.", batchItem.URL)
		return "", errors.New("Unrecognized service: %s", 400, parts[0])
	}

	return domain + parts[1], nil
}

// Create a request for this batch item
func (batchItem BatchItem) NewRequest(identityID string) (*http.Request, *errors.JsonError) {
	data, _ := json.Marshal(batchItem.Body)
	url, jsonErr := batchItem.FullURL()
	if jsonErr != nil {
		return nil, jsonErr
	}

	request, err := http.NewRequest(batchItem.Method, url, bytes.NewBuffer(data))
	if err != nil {
		log.Printf("An error occurred making the new batch request: %s", err)
		return nil, errors.New("An Internal Server error occurred making the request", 500)
	}

	for header, val := range batchItem.Headers {
		request.Header.Add(header, val)
	}
	request.Header.Add("X-Unified-IdentityID", identityID)
	return request, nil
}

// Make a request for this batch item
func (batchItem BatchItem) Do(request *http.Request) (BatchResponseItem, error) {
	client := &http.Client{}
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

type BatchItems []BatchItem

// Request a single item from the BatchItems
func (batchItems BatchItems) RequestItem(batchItem BatchItem, response chan interface{}, identityID string) {
	request, jsonErr := batchItem.NewRequest(identityID)
	if jsonErr != nil {
		log.Printf("An error occurred creating request: %s %+v", jsonErr, batchItem)
		response <- jsonErr
	}
	responseItem, err := batchItem.Do(request)
	if err != nil {
		log.Printf("An error occurred making request: %s %+v", err, batchItem)
		response <- err
	}

	response <- responseItem
}

// Create an error message for the batch item
func (batchItems BatchItems) MakeError(code int, jsonErr *errors.JsonError) BatchResponseItem {
	return BatchResponseItem{
		Code: code,
		Body: jsonErr,
	}
}

// Runs all of the jobs in this list of batch items
func (batchItems BatchItems) RunBatch(identityID string) (BatchResponse) {

	batchResponseChans := make([]chan interface{}, len(batchItems))
	for idx, batchItem := range batchItems {
		batchResponseChans[idx] = make(chan interface{})
		go batchItems.RequestItem(batchItem, batchResponseChans[idx], identityID)
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
