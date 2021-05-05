package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/pgxpool"
	"github.com/schollz/progressbar"
	"go.uber.org/ratelimit"
	"lightning/utils/config"
	"lightning/utils/db"
	"lightning/utils/structs"
	"net/http"
	"net/url"
	"path"
	"sync"
	"time"
)

const (
	aggsScheme = "https"
	aggsHost   = "api.polygon.io/v2/aggs"
	layout     = "2006-01-02"
	rateLimit  = 50
	timespan   = "day"
	multiplier = 1
)

func ParseTime(date string) string {
	// Convert all dates to a PROPER format
	f_, _ := time.Parse(layout, date)
	d := f_.Format(layout)
	return d
}

func MakeStocksAggsQueryStr(ticker string, multiplier string, timespan string, from_ string, to_ string, apiKey string) *url.URL {
	p, err := url.Parse(aggsScheme + "://" + aggsHost)
	if err != nil {
		fmt.Println(err)
	}
	p.Path = path.Join(p.Path, "ticker", ticker, "range", multiplier, timespan, from_, to_)
	q := url.Values{}
	q.Add("unadjusted", "true")
	q.Add("sort", "asc")
	q.Add("apiKey", apiKey)
	p.RawQuery = q.Encode()
	return p
}

func MakeAllStocksAggsQueries(tickers []string, timespan string, from_ string, to_ string, apiKey string) []*url.URL {
	var urls []*url.URL
	for _, ticker := range tickers {
		u := MakeStocksAggsQueryStr(ticker, "1", timespan, from_, to_, apiKey)
		urls = append(urls, u)
	}
	return urls
}

func MakeAllStocksAggsRequests(urls []*url.URL) <-chan structs.StocksAggResponseParams {

	rateLimiter := ratelimit.New(rateLimit)
	c := make(chan structs.StocksAggResponseParams, len(urls))
	bar := progressbar.Default(int64(len(urls)), "Downloading")
	prev := time.Now()

	go func() {
		var wg sync.WaitGroup
		wg.Add(len(urls))

		for _, u := range urls {
			err := bar.Add(1)
			if err != nil {
				fmt.Println("Some Error with the progress bar: ", err)
			}

			now := rateLimiter.Take()
			target := new(structs.StocksAggResponseParams)

			go func(u *url.URL) {
				defer wg.Done()
				resp, err := http.Get(u.String())
				if err != nil {
					fmt.Println("Some error: ", err)
				} else {
					err = json.NewDecoder(resp.Body).Decode(&target)
					c <- *target
				}
				err = resp.Body.Close()
				if err != nil {
					return
				}
			}(u)

			now.Sub(prev)
			prev = now
		}

		wg.Wait()
		close(c)
	}()

	return c
}

func main() {
	var urls []*url.URL

	// Read all the equities into a list, grab the length
	equitiesList := config.ReadEquitiesList()

	// Convert all dates to a PROPER format
	from_ := ParseTime("2015-01-01")
	to_ := ParseTime("2020-11-22")

	// Make all urls, dont do it on the fly
	// Set polygonIo cred
	polygonApiKey := config.SetPolygonCred("me")
	urls = MakeAllStocksAggsQueries(equitiesList, timespan, from_, to_, polygonApiKey)

	// Set DB params and make a connection pool
	postgresParams := new(config.DbParams)
	err := config.SetDBParams(postgresParams, "postgres")
	if err != nil {
		fmt.Println(err)
	}
	postgresConnStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		postgresParams.User,
		postgresParams.Password,
		postgresParams.Host,
		postgresParams.Port,
		postgresParams.Dbname,
	)

	poolConfig, err := pgxpool.ParseConfig(postgresConnStr)
	if err != nil {
		fmt.Println("Unable to create config str...", err)
	}

	connPool, err := pgxpool.ConnectConfig(context.Background(), poolConfig)
	if err != nil {
		fmt.Println("Unable to create conn...", err)
	}
	defer connPool.Close()

	c := MakeAllStocksAggsRequests(urls)

	bar2 := progressbar.Default(int64(len(urls)), "Uploading")
	for payload := range c {
		db.PushIntoDB(payload, connPool, timespan, multiplier, layout)
		err2 := bar2.Add(1)
		if err2 != nil {
			fmt.Println("Some Error with the progress bar: ", err2)
		}
	}

}
