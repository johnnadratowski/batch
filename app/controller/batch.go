package controller

import (
	"github.com/Unified/golang-lib/lib/errors"
	"github.com/gocraft/web"

	"encoding/json"
	"fmt"
	"log"

	"github.com/johnnadratowski/batch/app/context"
	"github.com/johnnadratowski/batch/app/model"
)

var MAX_REQUESTS int = 1000
var MAX_REQUESTS_ASYNC int = 10000

// Batch processes batch requests
func Batch(c *context.Context, rw web.ResponseWriter, req *web.Request) {
	var batchItems model.BatchItems
	if err := json.NewDecoder(req.Body).Decode(&batchItems); err != nil {
		err := fmt.Errorf("Unable to parse JSON")
		fmt.Fprint(rw, err)
		return
	}

	if len(batchItems) > MAX_REQUESTS {
		err := fmt.Errorf("Too many batch requests at once. Max allowed: %d Sent: %d", MAX_REQUESTS, len(batchItems))
		fmt.Fprint(rw, err)
		return
	} else if len(batchItems) == 0 {
		err := fmt.Errorf("No batch items recieved")
		fmt.Fprint(rw, err)
		return
	}

	batchResponse := batchItems.RunBatch(c.IdentityID)

	err := json.NewEncoder(rw).Encode(batchResponse)
	if err != nil {
		log.Printf("An error occurred writing response: %s", err)
		errors.Write(rw, 500, "An internal server error occurred")
		return
	}
}

// AsyncBatch processes batch requests asynchronously
func AsyncBatch(c *context.Context, rw web.ResponseWriter, req *web.Request) {
	var batchItems model.BatchItems
	if err := json.NewDecoder(req.Body).Decode(&batchItems); err != nil {
		err := fmt.Errorf("Unable to parse JSON")
		fmt.Fprint(rw, err)
		return
	}

	if len(batchItems) > MAX_REQUESTS_ASYNC {
		err := fmt.Errorf("Too many async batch requests at once. Max allowed: %d Sent: %d", MAX_REQUESTS, len(batchItems))
		fmt.Fprint(rw, err)
		return
	} else if len(batchItems) == 0 {
		err := fmt.Errorf("No batch items recieved")
		fmt.Fprint(rw, err)
		return
	}

	requestID, err := batchItems.RunBatchAsync(c.IdentityID)
	if err != nil {
		fmt.Fprint(rw, err)
		return
	}

	rw.Header().Set("LOCATION", "/batch/async/"+requestID)
	rw.WriteHeader(202)
}

// AsyncBatchRetrieve retrieves an asynchronous batch requests data
func AsyncBatchRetrieve(c *context.Context, rw web.ResponseWriter, req *web.Request) {
	requestID := req.PathParams["requestID"]
	batchResponse, err := model.RetrieveAsyncResponse(requestID)
	if err != nil {
		fmt.Fprint(rw, err)
		return
	} else if len(batchResponse) == 0 {
		rw.Header().Set("LOCATION", "/batch/async/"+requestID)
		rw.WriteHeader(202)
	} else {
		err := json.NewEncoder(rw).Encode(batchResponse)
		if err != nil {
			log.Printf("An error occurred writing response: %s", err)
			errors.Write(rw, 500, "An internal server error occurred")
			return
		}
	}
}
