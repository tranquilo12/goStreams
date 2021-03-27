package db

import (
	"encoding/json"
	"fmt"
	"github.com/schollz/progressbar/v3"
	"lightning/utils/structs"
	"net/http"
	"sync"
	"time"
)
import "net/url"
import "go.uber.org/ratelimit"

func MakeAllAggRequests(urls []*url.URL, timespan string, multiplier int) <-chan []structs.AggregatesBars {
	// we are already receiving the AggregatesBarsRequests (un-flattened) here, so the job is to send over this data
	// to the flattener

	bar := progressbar.Default(int64(len(urls)), "Downloading...")

	// create a rate limiter to stop over-requesting
	rateLimiter := ratelimit.New(rateLimit)

	// create a channel to make sure all requests are not being thrown away, of the flattened type
	c := make(chan []structs.AggregatesBars, len(urls))
	prev := time.Now()

	// use WaitGroup to make things more smooth with channels
	var wg sync.WaitGroup

	// create a buffer of the waitGroup, of the same length as urls
	wg.Add(len(urls))

	for _, u := range urls {
		now := rateLimiter.Take()
		target := new(structs.AggregatesBarsResponse)

		go func(u *url.URL) {
			defer wg.Done()
			resp, err := http.Get(u.String())
			if err != nil {
				fmt.Println("Some Error: ", err)
				panic(err)
			} else {
				err = json.NewDecoder(resp.Body).Decode(&target)
				flattenedTarget := AggBarFlattenPayloadBeforeInsert1(*target, timespan, multiplier)
				c <- flattenedTarget
			}
			resp.Body.Close()
		}(u)

		now.Sub(prev)
		prev = now

		var barerr = bar.Add(1)
		if barerr != nil {
			fmt.Println("\nSomething wrong with bar1: ", barerr)
		}
	}
	wg.Wait()
	close(c)

	return c
}

func MakeTickerTypesRequest(apiKey string) *structs.TickerTypeResponse {
	TickerTypesUrl := MakeTickerTypesUrl(apiKey)
	TickerTypesTarget := new(structs.TickerTypeResponse)

	resp, err := http.Get(TickerTypesUrl.String())
	if err != nil {
		panic(err)
	}

	err = json.NewDecoder(resp.Body).Decode(&TickerTypesTarget)
	if err != nil {
		panic(err)
	}

	return TickerTypesTarget
}
