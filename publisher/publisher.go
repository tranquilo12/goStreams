package publisher

import (
	"encoding/json"
	"fmt"
	"go.uber.org/ratelimit"
	"net/http"
	"sync"

	//"fmt"
	"github.com/nitishm/go-rejson"
	"strings"

	//"github.com/go-pg/pg/v10"
	//"github.com/go-redis/redis/v8"
	"github.com/gomodule/redigo/redis"
	//"github.com/schollz/progressbar/v3"
	//"go.uber.org/ratelimit"
	//"lightning/utils/db"
	"lightning/utils/structs"
	//"net/http"
	"net/url"
	//"sync"
	"time"
)

const (
	apiKey     = "9AheK9pypnYOf_DU6TGpydCK6IMEVkIw"
	timespan   = "minute"
	from_      = "2021-01-01"
	to_        = "2021-03-20"
	multiplier = 1
)

func CreateAggKey(u *url.URL) string {
	// for the idx date
	idxDate := strings.ReplaceAll("$"+time.Now().Format("2006-02-01"), "-", "_")

	// now we need to check for every one of these keys
	splitPath := strings.Split(u.Path, "/")
	ticker := splitPath[4]
	aggRange := splitPath[6] + "_" + splitPath[7] //+ "._" +

	fromTo := strings.ReplaceAll(strings.Join([]string{"from", splitPath[8], "to", splitPath[9]}, "_"), "-", "_")
	fullKey := strings.Join([]string{idxDate, ticker, aggRange, fromTo}, "_")
	return fullKey
}

func MarshalAggAndPushKeyToRedis(rh *rejson.Handler, data *structs.AggregatesBarsResponse, u *url.URL) error {
	fullKey := CreateAggKey(u)
	_, err := rh.JSONSet(fullKey, ".", data) // JSONSet <idxDate> [path] data struct
	if err != nil {
		panic(err)
	}
	return err
}

func MarshalAggAndPublishToAgg(conn redis.Conn, data *structs.AggregatesBarsResponse) error {
	dataStr, _ := json.Marshal(data)
	_, err := redis.Bool(conn.Do("PUBLISH", "AGG", dataStr))
	if err != nil {
		panic(err)
	}
	return err
}

func AggPublisher(conn redis.Conn, rh *rejson.Handler, urls []*url.URL, publishToChannel bool) error {
	// err and response variables to make things easy
	var err error
	var resp *http.Response

	// use WaitGroup to make things more smooth with goroutines
	var wg sync.WaitGroup

	// create a buffer of the waitGroup, of the same length as urls
	wg.Add(len(urls))

	// create a rate limiter to stop over-requesting
	prev := time.Now()
	rateLimiter := ratelimit.New(50)

	for _, u := range urls {
		now := rateLimiter.Take()
		target := new(structs.AggregatesBarsResponse)

		go func(u *url.URL) {
			defer wg.Done()
			resp, err = http.Get(u.String())

			if err != nil {
				fmt.Println("Error retrieving URL: ", err)
				panic(err)
			} else {
				err = json.NewDecoder(resp.Body).Decode(&target)

				//flattenedTarget := db.AggBarFlattenPayloadBeforeInsert1(*target, timespan, multiplier)
				err := MarshalAggAndPushKeyToRedis(rh, target, u)
				if err != nil {
					fmt.Println("Error MarshallingAgg: ", err)
					panic(err)
				}

				// situation where things have to be published to channel
				if publishToChannel {
					err = MarshalAggAndPublishToAgg(conn, target)
				}
			}
		}(u)

		now.Sub(prev)
		prev = now

	}
	wg.Wait()
	resp.Body.Close()
	return err
}
