package db

import (
	"encoding/json"
	"fmt"
	"lightning/utils/structs"
	"time"

	"github.com/gomodule/redigo/redis"
)

// GetRedisPool Get redis pool connection object that can be used to get redis connection object
func GetRedisPool(port int, endpoint string) *redis.Pool {
	addr := fmt.Sprintf("%s:%d", endpoint, port)
	pool := &redis.Pool{
		MaxIdle:     30000,
		MaxActive:   35000,
		IdleTimeout: 240 * time.Minute,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", addr)
		},
	}
	return pool
}

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

func Set(pool *redis.Pool, key string, value []byte) error {
	conn := pool.Get()
	defer conn.Close()

	_, err := conn.Do("SET", key, value)
	if err != nil {
		v := string(value)
		if len(v) > 15 {
			v = v[0:12] + "..."
		}
		return fmt.Errorf("error setting key %s to %s: %v", key, v, err)
	}
	return err
}

func GetKeys(pool *redis.Pool, pattern string) ([]string, error) {

	conn := pool.Get()
	defer conn.Close()

	iter := 0
	var keys []string
	for {
		arr, err := redis.Values(conn.Do("SCAN", iter, "MATCH", pattern))
		if err != nil {
			return keys, fmt.Errorf("error retrieving '%s' keys", pattern)
		}

		iter, _ = redis.Int(arr[0], nil)
		k, _ := redis.Strings(arr[1], nil)
		keys = append(keys, k...)

		if iter == 0 {
			break
		}
	}

	return keys, nil
}

func Delete(pool *redis.Pool, key string) error {
	conn := pool.Get()
	defer conn.Close()
	_, err := conn.Do("DEL", key)
	return err
}

// PushTickerVxIntoRedis Reads from the channel and pushes the ticker struct into redis,
// the key is the ticker path that will be stored in the file system eventually, the value is the ticker struct
func PushTickerVxIntoRedis(insertIntoRedis <-chan []structs.TickerVx, pool *redis.Pool) error {
	// use WaitGroup to make things more smooth with channels
	var allTickers []string

	// for each insertIntoDB that follows...spin off another go routine
	for val, ok := <-insertIntoRedis; ok; val, ok = <-insertIntoRedis {
		if ok && val != nil {
			for _, v := range val {
				allTickers = append(allTickers, v.Ticker)
			}
		}
	}

	// Create an args that's an array of strings, and process the redis command.
	res, err := json.Marshal(allTickers)
	Check(err)

	//args := []interface{}{"allTickers", res}
	err = Set(pool, "allTickers", res)
	Check(err)

	//_ = ProcessRedisCommand[[]string](pool, "SET", args, false, "string")
	return nil
}

// ProcessRedisCommand takes a redis command and returns the result
// Trying out generics here, this function can return either a []string or a []byte
func ProcessRedisCommand[T []string | []byte](
	pool *redis.Pool,
	cmd string,
	args []interface{},
	deleteKey bool,
	retType string,
) T {
	var res T
	rConn := pool.Get()
	defer func(rConn redis.Conn) {
		_ = rConn.Close()
	}(rConn)

	// Send the command to redis, can be GET, SET, etc.
	err := rConn.Send(cmd, args...)
	Check(err)

	// Flush the buffer, clears the output buffer
	err = rConn.Flush()
	Check(err)

	// Receive the value from redis, if the command is GET, and depending on the type,
	// can be either []string or []byte
	if cmd == "GET" || cmd == "KEYS" {

		// Depending upon the return type, return either a []string or []byte
		switch retType {
		case "string":
			r, err := redis.Strings(rConn.Receive())
			Check(err)
			res = any(r).(T)
		default:
			r, err := redis.Bytes(rConn.Receive())
			Check(err)
			res = any(r).(T)
		}
	}

	// if deleteKey, then delete the key
	if deleteKey {
		err = rConn.Send("DEL", args)
		Check(err)
	}

	return res
}
