package controller

import (
	"fmt"
	"net/http"

	"github.com/gocraft/web"

	"github.com/Unified/batch/app/context"
	"github.com/Unified/pmn/lib/errors"
	webLib "github.com/Unified/pmn/lib/web"
)

// Not found is the catch-all endpoint for when no endpoint is found
func NotFound(c *context.Context, rw web.ResponseWriter, req *web.Request) {
	if req.Method == "OPTIONS" {
		webLib.Write204(rw)
	} else {
		rw.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(rw, errors.New("Not Found", 404).Error())
	}
}
