/*
The routes package is used to configure the URL routes on the server
*/
package route

import (
	"github.com/johnnadratowski/batch/app/context"
	"github.com/johnnadratowski/batch/app/controller"

	"github.com/gocraft/web"
)

// Router handles all incoming requests, calls middleware and controller functions.
func Router() (root *web.Router) {
	root = web.New(context.Context{})

	root.Middleware(web.ShowErrorsMiddleware)

	root.Error(controller.Error)

	// Leave in here as simple test route for load balancers and such
	root.Get("/ping", controller.Ping)

	// Catch-all Route
	root.NotFound(controller.NotFound)

	batchRoot := root.Subrouter(context.Context{}, "/")

	batchRoot.Post("/batch", controller.Batch)
	batchRoot.Post("/batch/async", controller.AsyncBatch)
	batchRoot.Get("/batch/async/:requestID", controller.AsyncBatchRetrieve)

	return
}
