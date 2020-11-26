package db

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/pgxpool"
)

const (
	polygonStocksAggCandlesCols         = "ticker, timespan, multiplier, volume, vwap, open, close, high, low, timestamp"
	polygonStocksAggCandlesPlaceHolders = "$1, $2, $3, $4, $5, $6, $7, $8, $9, $10"
	polygonStocksAggCandlesIdx          = "ticker, timespan, multiplier, vwap, timestamp"
)

var polygonStocksAggCandlesInsertTemplate = fmt.Sprintf(
	"INSERT INTO polygon_stocks_agg_candles(%s) VALUES (%s) ON CONFLICT (%s) DO NOTHING",
	polygonStocksAggCandlesCols,
	polygonStocksAggCandlesPlaceHolders,
	polygonStocksAggCandlesIdx,
)

func PushIntoDB(target interface{}, connPool *pgxpool.Pool) {
	if _, err := connPool.Exec(context.Background(), polygonStocksAggCandlesInsertTemplate, target); err != nil {
		fmt.Println("Insert Error: ", err)
	}
}
