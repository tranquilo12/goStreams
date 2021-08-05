package db

import (
	"encoding/json"
	"fmt"
	"github.com/go-pg/pg/v10"
	"github.com/go-redis/redis/v7"
	"github.com/schollz/progressbar/v3"
	"go.uber.org/ratelimit"
	"lightning/utils/config"
	"lightning/utils/structs"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

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

func MakeAllTickersVxRequests(u *url.URL) chan []structs.TickerVx {
	var vxResponse *structs.TickersVxResponse
	var response *http.Response
	var p *url.URL
	var err error
	apiKey := u.Query()["apiKey"]
	var newCursor string

	// create a channel to make sure all requests are not being thrown away, of the flattened type.
	c := make(chan []structs.TickerVx, 100000)

	response, err = http.Get(u.String())
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()

	j := 0
	i := 0
	for {
		if response.StatusCode == 200 {
			err = json.NewDecoder(response.Body).Decode(&vxResponse)
			if err != nil {
				panic(err)
			}

			flattenVxResponse := structs.TickersVxFlattenPayloadBeforeInsert(*vxResponse)
			c <- flattenVxResponse

			if vxResponse.NextUrl != "" {

				p, err = url.Parse(vxResponse.NextUrl)
				if err != nil {
					panic(err)
				}
				oldCursor := p.Query()["cursor"][0]

				q := p.Query()
				q.Add("apiKey", apiKey[0])
				p.Host = "api.polygon.io:443"
				p.RawQuery = q.Encode()

				//fmt.Println(p.String())
				response, err = http.Get(p.String())
				if err != nil {
					panic(err)
				}

				if i > 0 {
					p, err = url.Parse(vxResponse.NextUrl)
					if err != nil {
						panic(err)
					}
					newCursor = p.Query()["cursor"][0]
					if newCursor == oldCursor {
						j += 1
						if j > 20 {
							break
						}
					}
				}
				i += 1
			}
		} else {
			break
		}
	}
	close(c)
	return c
}

func MakeAllTickerDetailsRequestsAndPushToDB(urls []*url.URL, pgDB *pg.DB) error {

	// we are already receiving the AggregatesBarsRequests (un-flattened) here, so the job is to send over this data
	// to the flattener
	bar := progressbar.Default(int64(len(urls)), "Downloading and inserting Ticker Details...")

	// create a rate limiter to stop over-requesting
	rateLimiter := ratelimit.New(300)
	prev := time.Now()

	// use WaitGroup to make things more smooth with channels
	var wg sync.WaitGroup

	// create a buffer of the waitGroup, of the same length as urls
	wg.Add(len(urls))

	for _, u := range urls {
		now := rateLimiter.Take()
		target := new(structs.TickerDetails)

		go func() {
			resp, err := http.Get(u.String())

			if err != nil {
				fmt.Println("Some Error: ", err)
				panic(err)
			} else {
				defer resp.Body.Close()
				err = json.NewDecoder(resp.Body).Decode(&target)
				_, err := pgDB.Model(target).OnConflict("(symbol) DO NOTHING").Insert()
				if err != nil {
					panic(err)
				}
			}
			wg.Done()
		}()

		now.Sub(prev)
		prev = now

		var barerr = bar.Add(1)
		if barerr != nil {
			fmt.Println("\nSomething wrong with bar: ", barerr)
		}
	}
	wg.Wait()
	return nil
}

// GetAllTickers Just get a list of all the tickers that are present in "ticker_vxes"

func GetAllTickersFromPolygonioDirectly() *[]string {
	apiKey := config.SetPolygonCred("other")
	u := MakeTickerVxQuery(apiKey)
	Chan1 := MakeAllTickersVxRequests(u)
	strResult := GetTickerVxs(Chan1)
	strArrResults := strings.Split(strResult, ",")
	return &strArrResults
}

func GetAllTickersFromRedis(redisClient *redis.Client) *[]string {
	result := redisClient.Get("allTickers")
	strResult, err := result.Result()
	if err != nil {
		apiKey := config.SetPolygonCred("other")
		u := MakeTickerVxQuery(apiKey)
		Chan1 := MakeAllTickersVxRequests(u)
		err := PushTickerVxIntoRedis(Chan1, redisClient)
		if err != nil {
			panic(err)
		}

		result := redisClient.Get("allTickers")
		strResult, err = result.Result()
	}
	strArrResults := strings.Split(strResult, ",")
	return &strArrResults
}

func GetDifferenceBtwTickersInMemAndS3(slice1 []string, slice2 []string) []string {
	var diff []string

	// Loop two times, first to find slice1 strings not in slice2,
	// second loop to find slice2 strings not in slice1
	for i := 0; i < 2; i++ {
		for _, s1 := range slice1 {
			found := false
			for _, s2 := range slice2 {
				if s1 == s2 {
					found = true
					break
				}
			}
			// String not found. We add it to return slice
			if !found {
				diff = append(diff, s1)
			}
		}
		// Swap the slices, only if it was the first loop
		if i == 0 {
			slice1, slice2 = slice2, slice1
		}
	}

	return diff
}

func MakeAllTickerNews2Requests(u *url.URL) chan []structs.TickerNews2 {
	var News2Response *structs.TickerNews2Response
	var response *http.Response
	var p *url.URL
	var err error
	apiKey := u.Query()["apiKey"]
	var newCursor string

	// create a channel to make sure all requests are not being thrown away, of the flattened type.
	c := make(chan []structs.TickerNews2, 100000)

	response, err = http.Get(u.String())
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()

	j := 0
	i := 0
	for {
		if response.StatusCode == 200 {
			err = json.NewDecoder(response.Body).Decode(&News2Response)
			if err != nil {
				panic(err)
			}

			flattenVxResponse := structs.TickerNews2FlattenPayloadBeforeInsert(*News2Response)
			c <- flattenVxResponse

			if News2Response.NextURL != "" {

				p, err = url.Parse(News2Response.NextURL)
				if err != nil {
					panic(err)
				}
				oldCursor := p.Query()["cursor"][0]

				q := p.Query()
				q.Add("apiKey", apiKey[0])
				p.Host = "api.polygon.io:443"
				p.RawQuery = q.Encode()

				fmt.Println(p.String())
				response, err = http.Get(p.String())
				if err != nil {
					panic(err)
				}

				if i > 0 {
					p, err = url.Parse(News2Response.NextURL)
					if err != nil {
						panic(err)
					}
					newCursor = p.Query()["cursor"][0]
					if newCursor == oldCursor {
						j += 1
						if j > 20 {
							break
						}
					}
				}
				i += 1
			} else {
				break
			}
		} else {
			break
		}
	}
	close(c)
	return c
}
