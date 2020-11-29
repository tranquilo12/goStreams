package db

import (
	"context"
	"fmt"
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgxpool"
	"lightning/utils/structs"
	"time"
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

func PushIntoDB(target structs.StocksAggResponseParams, connPool *pgxpool.Pool, timespan string, multiplier int, layout string) {
	batch := &pgx.Batch{}
	numInserts := len(target.Results)

	for i := range target.Results {
		var r structs.AggV2
		r = target.Results[i]
		t := time.Unix(0, int64(r.T))
		batch.Queue(polygonStocksAggCandlesInsertTemplate, target.Ticker, timespan, multiplier, r.V, r.Vw, r.O, r.C, r.H, r.L, t.Format(layout))
	}

	br := connPool.SendBatch(context.Background(), batch)
	for i := 0; i < numInserts; i++ {
		_, err := br.Exec()
		if err != nil {
			fmt.Println("Unable to execute statement in batched queue: ", err)
		}
	}

	err := br.Close()
	if err != nil {
		fmt.Println("Unable to close batch: ", err)
	}
}
