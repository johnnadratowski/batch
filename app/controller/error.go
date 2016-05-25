package controller

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gocraft/web"

	"github.com/Unified/batch/app/context"
	"github.com/Unified/pmn/lib/errors"
)

func Error(c *context.Context, rw web.ResponseWriter, req *web.Request, err interface{}) {
	rw.WriteHeader(http.StatusInternalServerError)

	jsonErr := errors.New("Internal Server Error", 500)
	fmt.Fprint(rw, jsonErr)

	log.Println("*******************************************************************")
	log.Println("*****************A panic occurred during processing!***************")
	log.Println("*******************************************************************")
	log.Printf("Context: %+v", c)
	log.Printf(err.(string))

	return
}
