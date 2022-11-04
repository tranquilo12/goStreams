package publisher

import (
	"context"
	"encoding/json"
	"fmt"
	qdb "github.com/questdb/go-questdb-client"
	"go.uber.org/ratelimit"
	"lightning/utils/config"
	"lightning/utils/db"
	"lightning/utils/structs"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// UpdateUrlsRetry updates the urls that failed to download
func UpdateUrlsRetry(ctx context.Context, u string) {
	// Connect to QDB
	conn := db.QDBConnectPG(ctx)
	defer conn.Close()

	// Form the query
	query := fmt.Sprintf("UPDATE 'urls' set retry = true where url = '%s';", u)

	// Begin the transaction, execute the query, and commit the transaction
	tx, err := conn.Begin(ctx)
	db.CheckErr(err)

	_, err = tx.Exec(ctx, query)
	db.CheckErr(err)

	err = tx.Commit(ctx)
	db.CheckErr(err)
}

// DownloadFromPolygonIO downloads the prices from PolygonIO
func DownloadFromPolygonIO(
	client *http.Client,
	u url.URL,
	res *structs.AggregatesBarsResponse,
) error {
	// Get a context that can be cancelled within this function
	ctx, cancel := context.WithCancel(context.Background())

	// Create a new client
	resp, err := client.Get(u.String())
	if err != nil {
		UpdateUrlsRetry(ctx, u.String())
	} else {
		defer resp.Body.Close()
	}

	// Decode the response
	if resp.StatusCode == http.StatusOK {
		if err = json.NewDecoder(resp.Body).Decode(&res); err != nil {
			UpdateUrlsRetry(ctx, u.String())
		}
	} else {
		UpdateUrlsRetry(ctx, u.String())
	}

	cancel()
	return err
}

// AggChannelWriter writes the aggregates to Kafka
func AggChannelWriter(
	urls []string,
) error {
	// use WaitGroup to make things more smooth with goroutines
	var wg sync.WaitGroup

	// create a buffer of the waitGroup, of the same length as urls
	wg.Add(len(urls))

	// Get the http client, it's a modified version with extended timeout and other features.
	httpClient := config.GetHttpClient()

	// Get the channel that will be used to write to QuestDB
	c := make(chan structs.AggregatesBarsResponse, len(urls))

	// Add another wait group for the goroutine below
	wg.Add(1)

	// Make a goroutine that will accept data from a channel and push to questDB
	ctx := context.TODO()

	// Get newline sender
	sender, err := qdb.NewLineSender(ctx)
	db.CheckErr(err)

	// Go Func to insert data into the database, reads from the channel
	go func() {
		// Makes sure wg closes
		defer wg.Done()

		// Get the values from the channel
		for res := range c {
			if res.Results != nil {
				for _, v := range res.Results {
					err := sender.Table("aggs").
						Symbol("ticker", res.Ticker).
						StringColumn("timespan", "minute").
						Int64Column("multiplier", int64(1)).
						Float64Column("open", v.O).
						Float64Column("high", v.H).
						Float64Column("low", v.L).
						Float64Column("close", v.C).
						Float64Column("volume", v.V).
						Float64Column("vw", v.Vw).
						Float64Column("n", float64(v.N)).
						At(ctx, time.UnixMilli(int64(v.T)).UnixNano())
					db.CheckErr(err)
				}

				// Make sure the sender is flushed
				err := sender.Flush(ctx)
				db.CheckErr(err)
			}
		}
	}()

	// Max allow 1000 requests per second
	prev := time.Now()
	rateLimiter := ratelimit.New(50000)

	// Iterate over every url, create goroutine for each url download and rate limit the requests
	for _, u := range urls {
		now := rateLimiter.Take()
		go channelWriter(c, httpClient, u, &wg)
		now.Sub(prev)
		prev = now
	}

	// Wait for all the goroutines to finish, and close the channel, and then the sender.
	go func() {
		// Wait for all the goroutines to finish
		wg.Wait()

		// Close the channel
		close(c)

		// Close the sender
		if err = sender.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	return nil
}

func channelWriter(
	chan1 chan structs.AggregatesBarsResponse,
	httpClient *http.Client,
	u string,
	wg *sync.WaitGroup,
) {
	// Makes sure wg closes
	defer wg.Done()

	// Convert the u(string) to a *url.URL
	FinalUrl, err := url.Parse(u)
	db.CheckErr(err)

	// Download the data from PolygonIO
	var res structs.AggregatesBarsResponse
	err = DownloadFromPolygonIO(httpClient, *FinalUrl, &res)
	if err != nil {
		// Send a nil results to the channel
		chan1 <- structs.AggregatesBarsResponse{Results: nil}
	} else {
		// Send the data to QDB, if response is not empty
		if res.Results != nil {
			chan1 <- res
		} else {
			// Send a nil results to the channel
			chan1 <- structs.AggregatesBarsResponse{Results: nil}
		}
	}
}
