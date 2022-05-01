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
	"time"
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

// PushAggIntoFFS writes the Aggregate data to the file system
func PushAggIntoFFS(wg *sync.WaitGroup, k string, rPool *redis.Pool, bar *progressbar.ProgressBar) {
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
	resBytes := ProcessRedisCommand[[]byte](rPool, "GET", args, true, "byte")

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
		// Get the keys from redis
		args := []interface{}{"*aggs*"}
		allKeys = ProcessRedisCommand[[]string](rPool, "KEYS", args, false, "string")

		// If there are no keys, stay in the loop, don't exit
		if len(allKeys) == 0 {
			continue
		} else {
			// If there are keys, write them to the flat file system
			for _, key := range allKeys {
				time.Sleep(time.Millisecond * 5)
				wg.Add(1)
				go PushAggIntoFFS(&wg, key, rPool, bar)
			}
		}
		wg.Wait()
	}
}
