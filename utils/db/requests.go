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
				flattenedTarget := AggBarFlattenPayloadBeforeInsert(*target, timespan, multiplier)
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

func MakeAllTickersVxRequests(urls []*url.URL) <-chan []structs.TickerVx {

	// we are already receiving the AggregatesBarsRequests (un-flattened) here, so the job is to send over this data
	// to the flattener
	// create a rate limiter to stop over-requesting
	rateLimiter := ratelimit.New(rateLimit)

	// create a channel to make sure all requests are not being thrown away, of the flattened type
	c := make(chan []structs.TickerVx, 100000000)
	prev := time.Now()

	// use WaitGroup to make things more smooth with channels
	var wg sync.WaitGroup

	// create a buffer of the waitGroup, of the same length as urls
	wg.Add(len(urls))

	// Create a progress bar, unlimited
	bar := progressbar.Default(-1, "Downloading...")

	for _, u := range urls {
		now := rateLimiter.Take()
		firstResponse := new(structs.TickersVxResponse)

		// For each of their urls, i.e 500 of each of these things
		//bar2 := uiprogress.AddBar(500).AppendCompleted().PrependCompleted()
		go func(u *url.URL) {
			defer wg.Done()
			resp, err := http.Get(u.String())

			if err != nil {
				fmt.Println("Some Error: ", err)
				panic(err)
			} else {
				err = json.NewDecoder(resp.Body).Decode(&firstResponse)
				flattenedTarget := TickersVxFlattenPayloadBeforeInsert(*firstResponse)
				c <- flattenedTarget
				err := bar.Add(1)
				if err != nil {
					panic(err)
				}

				// First nextPagePath is a string, then it will be changed to URL type later
				nextPagePath := &firstResponse.NextUrl
				for *nextPagePath != "" {
					nextFlattenedTarget := new(structs.TickersVxResponse)
					// change the type to URL here
					nextPageURL := MakeTickersVxNextQueries(nextPagePath)
					nextResponse, err := http.Get(nextPageURL.String())
					if err != nil {
						fmt.Println("Some Error: ", err)
						panic(err)
					} else {
						err = json.NewDecoder(nextResponse.Body).Decode(&nextFlattenedTarget)
						// nextPagePath will be re-written here...
						nextPagePath = &nextFlattenedTarget.NextUrl
						// flatten the payload here as well
						nextFlattenedTarget := TickersVxFlattenPayloadBeforeInsert(*nextFlattenedTarget)
						c <- nextFlattenedTarget
						nextResponse.Body.Close()
					}
					err = bar.Add(1)
					if err != nil {
						panic(err)
					}
				}
			}
			resp.Body.Close()
		}(u)
		now.Sub(prev)
		prev = now
	}
	wg.Wait()
	close(c)
	return c
}

func MakeAllTickersRequests(urls []*url.URL) <-chan []structs.Tickers {

	// we are already receiving the AggregatesBarsRequests (un-flattened) here, so the job is to send over this data
	// to the flattener
	bar := progressbar.Default(int64(len(urls)), "Downloading Tickers...")

	// create a rate limiter to stop over-requesting
	rateLimiter := ratelimit.New(rateLimit)

	// create a channel to make sure all requests are not being thrown away, of the flattened type
	c := make(chan []structs.Tickers, len(urls))
	prev := time.Now()

	// use WaitGroup to make things more smooth with channels
	var wg sync.WaitGroup

	// create a buffer of the waitGroup, of the same length as urls
	wg.Add(len(urls))

	for _, u := range urls {
		now := rateLimiter.Take()
		target := new(structs.TickersResponse)

		go func(u *url.URL) {
			defer wg.Done()
			resp, err := http.Get(u.String())
			if err != nil {
				fmt.Println("Some Error: ", err)
				panic(err)
			} else {
				err = json.NewDecoder(resp.Body).Decode(&target)
				flattenedTarget := TickersFlattenPayloadBeforeInsert(*target)
				c <- flattenedTarget
			}
			resp.Body.Close()
		}(u)

		now.Sub(prev)
		prev = now

		var barerr = bar.Add(1)
		if barerr != nil {
			fmt.Println("\nSomething wrong with bar: ", barerr)
		}
	}
	wg.Wait()
	close(c)
	return c
}
