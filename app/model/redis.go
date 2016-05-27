package model

import (
	"fmt"
	"github.com/Unified/pmn/lib/config"
	"gopkg.in/redis.v2"
)

// Used to retrieve a new Redis client
var NewRedisClient = func(host, port, password, db string) *redis.Client {
	return redis.NewTCPClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", config.Get("redis_host"), config.Get("redis_port")),
		Password: config.Get("redis_password"),
		DB:       int64(config.GetInt("redis_db")),
	})
}
