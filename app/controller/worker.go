package controller

import (
	"time"
	"github.com/Unified/pmn/lib/config"
)

func StartWorker() {
	for {
		time.Sleep(config.GetInt("worker_sleep"))
	}
}
