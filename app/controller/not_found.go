package controller

import (
	"fmt"
	"net/http"

	"github.com/gocraft/web"

	"github.com/johnnadratowski/batch/app/context"
)

// Not found is the catch-all endpoint for when no endpoint is found
func NotFound(c *context.Context, rw web.ResponseWriter, req *web.Request) {
	if req.Method == "OPTIONS" {
		rw.WriteHeader(204)
	} else {
		rw.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(rw, fmt.Errorf("Not Found").Error())
	}
}
