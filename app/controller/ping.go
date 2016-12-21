package controller

import (
	"fmt"

	"github.com/gocraft/web"

	"github.com/johnnadratowski/batch/app/context"
)

// Ping test endpoint
func Ping(c *context.Context, rw web.ResponseWriter, req *web.Request) {
	rw.Header().Set("Content-Type", "text/plain")
	fmt.Fprint(rw, "PONG!")
}
