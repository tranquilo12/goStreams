package db

import (
	"fmt"
	"lightning/utils/structs"
	"strconv"
	"time"
)

const (
	polygonStocksAggCandlesCols         = "ticker, timespan, multiplier, volume, vwap, open, close, high, low, timestamp"
	polygonStocksAggCandlesPlaceHolders = "$1, $2, $3, $4, $5, $6, $7, $8, $9, $10"
	polygonStocksAggCandlesIdx          = "ticker, timespan, multiplier, vwap, timestamp"
)

var PolygonStocksAggCandlesInsertTemplate = fmt.Sprintf(
	"INSERT INTO polygon_stocks_agg_candles(%s) VALUES (%s) ON CONFLICT (%s) DO NOTHING;",
	polygonStocksAggCandlesCols,
	polygonStocksAggCandlesPlaceHolders,
	polygonStocksAggCandlesIdx,
)

func msToTime(ms string) (time.Time, error) {
	msInt, err := strconv.ParseInt(ms, 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(0, msInt*int64(time.Millisecond)), nil
}

func FlattenPayloadBeforeInsert(target structs.StocksAggResponseParams, timespan string, multiplier int, layout string) []structs.ExpandedStocksAggResponseParams {
	var output []structs.ExpandedStocksAggResponseParams
	for i := range target.Results {
		var r structs.AggV2
		r = target.Results[i]
		t, err := msToTime(strconv.FormatInt(int64(r.T), 10))
		if err != nil {
			fmt.Println("Some Error parsing date: ", err)
		}
		newArr := structs.ExpandedStocksAggResponseParams{
			Ticker:     target.Ticker,
			Timespan:   timespan,
			Multiplier: multiplier,
			V:          r.V,
			Vw:         r.Vw,
			O:          r.O,
			C:          r.C,
			H:          r.H,
			L:          r.L,
			T:          t.Format(layout),
		}
		output = append(output, newArr)
	}
	return output
}
