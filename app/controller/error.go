package controller

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gocraft/web"

	"github.com/johnnadratowski/batch/app/context"
)

func Error(c *context.Context, rw web.ResponseWriter, req *web.Request, err interface{}) {
	rw.WriteHeader(http.StatusInternalServerError)

	err = fmt.Errorf("Internal Server Error")
	fmt.Fprint(rw, err)

	log.Println("*******************************************************************")
	log.Println("*****************A panic occurred during processing!***************")
	log.Println("*******************************************************************")
	log.Printf("Context: %+v", c)
	log.Printf(err.(string))

	return
}
