package structs

// structs for response from the above http query
type (
	AggV2 struct {
		V  float64 `json:"v"`
		Vw float64 `json:"vw"`
		O  float64 `json:"o"`
		C  float64 `json:"c"`
		H  float64 `json:"h"`
		L  float64 `json:"l"`
		T  float64 `json:"t"`
		N  float64 `json:"n"`
	}
	StocksAggResponseParams struct {
		Ticker       string  `json:"ticker"`
		Status       string  `json:"status"`
		QueryCount   int     `json:"queryCount"`
		ResultsCount int     `json:"resultsCount"`
		Adjusted     bool    `json:"adjusted"`
		Results      []AggV2 `json:"results"`
		RequestId    string  `json:"request_id"`
	}
	StocksAggRequestsURL struct {
		Url string `json:"url"`
	}
)

// for reading the json file 'responses.json', we need to have a super structure of all the other queries
type (
	StocksAggPart struct {
		Request  StocksAggRequestsURL    `json:"request"`
		Response StocksAggResponseParams `json:"response"`
	}
	ResponsesJSONFile struct {
		StocksAgg StocksAggPart `json:"aggregates"`
	}
)
