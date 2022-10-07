package db

import (
	"encoding/json"
	"github.com/schollz/progressbar/v3"
	"lightning/utils/structs"
	"net/http"
	"net/url"
)

func AddApiKeyToUrl(u string, apiKey string) *url.URL {
	parsedUrl, err := url.Parse(u)
	CheckErr(err)
	q := parsedUrl.Query()
	q.Add("apiKey", apiKey)
	parsedUrl.RawQuery = q.Encode()
	return parsedUrl
}

func GetAndCheckResponse(u *url.URL) *http.Response {
	// Make the query
	response, err := http.Get(u.String())
	CheckErr(err)
	return response
}

// FetchAllTickers recursively fetches all tickers and pushes it to QuestDB
func FetchAllTickers(apiKey string, TickerChan chan structs.TickersStruct) {
	// Get a progress bar2
	bar2 := progressbar.Default(30, "Fetching tickers...")

	// Get the ticker URL
	parsedURL := MakeTickerURL(apiKey)

	// Get and clean response
	response := GetAndCheckResponse(parsedURL)

	// Push response to chan and make another request using the "next_url" and apiKey, until next_url is not available.
	for {
		if response.StatusCode == 200 {
			// Decode to the structs.TickersStruct var
			ticker := structs.TickersStruct{}
			err := json.NewDecoder(response.Body).Decode(&ticker)
			CheckErr(err)

			// Push to channel
			TickerChan <- ticker

			// Check if everything is good, and if we have a next url
			if ticker.NextURL != "" {
				nextURL := AddApiKeyToUrl(ticker.NextURL, apiKey)
				response = GetAndCheckResponse(nextURL)
			} else {
				return
			}
		} else {
			return
		}

		// Update the progress bar2
		err := bar2.Add(1)
		CheckErr(err)
	}
}
