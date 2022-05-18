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
	aggsHost        = "api.polygon.io/v2/aggs"
	tickerTypesHost = "api.polygon.io/v3/reference/types"
	tickersVxHost   = "api.polygon.io/v3/reference/tickers"
	tickerNews2Host = "api.polygon.io/v3/reference/news"
	//tickerDetailsHost = "api.polygon.io/v1/meta"
	//dailyOpenCloseHost   = "api.polygon.io/v1/open-close"
	//groupedDailyBarsHost = "api.polygon.io/v2/aggs/grouped/locale/us/market/stocks"

	layout = "2006-01-02" // go uses this date as a format specifier
	//timespan   = "day"
	//multiplier = 1
	//from_      = "2020-11-22"
	//to_        = "2020-12-15"
)

// CreateLinearDatePairs creates a slice of arrays [[date, date+1], [date+1, date+2]...]
// gap has been added to the params, so the gap between the dates is defined by the gap * time.Hour
// So it has to be a multiple of 24
func CreateLinearDatePairs(from string, to string, gap int) []structs.StartEndDateStruct {
	// Create a slice of structs
	startDate, _ := time.Parse("2006-01-02", from)
	endDate, _ := time.Parse("2006-01-02", to)

	// Create the dateList slice
	var dateList []structs.StartEndDateStruct

	// Only allow for gaps that are multiples of 24
	if gap%24 == 0 {
		// Instantiate the prevDate and prevDatePlusOne
		// For 50K minutes between time gaps, gap = 816
		gapHours := time.Duration(gap) * time.Hour
		prevDate := startDate
		prevDatePlusOne := prevDate.Add(gapHours)

		for {
			// If the next date is after the end date, break
			if prevDatePlusOne.Before(endDate) {
				// Make sure start and end are well-defined
				dp := structs.StartEndDateStruct{
					Start: prevDate.Format(layout),
					End:   prevDatePlusOne.Format(layout),
				}
				dateList = append(dateList, dp)

				// make sure prevDate is set again
				prevDate = prevDatePlusOne
				prevDatePlusOne = prevDatePlusOne.Add(gapHours)
			} else {
				// The next date exceeds the end date, so break
				break
			}
		}
	} else {
		// The gap is not a multiple of 24, so return an empty slice (dateList)
		fmt.Println("Gap must be a multiple of 24")
	}

	return dateList
}

// MakeStocksAggUrl A function that makes urls like: /v2/aggs/ticker/{stocksTicker}/range/{multiplier}/{timespan}/{from}/{to}
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
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

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

// MakeAllStocksAggsUrls A quick function that uses MakeStocksAggUrl and iterates through combos and returns a list of urls that will be queried.
func MakeAllStocksAggsUrls(
	tickers []string,
	timespan string,
	from_ string,
	to_ string,
	apiKey string,
	withLinearDates int,
	adjusted int,
	gap int,
) []*url.URL {
	// no need for channels in this yet, just a quick function that makes all the queries and sends it back
	fmt.Println("Making all urls...")

	var urls []*url.URL
	if withLinearDates == 1 {
		datePairs := CreateLinearDatePairs(from_, to_, gap)
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

// MakeTickerVxQuery A function that takes the API string and time, and generates an url.
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
