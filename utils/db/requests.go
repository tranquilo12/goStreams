package db

import (
	"encoding/json"
	"fmt"
	"github.com/schollz/progressbar/v3"
	"io"
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

// MakeTickerTypesRequest makes a request to the ticker types endpoint
func MakeTickerTypesRequest(apiKey string) *structs.TickerTypeResponse {
	TickerTypesUrl := MakeTickerTypesUrl(apiKey)
	TickerTypesTarget := new(structs.TickerTypeResponse)

	resp, err := http.Get(TickerTypesUrl.String())
	if err != nil {
		panic(err)
	}

	err = json.NewDecoder(resp.Body).Decode(&TickerTypesTarget)
	if err != nil {
		panic(err)
	}

	return TickerTypesTarget
}

// MakeAllTickersVxRequests makes a request to the API for all tickerVx data
func MakeAllTickersVxRequests(u *url.URL) chan []structs.TickerVx {
	var vxResponse *structs.TickersVxResponse
	var response *http.Response
	var p *url.URL
	var err error
	apiKey := u.Query()["apiKey"]
	var newCursor string

	// create a channel to make sure all requests are not being thrown away, of the flattened type.
	c := make(chan []structs.TickerVx, 100000)

	response, err = http.Get(u.String())
	if err != nil {
		panic(err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(response.Body)

	j := 0
	i := 0
	for {
		if response.StatusCode == 200 {
			err = json.NewDecoder(response.Body).Decode(&vxResponse)
			if err != nil {
				panic(err)
			}

			flattenVxResponse := structs.TickersVxFlattenPayloadBeforeInsert(*vxResponse)
			c <- flattenVxResponse

			if vxResponse.NextUrl != "" {

				p, err = url.Parse(vxResponse.NextUrl)
				if err != nil {
					panic(err)
				}
				oldCursor := p.Query()["cursor"][0]

				q := p.Query()
				q.Add("apiKey", apiKey[0])
				p.Host = "api.polygon.io:443"
				p.RawQuery = q.Encode()

				//fmt.Println(p.String())
				response, err = http.Get(p.String())
				if err != nil {
					panic(err)
				}

				if i > 0 {
					p, err = url.Parse(vxResponse.NextUrl)
					if err != nil {
						panic(err)
					}
					newCursor = p.Query()["cursor"][0]
					if newCursor == oldCursor {
						j += 1
						if j > 20 {
							break
						}
					}
				}
				i += 1
			}
		} else {
			break
		}
	}
	close(c)
	return c
}

// MakeAllTickerNews2Requests makes all the requests to polygon.io to get all the news for a ticker
func MakeAllTickerNews2Requests(u *url.URL) chan []structs.TickerNews2 {
	var News2Response *structs.TickerNews2Response
	var response *http.Response
	var p *url.URL
	var err error
	apiKey := u.Query()["apiKey"]
	var newCursor string

	// create a channel to make sure all requests are not being thrown away, of the flattened type.
	c := make(chan []structs.TickerNews2, 100000)

	response, err = http.Get(u.String())
	if err != nil {
		panic(err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			panic(err)
		}
	}(response.Body)

	j := 0
	i := 0
	for {
		if response.StatusCode == 200 {
			err = json.NewDecoder(response.Body).Decode(&News2Response)
			if err != nil {
				panic(err)
			}

			flattenVxResponse := structs.TickerNews2FlattenPayloadBeforeInsert(*News2Response)
			c <- flattenVxResponse

			if News2Response.NextURL != "" {

				p, err = url.Parse(News2Response.NextURL)
				if err != nil {
					panic(err)
				}
				oldCursor := p.Query()["cursor"][0]

				q := p.Query()
				q.Add("apiKey", apiKey[0])
				p.Host = "api.polygon.io:443"
				p.RawQuery = q.Encode()

				fmt.Println(p.String())
				response, err = http.Get(p.String())
				if err != nil {
					panic(err)
				}

				if i > 0 {
					p, err = url.Parse(News2Response.NextURL)
					if err != nil {
						panic(err)
					}
					newCursor = p.Query()["cursor"][0]
					if newCursor == oldCursor {
						j += 1
						if j > 20 {
							break
						}
					}
				}
				i += 1
			} else {
				break
			}
		} else {
			break
		}
	}
	close(c)
	return c
}

// FetchAllTickers recursively fetches all tickers and pushes it to QuestDB
func FetchAllTickers(apiKey string, Tickerchan chan structs.TickersStruct) {
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
			Tickerchan <- ticker

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
