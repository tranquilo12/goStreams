package db

import (
	"fmt"
	"time"

	//"github.com/go-redis/redis/v7"
	"github.com/gomodule/redigo/redis"
)

func GetRedisPool(port int, endpoint string) *redis.Pool {
	addr := fmt.Sprintf("%s:%d", endpoint, port)
	pool := &redis.Pool{
		MaxIdle:     1000,
		IdleTimeout: 240 * time.Minute,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", addr)
		},
	}
	return pool
}

//client := redis.NewClient(&redis.Options{
//	Network:            "tcp",
//	Addr:               addr,
//	Password:           "",
//	DB:                 0,
//	MaxRetries:         0,
//	MinRetryBackoff:    0,
//	MaxRetryBackoff:    0,
//	DialTimeout:        0,
//	ReadTimeout:        0,
//	WriteTimeout:       0,
//	PoolSize:           1000,
//	MinIdleConns:       0,
//	MaxConnAge:         10000,
//	PoolTimeout:        1000,
//	IdleTimeout:        1000,
//	IdleCheckFrequency: 0,
//	TLSConfig:          nil,
//})

//	return client
//}
