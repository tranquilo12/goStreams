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
	//apiKey     = "UPUKTfaMrhV1k5ZyCvBUbv_1pAjZ24zkUbLD_T"
	layout = "2006-01-02"
)

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

func MakeStocksAggsRequest(url *url.URL, target interface{}, c chan string, wg *sync.WaitGroup, connPool *pgxpool.Pool) error {
	defer (*wg).Done()

	resp, err := http.Get(url.String())
	if err != nil {
		c <- "We could not reach url: " + url.String()
	} else {
		c <- "We have reached url: " + url.String()
	}
	defer resp.Body.Close()

	var val = json.NewDecoder(resp.Body).Decode(&target)
	go db.PushIntoDB(val, connPool)
	return val
}

func MakeAllStocksAggsRequests(urls []*url.URL, bar *progressbar.ProgressBar, ratelimiter ratelimit.Limiter, connPool *pgxpool.Pool) {
	c := make(chan string)
	var wg sync.WaitGroup
	var i int

	prev := time.Now()
	for _, u := range urls {
		i += 1
		wg.Add(1)
		var err = bar.Add(1)
		if err != nil {
			fmt.Println("Some Error with the progress bar: ", err)
		}

		now := ratelimiter.Take()
		target := new(structs.StocksAggResponseParams)
		go MakeStocksAggsRequest(u, target, c, &wg, connPool)

		// add the function to push data into db here
		now.Sub(prev)
		prev = now
	}

	// this function literal (also called 'anonymous function' or 'lambda expression' in other languages)
	// is useful because 'go' needs to prefix a function and we can save some space by not declaring a whole new func
	go func() {
		wg.Wait()
		close(c)
	}()

	// this shorthand loop is syntactic sugar for an endless loop that just waits for results to come in through the 'c' channel
	for msg := range c {
		fmt.Println(msg)
	}

}

func main() {
	var urls []*url.URL
	var count64 int64
	rateLimiter := ratelimit.New(1)

	// Read all the equities into a list, grab the length
	equitiesList := config.ReadEquitiesList()
	count := len(equitiesList)
	count64 = int64(count)

	// Convert all dates to a PROPER format
	f_, _ := time.Parse(layout, "2015-01-01")
	var from_ = f_.Format(layout)
	t_, _ := time.Parse(layout, "2020-11-22")
	var to_ = t_.Format(layout)

	// Make all urls, dont do it on the fly
	// Set polygonIo cred
	polygonApiKey := config.SetPolygonCred("me")
	urls = MakeAllStocksAggsQueries(equitiesList, "day", from_, to_, polygonApiKey)

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

	// I need to see a progress bar!
	bar := progressbar.Default(count64, "Downloading")
	MakeAllStocksAggsRequests(urls, bar, rateLimiter, connPool)
}
