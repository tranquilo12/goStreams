package db

import (
	"fmt"
	"net/url"
	"path"
	"strconv"
	"time"
)

const (
	//Scheme               = "https" // universal for all schemes here
	aggsHost        = "api.polygon.io/v2/aggs"
	tickerTypesHost = "api.polygon.io/v2/reference/types"
	tickersVxHost   = "api.polygon.io/v3/reference/tickers"
	tickersHost     = "api.polygon.io/v2/reference/tickers"
	tickerNews2Host = "api.polygon.io/v2/reference/news"
	//dailyOpenCloseHost   = "api.polygon.io/v1/open-close"
	//groupedDailyBarsHost = "api.polygon.io/v2/aggs/grouped/locale/us/market/stocks"

	layout    = "2006-01-02" // go uses this date as a format specifier
	rateLimit = 50           // can be changed
	//timespan   = "day"
	//multiplier = 1
	//from_      = "2020-11-22"
	//to_        = "2020-12-15"
)

// daily open close url: /v1/open-close/{stocksTicker}/{date}
//func MakeDailyOpenCloseStr(stocksTicker string, date string, apiKey string) *url.URL {
//	p, err := url.Parse(Scheme + "://" + dailyOpenCloseHost)
//	if err != nil {
//		fmt.Println(err)
//		panic(err)
//	}
//
//	// Make the entire path
//	p.Path = path.Join(p.Path, stocksTicker, date)
//
//	// make the url values
//	q := url.Values{}
//	q.Add("unadjusted", "true")
//	q.Add("sort", "asc")
//	q.Add("apiKey", apiKey)
//	p.RawQuery = q.Encode()
//
//	return p
//}

//func MakeDailyOpenCloseQueries(tickers []string, date string, apiKey string)[]*url.URL{
//	var urls []*url.URL
//	for _, ticker := range tickers {
//		u := MakeDailyOpenCloseStr(ticker, date, apiKey)
//		urls = append(urls, u)
//	}
//	return urls
//}

// grouped daily bars url: /v2/aggs/grouped/locale/us/market/stocks/{date}
//func MakeGroupedDailyBarsStr(date string, apiKey string) *url.URL {
//	p, err := url.Parse(Scheme + "://" + groupedDailyBarsHost)
//	if err != nil {
//		fmt.Println(err)
//		panic(err)
//	}
//
//	// Make the entire path
//	p.Path = path.Join(p.Path, date)
//
//	// make the url values
//	q := url.Values{}
//	q.Add("unadjusted", "true")
//	q.Add("sort", "asc")
//	q.Add("apiKey", apiKey)
//	p.RawQuery = q.Encode()
//
//	return p
//}

//func MakeGroupedDailyBarsQueries(dates []string,  apiKey string)[]*url.URL{
//	var urls []*url.URL
//	for _, date := range dates {
//		u := MakeGroupedDailyBarsStr(date, apiKey)
//		urls = append(urls, u)
//	}
//	return urls
//}

type FromToDateStruct struct {
	From string
	To   string
}

//
func MakeDatePairs(from_ string, to_ string) *[]FromToDateStruct {
	f_, err := time.Parse(layout, from_)
	if err != nil {
		panic(err)
	}

	t_, err := time.Parse(layout, to_)
	if err != nil {
		panic(err)
	}

	var datePairs []FromToDateStruct
	for {
		if f_.Equal(t_) {
			return &datePairs
		}
		if f_.Weekday() != time.Saturday && f_.Weekday() != time.Sunday {
			ft := FromToDateStruct{
				From: f_.Format(layout),
				To:   t_.Format(layout),
			}

			datePairs = append(datePairs, ft)
			f_ = f_.Add(time.Hour * 24)
		}
	}
}

// MakeAggQueryStr A function that makes urls like: /v2/aggs/ticker/{stocksTicker}/range/{multiplier}/{timespan}/{from}/{to}
func MakeAggQueryStr(stocksTicker string, multiplier string, timespan string, from_ string, to_ string, apiKey string) *url.URL {
	p, err := url.Parse("https://" + aggsHost)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	// Make the entire path
	p.Path = path.Join(p.Path, "ticker", stocksTicker, "range", multiplier, timespan, from_, to_)

	// make the url values
	q := url.Values{}
	q.Add("unadjusted", "true")
	q.Add("sort", "asc")
	q.Add("apiKey", apiKey)
	p.RawQuery = q.Encode()

	return p
}

// MakeAllStocksAggsQueries A quick function that uses MakeAggQueryStr and iterates through combos and returns a
// list of urls that will be queried.
func MakeAllStocksAggsQueries(tickers []string, timespan string, from_ string, to_ string, apiKey string) []*url.URL {
	// no need for channels in this yet, just a quick function that makes all the queries and sends it back
	fmt.Println("Making all urls...")
	var urls []*url.URL

	datePairs := MakeDatePairs(from_, to_)

	for _, ticker := range tickers {
		for _, val := range *datePairs {
			u := MakeAggQueryStr(ticker, "1", timespan, val.From, val.To, apiKey)
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

// MakeTickersQuery A function that takes in the apikey + page number to make urls.
func MakeTickersQuery(apiKey string, page int) *url.URL {
	p, err := url.Parse("https://" + tickersHost)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	// make the url values
	q := url.Values{}
	q.Add("page", strconv.Itoa(page))
	q.Add("perpage", "50")
	q.Add("apiKey", apiKey)
	q.Add("locale", "g")
	p.RawQuery = q.Encode()
	return p
}

// MakeAllTickersQuery A function that iterates through 'MakeTickersQuery' using the numPages parameter.
func MakeAllTickersQuery(apiKey string, numPages int) []*url.URL {
	var urls []*url.URL
	for i := 1; i <= numPages; i++ {
		u := MakeTickersQuery(apiKey, i)
		urls = append(urls, u)
	}
	return urls
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
