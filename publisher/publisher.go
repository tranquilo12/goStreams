package publisher

import (
	"encoding/json"
	"fmt"
	"github.com/adjust/rmq/v3"
	"github.com/gosuri/uiprogress"
	"github.com/nitishm/go-rejson"
	"go.uber.org/ratelimit"
	"lightning/utils/db"
	"net/http"
	"strings"
	"sync"

	"github.com/gomodule/redigo/redis"
	"lightning/utils/structs"
	"net/url"
	"time"
)

// CreateAggKey A Function that creates the index for the redis database.
// keys look like: $2021_11_05_AAPL_1_minute_from_2021_01_01_to_2021_04_01
// keys start with a $ to be index-able.
func CreateAggKey(u *url.URL) string {
	// for the idx date, will replace all the '-' with '_' as it plays better with redis.
	idxDate := strings.ReplaceAll("$"+time.Now().Format("2006-02-01"), "-", "_")

	// Split the entire 'path' or 'url', isolate each element and make a key.
	splitPath := strings.Split(u.Path, "/")

	// Get ticker from the 'url'
	ticker := splitPath[4]

	// Get agg range... combines the aggregate timespan and multiplier... e.g 1_minute
	aggRange := splitPath[6] + "_" + splitPath[7]

	// Get a fromTo string, e.g from_2021_01_01_to_2021_04_01
	fromTo := strings.ReplaceAll(strings.Join([]string{"from", splitPath[8], "to", splitPath[9]}, "_"), "-", "_")

	// Make the full key that can be accessed by the "Subscriber" and then pushed into the database.
	fullKey := strings.Join([]string{idxDate, ticker, aggRange, fromTo}, "_")
	return fullKey
}

// MarshalAggAndPushKeyToRedis A function that takes a RE-JSON handler, a bar response and a url and
// pushes the Marshalled JSON as a Redis obj.
func MarshalAggAndPushKeyToRedis(rh *rejson.Handler, data *structs.AggregatesBarsResponse, u *url.URL) error {
	// Use the function CreateAggKey from the url, to generate a unique key that can be "Subscribed" to.
	fullKey := CreateAggKey(u)

	// Push key to redis server
	_, err := rh.JSONSet(fullKey, ".", data) // JSONSet <idxDate> [path] data struct
	if err != nil {
		panic(err)
	}
	return err
}

// MarshalAggAndPublishToAgg If we have to publish to a redis channel instead, where we're expecting a streaming
// result, Experimental?
func MarshalAggAndPublishToAgg(conn redis.Conn, data *structs.AggregatesBarsResponse) error {
	// just Marshal data to string
	dataStr, _ := json.Marshal(data)

	// Get a bool result, from publishing the data to an "AGG" Redis Channel.
	_, err := redis.Bool(conn.Do("PUBLISH", "AGG", dataStr))
	if err != nil {
		panic(err)
	}
	return err
}

// AggPublisher Finally, the function that takes all the above functions in this file and tries to make sense out it.
// With all the urls pushed to this function, we rate-limit all requests, Marshal Json response either to a Redis structure
// and push it to redis with a key, or to a string and push it to a channel.
//func AggPublisher(redisPool *redis.Pool, rh *rejson.Handler, urls []*url.URL, publishToChannel bool) error {
func AggPublisher(urls []*url.URL) error {
	// err and response variables to make things easy
	var err error
	var errChan chan error
	//var resp *http.Response

	// use WaitGroup to make things more smooth with goroutines
	var wg sync.WaitGroup

	// create a buffer of the waitGroup, of the same length as urls
	wg.Add(len(urls))

	// create a rate limiter to stop over-requesting
	prev := time.Now()
	rateLimiter := ratelimit.New(1)

	client := db.GetRedisClient(7000)
	queueConnection, err := rmq.OpenConnectionWithRedisClient("AGG", client, errChan)
	if err != nil {
		fmt.Println("Something wrong with this queueConnection...")
	}

	taskQueue, err := queueConnection.OpenQueue("AGG")
	if err != nil {
		panic(err)
	}

	uiprogress.Start()
	bar := uiprogress.AddBar(len(urls))
	bar.AppendCompleted()
	bar.PrependElapsed()

	for _, u := range urls {
		time.Sleep(1)
		now := rateLimiter.Take()
		target := new(structs.AggregatesBarsResponse)

		go func(u *url.URL) {
			defer wg.Done()
			resp, err := http.Get(u.String())
			if err != nil {
				fmt.Println("Error retrieving URL: ", err)
				panic(err)
			} else {
				err = json.NewDecoder(resp.Body).Decode(&target)
				taskBytes, err := json.Marshal(target)
				if err != nil {
					res := fmt.Sprintf("Target Err: %s", err)
					fmt.Printf(res)
				}
				err = taskQueue.PublishBytes(taskBytes)
				for bar.Incr() {
					time.Sleep(1)
				}
			}
		}(u)

		now.Sub(prev)
		prev = now
	}
	wg.Wait()
	return err
}
