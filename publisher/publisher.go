package publisher

import (
	"context"
	"encoding/json"
	qdb "github.com/questdb/go-questdb-client"
	"github.com/schollz/progressbar/v3"
	"github.com/sirupsen/logrus"
	"go.uber.org/ratelimit"
	"lightning/utils/config"
	"lightning/utils/db"
	"lightning/utils/structs"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// DownloadFromPolygonIO downloads the prices from PolygonIO
func DownloadFromPolygonIO(
	client *http.Client,
	logger *logrus.Logger,
	u url.URL,
	res *structs.AggregatesBarsResponse,
) error {
	// Create a new client
	resp, err := client.Get(u.String())
	if err != nil {
		logger.WithFields(logrus.Fields{
			"url":   u.String(),
			"error": err,
		}).Info("Error in downloading data from PolygonIO")
		return nil
	}

	// Defer the closing of the body
	defer resp.Body.Close()

	// Decode the response
	if resp.StatusCode == http.StatusOK {
		err = json.NewDecoder(resp.Body).Decode(&res)
	}
	return err
}

// AggChannelWriter writes the aggregates to Kafka
func AggChannelWriter(
	urls []string,
	logger *logrus.Logger,
) error {
	// use WaitGroup to make things more smooth with goroutines
	var wg sync.WaitGroup

	// create a buffer of the waitGroup, of the same length as urls
	wg.Add(len(urls))

	// Create ui progress bar, formatted
	bar := progressbar.Default(int64(len(urls)))
	defer bar.Close()

	// Get the http client
	httpClient := config.GetHttpClient()

	// Get the channel that will be used to write to QuestDB
	c := make(chan structs.AggregatesBarsResponse, len(urls))

	// Make a goroutine that will accept data from a channel and push to questDB
	ctx := context.Background()
	wg.Add(1)

	go func() {
		defer wg.Done()

		sender, _ := qdb.NewLineSender(ctx)
		defer sender.Close()

		// Get the values from the channel
		for res := range c {
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
	}()

	// Max allow 1000 requests per second
	prev := time.Now()
	rateLimiter := ratelimit.New(1000)

	// Iterate over every url, create goroutine for each url download and rate limit the requests
	for _, u := range urls {
		now := rateLimiter.Take()
		go channelWriter(c, logger, httpClient, u, &wg, bar)
		now.Sub(prev)
		prev = now
	}

	// Wait for all the goroutines to finish, and close the channel
	go func() {
		wg.Wait()
		close(c)
	}()

	return nil
}

func channelWriter(
	chan1 chan structs.AggregatesBarsResponse,
	logger *logrus.Logger,
	httpClient *http.Client,
	u string,
	wg *sync.WaitGroup,
	bar *progressbar.ProgressBar,
) {
	// Makes sure wg closes
	defer wg.Done()

	// Convert the u(string) to a *url.URL
	FinalUrl, err := url.Parse(u)
	db.CheckErr(err)

	// Download the data from PolygonIO
	var res structs.AggregatesBarsResponse
	err = DownloadFromPolygonIO(httpClient, logger, *FinalUrl, &res)
	db.CheckErr(err)

	// Send the data to QDB, if response is not empty
	if res.Results != nil {
		chan1 <- res
	}

	// Progress bar update
	err = bar.Add(1)
	db.CheckErr(err)
}
