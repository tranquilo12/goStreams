package db

import (
	"context"
	"fmt"
	qdb "github.com/questdb/go-questdb-client"
	"github.com/schollz/progressbar/v3"
	"net/url"
	"path"
)

const (
	aggsHost = "api.polygon.io/v2/aggs"
	//tickerTypesHost = "api.polygon.io/v3/reference/types"
	tickersHost = "api.polygon.io/v3/reference/tickers"
	//tickerNews2Host = "api.polygon.io/v3/reference/news"

	//TimeLayout for every time layout
	TimeLayout = "2006-01-02" // go uses this date as a format specifier

	//tickerDetailsHost = "api.polygon.io/v1/meta"
	//dailyOpenCloseHost   = "api.polygon.io/v1/open-close"
	//groupedDailyBarsHost = "api.polygon.io/v2/aggs/grouped/locale/us/market/stocks"
)

// MakeStocksAggUrl A function that makes urls like:
// /v2/aggs/ticker/{stocksTicker}/range/{multiplier}/{timespan}/{from}/{to}
func MakeStocksAggUrl(
	stocksTicker string,
	multiplier string,
	timespan string,
	from_ string,
	to_ string,
	apiKey string,
	adjusted int,
) *url.URL {
	p, err := url.Parse("https://" + aggsHost)
	CheckErr(err)

	// Make the entire path
	p.Path = path.Join(p.Path, "ticker", stocksTicker, "range", multiplier, timespan, from_, to_)

	// make the url values
	q := url.Values{}
	if adjusted == 1 {
		q.Add("adjusted", "true")
	} else {
		q.Add("adjusted", "false")
	}
	q.Add("sort", "asc")
	q.Add("apiKey", apiKey)
	p.RawQuery = q.Encode()

	return p
}

// PushAllUrlsToTable A function that takes in a list of urls and pushes them to the database.
func PushAllUrlsToTable(
	ctx context.Context,
	tickers []string,
	timespan string,
	from_ string,
	to_ string,
	apiKey string,
	adjusted int,
) {
	// Get a progress bar
	pbar := progressbar.Default(int64(len(tickers)))

	// no need for channels in this yet, just a quick function that makes all the queries and sends it back
	fmt.Println("-	Making and pushing all urls to db...")

	// First create all the date pairs required
	datePairs := CreateDatePairs(from_, to_)

	// Connect to QDB and get sender
	sender, _ := qdb.NewLineSender(ctx)

	// Now for each ticker, and for each of the datePairs above, make urls.
	for _, ticker := range tickers {
		for _, dp := range *datePairs {
			// Make the urls here
			u := MakeStocksAggUrl(
				ticker,
				"1",
				timespan,
				dp.Start.Format(TimeLayout),
				dp.End.Format(TimeLayout),
				apiKey,
				adjusted,
			)

			// Push to db
			err := sender.Table("urls").
				Symbol("ticker", ticker).
				Int64Column("start", dp.Start.UnixMicro()).
				StringColumn("url", u.String()).
				BoolColumn("retry", false).
				At(ctx, dp.End.UnixNano())
			if err != nil {
				panic(err)
			}
		}

		// Update the progress bar
		pbar.Add(1)
	}

	// Make sure that the messages are sent over the network, for each ticker.
	err := sender.Flush(ctx)
	if err != nil {
		panic(err)
	}

	// Close the sender
	sender.Close()

	// Some cleanup
	err = pbar.Finish()
	if err != nil {
		fmt.Println(err)
	}

	// Just signal that it's done.
	fmt.Println("-	Done...")
}

// MakeTickerURL A function that takes the API string and time, and generates an url.
func MakeTickerURL(apiKey string) *url.URL {
	p, err := url.Parse("https://" + tickersHost)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	// make the url values
	q := url.Values{}
	q.Add("active", "true")
	q.Add("limit", "1000")
	q.Add("apiKey", apiKey)
	p.RawQuery = q.Encode()

	return p
}
