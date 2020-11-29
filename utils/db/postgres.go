package db

import (
	"context"
	"fmt"
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgxpool"
	"github.com/vbauerster/mpb"
	"github.com/vbauerster/mpb/decor"
	"lightning/utils/structs"
	"math/rand"
	"strconv"
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

func FlattenPayloadBeforeInsert(target structs.StocksAggResponseParams, timespan string, multiplier int, layout string) []structs.ExpandedStocksAggResponseParams {
	var output []structs.ExpandedStocksAggResponseParams
	for i := range target.Results {
		var r structs.AggV2
		r = target.Results[i]
		ft, err := strconv.ParseInt(strconv.FormatInt(int64(r.T), 10), 10, 64)
		if err != nil {
			panic(err)
		}
		t := time.Unix(ft, 0)

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

func PushGiantPayloadIntoDB(output []structs.ExpandedStocksAggResponseParams, connPool *pgxpool.Pool, p *mpb.Progress) {
	tx, err := connPool.Begin(context.Background())
	if err != nil {
		panic(err)
	}

	batch := &pgx.Batch{}

	// For each of these inserts
	for i := range output {
		batch.Queue(polygonStocksAggCandlesInsertTemplate,
			output[i].Ticker,
			output[i].Timespan,
			output[i].Multiplier,
			output[i].V,
			output[i].Vw,
			output[i].O,
			output[i].C,
			output[i].H,
			output[i].L,
			output[i].T)
	}

	// pull through the batch and exec each statement
	br := connPool.SendBatch(context.Background(), batch)
	numInserts := len(output)

	for i := 0; i < numInserts; i++ {
		// ANSI escape sequences are not supported on Windows OS
		task := fmt.Sprintf("Postgres Op#%02d:", i)
		job := fmt.Sprintf("Upserting...")

		// preparing delayed bars
		b := p.AddBar(rand.Int63n(101)+100,
			//mpb.BarQueueAfter(bars[j]),
			mpb.BarFillerClearOnComplete(),
			mpb.PrependDecorators(
				decor.Name(task, decor.WC{W: len(task) + 1, C: decor.DidentRight}),
				decor.OnComplete(decor.Name(job, decor.WCSyncSpaceR), "done!"),
				decor.OnComplete(decor.EwmaETA(decor.ET_STYLE_MMSS, 0, decor.WCSyncWidth), ""),
			),
			mpb.AppendDecorators(
				decor.OnComplete(decor.Percentage(decor.WC{W: 5}), ""),
			),
		)

		for !b.Completed() {
			start := time.Now()
			_, err := br.Exec()
			if err != nil {
				fmt.Println("Unable to execute statement in batched queue: ", err)
			}
			b.IncrBy(i + 1)
			b.DecoratorEwmaUpdate(time.Since(start))
		}

	}

	// commit everything
	err = tx.Commit(context.Background())
	if err != nil {
		fmt.Println("Unable to commit batch: ", err)
	}

	// Close this batch pool
	err = br.Close()
	if err != nil {
		fmt.Println("Unable to close batch: ", err)
	}
}
