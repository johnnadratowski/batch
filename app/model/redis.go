package model

import (
	"fmt"
	"gopkg.in/redis.v2"
)

// Used to retrieve a new Redis client
var NewRedisClient = func(host, port, password, db string) *redis.Client {
	return redis.NewTCPClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", REDIS_HOST, REDIS_PORT),
		Password: REDIS_PW,
		DB:       int64(REDIS_DB),
	})
}
