/*
The routes package is used to configure the URL routes on the server
*/
package route

import (
	"github.com/Unified/batch/app/context"
	"github.com/Unified/batch/app/controller"
	"github.com/Unified/pmn/lib/config"
	commonMW "github.com/Unified/pmn/lib/middleware"

	"github.com/gocraft/web"
)

// Router handles all incoming requests, calls middleware and controller functions.
func Router() (root *web.Router) {
	root = web.New(context.Context{})

	if config.IsDevEnv() {
		root.Middleware(commonMW.LoggerMiddleware)
		root.Middleware(web.ShowErrorsMiddleware)
	}

	if !config.IsDevEnv() {
		root.Error(controller.Error)
	}

	// Leave in here as simple test route for load balancers and such
	root.Get("/ping", controller.Ping)

	// Catch-all Route
	root.NotFound(controller.NotFound)

	batchRoot := root.Subrouter(context.Context{}, "/")

	// Business logic middleware comes last
	batchRoot.Middleware(commonMW.IdentityID)
	batchRoot.Middleware(commonMW.SetHeaders)

	batchRoot.Post("/batch", controller.Batch)
	batchRoot.Post("/batch/async", controller.AsyncBatch)

	return
}
