package db

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"github.com/go-pg/pg/v10"
	"github.com/gomodule/redigo/redis"
	"github.com/schollz/progressbar/v3"
	"lightning/utils/structs"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const DataDir = "/Users/shriramsunder/GolandProjects/goStreams/data"

func Check(err error) {
	if err != nil {
		panic(err)
	}
}

func CheckAndAssign[T any](val T, err error) T {
	Check(err)
	return val
}

func PushTickerTypesIntoDB(insertIntoDB *structs.TickerTypeResponse, db *pg.DB) error {
	flattenedInsertIntoDB := structs.TickerTypesFlattenPayloadBeforeInsert(insertIntoDB)
	_, err := db.Model(&flattenedInsertIntoDB).Insert()
	if err != nil {
		panic(err)
	}
	println("Inserted all TickerTypes into the DB...")
	return nil
}

func PushTickerVxIntoRedis(insertIntoRedis <-chan []structs.TickerVx, rConn redis.Conn) error {
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

	// Join all the tickers into a single string
	allTickersStr := strings.Join(allTickers[:], ",")

	// Create an args that's an array of strings, and process the redis command.
	args := []interface{}{"allTickers", allTickersStr}
	_ = ProcessRedisCommand[[]string](rConn, "SET", args, false, "string")
	return nil
}

func GetTickerVxs(insertIntoRedis <-chan []structs.TickerVx) string {
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
	allTickersStr := strings.Join(allTickers[:], ",")
	return allTickersStr
}

func PushTickerNews2IntoDB(insertIntoDB <-chan []structs.TickerNews2, db *pg.DB) error {
	// use WaitGroup to make things more smooth with channels
	var wg sync.WaitGroup

	bar := progressbar.Default(500, "Uploading TickerNews2 to db...")

	// for each insertIntoDB that follows...spin off another go routine
	for val, ok := <-insertIntoDB; ok; val, ok = <-insertIntoDB {
		if ok && val != nil {
			wg.Add(1)

			go func(val []structs.TickerNews2) {
				defer wg.Done()

				_, err := db.Model(&val).
					OnConflict("(id) DO NOTHING").
					Insert()
				if err != nil {
					panic(err)
				}

				var barerr = bar.Add(1)
				if barerr != nil {
					fmt.Println("\nSomething wrong with bar: ", barerr)
				}
			}(val)
		}
	}
	wg.Wait()
	return nil
}

// ProcessRedisCommand takes a redis command and returns the result
// Trying out generics here, this function can return either a []string or a []byte
func ProcessRedisCommand[T []string | []byte](
	rConn redis.Conn,
	cmd string,
	args []interface{},
	deleteKey bool,
	retType string,
) T {
	// Create a new res variable of type T, that's instantiated with nil
	// If there's anything to be returned, it will be assigned to res
	// otherwise, it will remain nil.
	// This is to ensure that command with "SET" will always return a nil.
	var res T

	// Send the command to redis, can be GET, SET, etc.
	err := rConn.Send(cmd, args...)
	Check(err)

	// Flush the buffer, clears the output buffer
	err = rConn.Flush()
	Check(err)

	// Receive the value from redis, if the command is GET, and depending on the type,
	// can be either []string or []byte
	if cmd == "GET" {
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

	// Close the redis connection
	err = rConn.Close()
	Check(err)

	return res
}

// PushAggIntoFFS writes the Aggregate data to the file system
func PushAggIntoFFS(wg *sync.WaitGroup, k string, rConn redis.Conn, bar *progressbar.ProgressBar) {
	// instantiate a new file buffer and defer wg.Done()
	var fileGZ bytes.Buffer
	defer wg.Done()

	// Change directory to the data directory, for this part of the program
	err := os.Chdir(DataDir)
	Check(err)

	// Replace the : with a -
	kPath := strings.Replace(k, ":", "-", -1)

	// Make the directory if it doesn't exist
	dirPath := strings.Split(kPath, "data.json")[0]
	err = CheckAndAssign(os.MkdirAll(dirPath, 0777), nil)
	Check(err)

	// Check if the directory exists, sometimes it's not created
	_, err = os.Stat(dirPath)
	Check(err)

	// Get the value from redis
	args := []interface{}{k}
	resBytes := ProcessRedisCommand[[]byte](rConn, "GET", args, true, "byte")

	// Write the bytes to the gzip writer
	zipper := gzip.NewWriter(&fileGZ)
	_, err = zipper.Write(resBytes)
	Check(err)

	// Close the gzip writer, don't defer the close as it causes a problem
	err = zipper.Close()
	Check(err)

	// Open/Create the file to the filesystem
	fullFilePath := filepath.Join(kPath + ".gz")
	wf, err := os.OpenFile(fullFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0777)
	Check(err)

	// Write the gzipped bytes to the flat file system
	_, err = wf.Write(fileGZ.Bytes())
	Check(err)

	// Close the file, don't defer the close as it causes a problem
	err = wf.Close()
	Check(err)

	// Add to the progress bar
	err = bar.Add(1)
	Check(err)
}

// PushAggIntoFFSCont writes the Aggregate data to the file system continuously.
func PushAggIntoFFSCont(rPool *redis.Pool) error {
	// use WaitGroup to make things more smooth with goroutines
	var wg sync.WaitGroup
	var allKeys []string

	// Define a progress bar
	bar := progressbar.New(30000)

	// for each insertIntoDB that follows...spin off another go routine
	for {
		// Get a conn from the Pool
		conn := rPool.Get()

		// Get the keys from redis
		args := []interface{}{"*"}
		allKeys = ProcessRedisCommand[[]string](conn, "KEYS", args, false, "string")

		// If there are no keys, stay in the loop, don't exit
		if len(allKeys) == 0 {
			continue
		} else {
			// Get another conn from the Pool
			conn := rPool.Get()

			// If there are keys, write them to the flat file system
			for _, key := range allKeys {
				wg.Add(1)
				go PushAggIntoFFS(&wg, key, conn, bar)
			}
		}
		wg.Wait()
	}
}
