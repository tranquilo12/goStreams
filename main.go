package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/pgxpool"
	"github.com/vbauerster/mpb/decor"

	"github.com/vbauerster/mpb"
	"go.uber.org/ratelimit"
	"lightning/utils/config"
	"lightning/utils/db"
	"lightning/utils/structs"
	"math/rand"
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

func MakeAllStocksAggsRequests(urls []*url.URL, p *mpb.Progress) <-chan structs.StocksAggResponseParams {

	rateLimiter := ratelimit.New(rateLimit)
	c := make(chan structs.StocksAggResponseParams, len(urls))
	prev := time.Now()

	go func() {
		var wg sync.WaitGroup
		wg.Add(len(urls))

		for i, u := range urls {
			task := fmt.Sprintf("Url#%02d:", i)
			job := "downloading"
			b := p.AddBar(rand.Int63n(201)+100,
				mpb.PrependDecorators(
					decor.Name(task, decor.WC{W: len(task) + 1, C: decor.DidentRight}),
					decor.Name(job, decor.WCSyncSpaceR),
					decor.CountersNoUnit("%d / %d", decor.WCSyncWidth),
				),
				mpb.AppendDecorators(decor.Percentage(decor.WC{W: 5})),
			)

			now := rateLimiter.Take()
			target := new(structs.StocksAggResponseParams)

			go func(u *url.URL, bar *mpb.Bar, incr int) {
				defer wg.Done()
				for !bar.Completed() {
					start := time.Now()

					resp, err := http.Get(u.String())
					if err != nil {
						fmt.Println("Some error: ", err)
					} else {
						err = json.NewDecoder(resp.Body).Decode(&target)
						c <- *target
					}
					resp.Body.Close()

					bar.IncrBy(incr)
					bar.DecoratorEwmaUpdate(time.Since(start))
				}
			}(u, b, i+1)

			now.Sub(prev)
			prev = now
			//bars = append(bars, b)
		}

		p.Wait()
		close(c)
	}()

	return c
}

func main() {
	var allOutputs []structs.ExpandedStocksAggResponseParams
	var urls []*url.URL
	j := 0

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

	DownloadingWg := new(sync.WaitGroup)
	p := mpb.New(mpb.WithWaitGroup(DownloadingWg))
	c := MakeAllStocksAggsRequests(urls, p)

	for payload := range c {

		// ANSI escape sequences are not supported on Windows OS
		task := fmt.Sprintf("Flatten Array#%02d:", j)
		job := fmt.Sprintf("Quickly...")

		// iterate up
		j += 1

		// preparing delayed bars
		b := p.AddBar(rand.Int63n(101)+100,
			//mpb.BarQueueAfter(bars[j]),
			mpb.BarFillerClearOnComplete(),
			mpb.PrependDecorators(
				decor.Name(task, decor.WC{W: len(task) + 1, C: decor.DidentRight}),
				decor.OnComplete(decor.Name(job, decor.WCSyncSpaceR), "done!"),
				decor.OnComplete(decor.EwmaETA(decor.ET_STYLE_MMSS, 0, decor.WCSyncWidth), ""),
			),
			mpb.AppendDecorators(
				decor.OnComplete(decor.Percentage(decor.WC{W: 5}), ""),
			),
		)

		for !b.Completed() {
			start := time.Now()
			output := db.FlattenPayloadBeforeInsert(payload, timespan, multiplier, layout)
			allOutputs = append(
				allOutputs,
				output...,
			)
			b.IncrBy(j + 1)
			b.DecoratorEwmaUpdate(time.Since(start))
		}

	}

	db.PushGiantPayloadIntoDB(allOutputs, connPool, p)
}
