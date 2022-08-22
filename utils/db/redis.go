package db

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
)

func Get(pool *redis.Pool, key string) ([]byte, error) {
	conn := pool.Get()
	defer conn.Close()

	var data []byte
	data, err := redis.Bytes(conn.Do("GET", key))
	if err != nil {
		return data, fmt.Errorf("error getting key %s: %v", key, err)
	}
	return data, err
}
