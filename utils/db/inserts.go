package db

import (
	"encoding/json"
	"fmt"
	"github.com/go-pg/pg/v10"
	"github.com/go-redis/redis/v7"
	"github.com/schollz/progressbar/v3"
	"lightning/utils/structs"
	"strings"
	"sync"
)

//const (
//	AggBarsCols         = "insert_date, ticker, status, queryCount, resultsCount, adjusted, v, vw, o, c, h, l, t, n, request_id, multiplier, timespan"
//	AggBarsPlaceHolders = "$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17"
//	AggBarsIdx          = "insert_date, vw, t, multiplier, timespan"
//)
//
//var AggBarsInsertTemplate = fmt.Sprintf(
//	"INSERT INTO aggregates_bars(%s) VALUES (%s) ON CONFLICT (%s) DO NOTHING;",
//	AggBarsCols,
//	AggBarsPlaceHolders,
//	AggBarsIdx,
//)

//func PushGiantPayloadIntoDB(output []structs.AggregatesBars, poolConfig *pgxpool.Config) {
//	ctx := context.Background()
//	connPool, err := pgxpool.ConnectConfig(ctx, poolConfig)
//	if err != nil {
//		fmt.Println("Unable to create conn...", err)
//	}
//	defer connPool.Close()
//
//	batch := &pgx.Batch{}
//	numInserts := len(output)
//	for k := range output[0 : numInserts-1] {
//		batch.Queue(AggBarsInsertTemplate,
//			output[k].InsertDate,
//			output[k].Ticker,
//			output[k].Status,
//			output[k].Querycount,
//			output[k].Resultscount,
//			output[k].Adjusted,
//			output[k].V,
//			output[k].Vw,
//			output[k].O,
//			output[k].C,
//			output[k].H,
//			output[k].L,
//			output[k].T,
//			output[k].N,
//			output[k].RequestID,
//			output[k].Multiplier,
//			output[k].Timespan,
//		)
//	}
//
//	// pull through the batch and exec each statement
//	br := connPool.SendBatch(context.Background(), batch)
//	for k := 0; k < numInserts-1; k++ {
//		_, err := br.Exec()
//		if err != nil {
//			fmt.Println("Unable to execute statement in batched queue: ", err)
//			os.Exit(1)
//		}
//	}
//
//	// Close this batch pool
//	err = br.Close()
//	if err != nil {
//		fmt.Println("Unable to close batch: ", err)
//	}
//}

func PushTickerTypesIntoDB(insertIntoDB *structs.TickerTypeResponse, db *pg.DB) error {
	flattenedInsertIntoDB := structs.TickerTypesFlattenPayloadBeforeInsert(insertIntoDB)
	_, err := db.Model(&flattenedInsertIntoDB).Insert()
	if err != nil {
		panic(err)
	}
	println("Inserted all TickerTypes into the DB...")
	return nil
}

func PushTickerVxIntoRedis(insertIntoRedis <-chan []structs.TickerVx, rClient *redis.Client) error {
	// use WaitGroup to make things more smooth with channels
	var allTickers []string

	// for each insertIntoDB that follows...spin off another go routine
	for val, ok := <-insertIntoRedis; ok; val, ok = <-insertIntoRedis {
		if ok && val != nil {
			for _, v := range val {
				allTickers = append(allTickers, v.Ticker)
			}
		}
	}
	allTickersStr := strings.Join(allTickers[:], ",")
	err := rClient.Set("allTickers", allTickersStr, 0).Err()
	if err != nil {
		panic(err)
	}

	return nil
}

func GetTickerVxs(insertIntoRedis <-chan []structs.TickerVx) string {
	// use WaitGroup to make things more smooth with channels
	var allTickers []string

	// for each insertIntoDB that follows...spin off another go routine
	for val, ok := <-insertIntoRedis; ok; val, ok = <-insertIntoRedis {
		if ok && val != nil {
			for _, v := range val {
				allTickers = append(allTickers, v.Ticker)
			}
		}
	}
	allTickersStr := strings.Join(allTickers[:], ",")
	return allTickersStr
}

func PushAggIntoRedis(insertIntoRedis <-chan structs.RedisAggBarsResults, rClient *redis.Client) error {
	// for each insertIntoDB that follows...spin off another go routine
	for val, ok := <-insertIntoRedis; ok; val, ok = <-insertIntoRedis {
		if ok {
			resBytes, err := json.Marshal(val.InsertThis)
			if err != nil {
				panic(err)
			}

			err = rClient.Set(val.Key, resBytes, 0).Err()
			if err != nil {
				panic(err)
			}
		}
	}
	return nil
}

func PushTickerNews2IntoDB(insertIntoDB <-chan []structs.TickerNews2, db *pg.DB) error {
	// use WaitGroup to make things more smooth with channels
	var wg sync.WaitGroup

	bar := progressbar.Default(500, "Uploading TickerNews2 to db...")

	// for each insertIntoDB that follows...spin off another go routine
	for val, ok := <-insertIntoDB; ok; val, ok = <-insertIntoDB {
		if ok && val != nil {
			wg.Add(1)

			go func(val []structs.TickerNews2) {
				defer wg.Done()

				_, err := db.Model(&val).
					OnConflict("(id) DO NOTHING").
					Insert()
				if err != nil {
					panic(err)
				}

				var barerr = bar.Add(1)
				if barerr != nil {
					fmt.Println("\nSomething wrong with bar: ", barerr)
				}
			}(val)
		}
	}
	wg.Wait()
	return nil
}
