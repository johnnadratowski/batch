package controller

import (
	"github.com/gocraft/web"

	"encoding/json"
	"github.com/Unified/batch/app/context"
	"github.com/Unified/batch/app/model"
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

	batchResponse := batchItems.RunBatch(c.IdentityID)

	err := json.NewEncoder(rw).Encode(batchResponse)
	if err != nil {
		log.Printf("An error occurred writing response: %s", err)
		errors.Write(rw, 500, "An internal server error occurred")
		return
	}

}
