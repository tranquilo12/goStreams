package db

import (
	"fmt"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"
)

const (
	//Scheme               = "https" // universal for all schemes here
	aggsHost        = "api.polygon.io/v2/aggs"
	tickerTypesHost = "api.polygon.io/v2/reference/types"
	tickersVxHost   = "api.polygon.io/vX/reference/tickers"
	tickersHost     = "api.polygon.io/v2/reference/tickers"
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
	var urls []*url.URL
	for _, ticker := range tickers {
		u := MakeAggQueryStr(ticker, "1", timespan, from_, to_, apiKey)
		urls = append(urls, u)
	}
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
func MakeTickerVxQuery(apiKey string, date time.Time) *url.URL {
	p, err := url.Parse("https://" + tickersVxHost)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	// make the url values
	q := url.Values{}
	q.Add("date", date.Format("2006-01-02"))
	q.Add("active", "true")
	q.Add("limit", "500")
	q.Add("apiKey", apiKey)
	p.RawQuery = q.Encode()

	return p
}

// MakeAllTickersVxSourceQueries A function that takes a range of dates and makes a series of "Source" urls,
// that can then be passed on to the 'MakeTickersVxNextQueries' to get the next queries string(s).
func MakeAllTickersVxSourceQueries(apiKey string, startDate time.Time, endDate time.Time) []*url.URL {
	var urls []*url.URL
	for d := startDate; d.After(endDate) == false; d = d.AddDate(0, 0, 1) {
		u := MakeTickerVxQuery(apiKey, d)
		urls = append(urls, u)
	}
	return urls
}

// MakeTickersVxNextQueries A function that take the "next" page url from the first 'MakeTickersVxQuery' result
// and extracts the nextPagePathString, to make the new url.
func MakeTickersVxNextQueries(nextPagePath *string) *url.URL {
	nextPagePathArr := strings.Split(*nextPagePath, "/")
	nextPagePathString := strings.Join(nextPagePathArr[3:], "/")

	// trim the tickers from the end
	tickersVxHost := strings.Split(tickersVxHost, "/")
	tickersVxHost2 := strings.Join(tickersVxHost[:3], "/")

	p, err := url.Parse("https://" + tickersVxHost2 + "/" + nextPagePathString)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	values, _ := url.ParseQuery(p.RawQuery)
	values.Set("limit", "500")
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
