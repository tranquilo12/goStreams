package db

import (
	"fmt"
	"github.com/go-redis/redis/v7"
)

func GetRedisClient(port int) *redis.Client {
	addr := fmt.Sprintf("localhost: %d", port)

	client := redis.NewClient(&redis.Options{
		Network:            "tcp",
		Addr:               addr,
		Password:           "",
		DB:                 0,
		MaxRetries:         0,
		MinRetryBackoff:    0,
		MaxRetryBackoff:    0,
		DialTimeout:        0,
		ReadTimeout:        0,
		WriteTimeout:       0,
		PoolSize:           0,
		MinIdleConns:       0,
		MaxConnAge:         0,
		PoolTimeout:        0,
		IdleTimeout:        0,
		IdleCheckFrequency: 0,
		TLSConfig:          nil,
	})

	return client
}
