package db

import (
	"fmt"
	"lightning/utils/structs"
	"net/url"
	"path"
	"time"
)

const (
	//Scheme               = "https" // universal for all schemes here
	aggsHost          = "api.polygon.io/v3/aggs"
	tickerTypesHost   = "api.polygon.io/v3/reference/types"
	tickersVxHost     = "api.polygon.io/v3/reference/tickers"
	tickerNews2Host   = "api.polygon.io/v3/reference/news"
	tickerDetailsHost = "api.polygon.io/v1/meta"
	//dailyOpenCloseHost   = "api.polygon.io/v1/open-close"
	//groupedDailyBarsHost = "api.polygon.io/v2/aggs/grouped/locale/us/market/stocks"

	layout = "2006-01-02" // go uses this date as a format specifier
	//timespan   = "day"
	//multiplier = 1
	//from_      = "2020-11-22"
	//to_        = "2020-12-15"
)

// MakeTickerDetailsQuery generate all the queries for this endpoint
func MakeTickerDetailsQuery(apiKey string, ticker string) *url.URL {
	p, err := url.Parse("https://" + tickerDetailsHost + "/symbols/" + ticker + "/company")
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	// make the url values
	q := url.Values{}
	q.Add("apiKey", apiKey)
	p.RawQuery = q.Encode()
	return p
}

func MakeAllTickerDetailsQueries(apiKey string, allTickers []string) []*url.URL {
	var allUrls []*url.URL
	for _, ticker := range allTickers {
		allUrls = append(allUrls, MakeTickerDetailsQuery(apiKey, ticker))
	}
	return allUrls
}

// CreateLinearDatePairs creates a slice of arrays [[date, date+1], [date+1, date+2]...]
func CreateLinearDatePairs(from string, to string) []structs.StartEndDateStruct {
	startDate, _ := time.Parse("2006-01-02", from)
	endDate, _ := time.Parse("2006-01-02", to)

	var dateList []structs.StartEndDateStruct

	prevDate := startDate
	prevDatePlusOne := prevDate.Add(time.Hour * 24)

	for {

		if prevDatePlusOne.Before(endDate) {

			dp := structs.StartEndDateStruct{
				Start: prevDate.Format(layout),
				End:   prevDatePlusOne.Format(layout),
			}

			dateList = append(dateList, dp)
			//fmt.Printf("start: %s, end: %s...\n", prevDate.Format(layout), prevDatePlusOne.Format(layout))

			// make sure prevDate is set again
			prevDate = prevDatePlusOne
			prevDatePlusOne = prevDatePlusOne.Add(time.Hour * 24)
		} else {
			break
		}
	}
	return dateList
}

// MakeStocksAggUrl A function that makes urls like: /v2/aggs/ticker/{stocksTicker}/range/{multiplier}/{timespan}/{from}/{to}
func MakeStocksAggUrl(stocksTicker string, multiplier string, timespan string, from_ string, to_ string, apiKey string, adjusted int) *url.URL {
	p, err := url.Parse("https://" + aggsHost)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	// Make the entire path
	p.Path = path.Join(p.Path, "ticker", stocksTicker, "range", multiplier, timespan, from_, to_)

	// make the url values
	q := url.Values{}
	if adjusted == 1 {
		q.Add("unadjusted", "false")
	} else {
		q.Add("unadjusted", "true")
	}
	q.Add("sort", "asc")
	q.Add("apiKey", apiKey)
	p.RawQuery = q.Encode()

	return p
}

// MakeAllStocksAggsUrls A quick function that uses MakeStocksAggUrl and iterates through combos and returns a list of urls that will be queried.
func MakeAllStocksAggsUrls(tickers []string, timespan string, from_ string, to_ string, apiKey string, withLinearDates int, adjusted int) []*url.URL {
	// no need for channels in this yet, just a quick function that makes all the queries and sends it back
	fmt.Println("Making all urls...")

	var urls []*url.URL
	if withLinearDates == 1 {
		datePairs := CreateLinearDatePairs(from_, to_)
		for _, ticker := range tickers {
			for _, dp := range datePairs {
				u := MakeStocksAggUrl(ticker, "1", timespan, dp.Start, dp.End, apiKey, adjusted)
				urls = append(urls, u)
			}
		}
	} else {
		for _, ticker := range tickers {
			u := MakeStocksAggUrl(ticker, "1", timespan, from_, to_, apiKey, adjusted)
			urls = append(urls, u)
		}

	}

	fmt.Println("Done...")
	return urls
}

// MakeTickerTypesUrl A function that takes the API Key and generates the TickerTypes host.
func MakeTickerTypesUrl(apiKey string) *url.URL {
	p, err := url.Parse("https://" + tickerTypesHost)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	// make the url values
	q := url.Values{}
	q.Add("apiKey", apiKey)
	p.RawQuery = q.Encode()

	return p
}

// MakeTickerVxQuery A function that takes the API string and time, and generates a url.
func MakeTickerVxQuery(apiKey string) *url.URL {
	p, err := url.Parse("https://" + tickersVxHost)
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

// MakeTickerNews2Query A function that takes in the apikey + page number to make urls.
func MakeTickerNews2Query(apiKey string, ticker string, from_ string) *url.URL {
	p, err := url.Parse("https://" + tickerNews2Host)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	// make the url values
	q := url.Values{}
	q.Add("ticker", ticker)
	q.Add("limit", "1000")
	q.Add("apiKey", apiKey)
	q.Add("published_utc.gte", from_)
	p.RawQuery = q.Encode()
	return p
}
