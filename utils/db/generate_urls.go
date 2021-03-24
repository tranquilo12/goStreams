package db

import (
	"fmt"
	"net/url"
	"path"
)

const (
	//Scheme               = "https" // universal for all schemes here
	aggsHost = "api.polygon.io/v2/aggs"
	//dailyOpenCloseHost   = "api.polygon.io/v1/open-close"
	//groupedDailyBarsHost = "api.polygon.io/v2/aggs/grouped/locale/us/market/stocks"

	//layout     = "2006-01-02" // go uses this date as a format specifier
	rateLimit = 50 // can be changed
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

// aggregates bars url: /v2/aggs/ticker/{stocksTicker}/range/{multiplier}/{timespan}/{from}/{to}
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

func MakeAllStocksAggsQueries(tickers []string, timespan string, from_ string, to_ string, apiKey string) []*url.URL {
	var urls []*url.URL
	for _, ticker := range tickers {
		u := MakeAggQueryStr(ticker, "1", timespan, from_, to_, apiKey)
		urls = append(urls, u)
	}
	return urls
}
