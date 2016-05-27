package controller

import (
	"github.com/gocraft/web"

	"encoding/json"
	"github.com/Unified/batch/app/context"
	"github.com/Unified/batch/app/model"
	"github.com/Unified/pmn/lib/config"
	"github.com/Unified/pmn/lib/errors"
	"log"
)

// Batch processes batch requests
func Batch(c *context.Context, rw web.ResponseWriter, req *web.Request) {
	var batchItems model.BatchItems
	if err := json.NewDecoder(req.Body).Decode(&batchItems); err != nil {
		errors.Write(rw, 400, "Unable to parse request body JSON")
		return
	}

	if len(batchItems) > config.GetInt("max_batch_requests") {
		errors.Write(rw, 400, "Too many batch requests at once. Max allowed: %d", config.GetInt("max_batch_requests"))
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
		errors.Write(rw, 400, "Unable to parse request body JSON")
		return
	}

	if len(batchItems) > config.GetInt("max_batch_requests_async") {
		errors.Write(rw, 400, "Too many async batch requests at once. Max allowed: %d", config.GetInt("max_batch_requests_async"))
		return
	}

	requestID, jsonErr := batchItems.RunBatchAsync(c.IdentityID)
	if jsonErr != nil {
		jsonErr.Write(rw)
		return
	}

	err := json.NewEncoder(rw).Encode(requestID)
	if err != nil {
		log.Printf("An error occurred writing response: %s", err)
		errors.Write(rw, 500, "An internal server error occurred")
		return
	}

}
