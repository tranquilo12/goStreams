package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/schollz/progressbar/v3"
	"go.uber.org/ratelimit"
	"lightning/utils/config"
	"lightning/utils/db"
	"lightning/utils/structs"
	"net/http"
	"net/url"
	"os"
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

func MakeAllStocksAggsRequests(urls []*url.URL, bar *progressbar.ProgressBar) <-chan structs.StocksAggResponseParams {

	rateLimiter := ratelimit.New(rateLimit)
	c := make(chan structs.StocksAggResponseParams, len(urls))
	prev := time.Now()

	go func() {
		var wg sync.WaitGroup
		wg.Add(len(urls))

		for _, u := range urls {
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
				resp.Body.Close()
			}(u)

			now.Sub(prev)
			prev = now
			var err = bar.Add(1)
			if err != nil {
				fmt.Println("\nSomething wrong with bar1: ", err)
			}
		}

		wg.Wait()
		close(c)
	}()

	return c
}

func PushGiantPayloadIntoDB(output []structs.ExpandedStocksAggResponseParams, connPool *pgxpool.Pool) {

	batch := &pgx.Batch{}
	numInserts := len(output)
	for k := range output[0 : numInserts-1] {
		batch.Queue(db.PolygonStocksAggCandlesInsertTemplate,
			output[k].Ticker,
			output[k].Timespan,
			output[k].Multiplier,
			output[k].V,
			output[k].Vw,
			output[k].O,
			output[k].C,
			output[k].H,
			output[k].L,
			output[k].T)
	}

	// pull through the batch and exec each statement
	br := connPool.SendBatch(context.Background(), batch)
	for k := 0; k < numInserts-1; k++ {
		_, err := br.Exec()
		if err != nil {
			fmt.Println("Unable to execute statement in batched queue: ", err)
			os.Exit(1)
		}
	}

	// Close this batch pool
	var err = br.Close()
	if err != nil {
		fmt.Println("Unable to close batch: ", err)
	}
}

func PushBatchedPayloadIntoDB(output []structs.ExpandedStocksAggResponseParams, connPool *pgxpool.Pool, batchSize int) {
	var i int
	var j int
	batch := &pgx.Batch{}
	numInserts := len(output)

	bar := progressbar.Default(int64(numInserts/batchSize), "Uploading...")
	for i = 0; i < numInserts-1; i += batchSize {
		if ((numInserts - 1) - i) == 1 {
			j = i + 1
		} else {
			j = i + batchSize
		}

		for k := range output[i:j] {
			batch.Queue(db.PolygonStocksAggCandlesInsertTemplate,
				output[k].Ticker,
				output[k].Timespan,
				output[k].Multiplier,
				output[k].V,
				output[k].Vw,
				output[k].O,
				output[k].C,
				output[k].H,
				output[k].L,
				output[k].T)
		}

		// pull through the batch and exec each statement
		br := connPool.SendBatch(context.Background(), batch)
		for k := 0; k < (j - i); k++ {
			_, err := br.Exec()
			if err != nil {
				fmt.Println("Unable to execute statement in batched queue: ", err)
				os.Exit(1)
			}
		}

		// Close this batch pool
		var err = br.Close()
		if err != nil {
			fmt.Println("Unable to close batch: ", err)
		}

		err = bar.Add(1)
		if err != nil {
			fmt.Println("Something wrong with inserting batches bar ", err)
		}
	}
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

	bar1 := progressbar.Default(int64(len(urls)), "Downloading...")
	c := MakeAllStocksAggsRequests(urls, bar1)

	bar2 := progressbar.Default(int64(len(urls)), "Flattening...")
	var bigPayload []structs.ExpandedStocksAggResponseParams

	for payload := range c {
		output := db.FlattenPayloadBeforeInsert(payload, timespan, multiplier, layout)

		if len(output) > 0 {
			//	PushGiantPayloadIntoDB(output, connPool)
			bigPayload = append(bigPayload, output...)
		}

		err = bar2.Add(1)
		if err != nil {
			fmt.Println("\nSomething wrong with bar2: ", err)
		}
	}

	PushBatchedPayloadIntoDB(bigPayload, connPool, 5000)

}
