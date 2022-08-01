package db

import (
	"fmt"
	"github.com/go-pg/pg/v10"
	"github.com/schollz/progressbar/v3"
	"lightning/utils/structs"
	"strings"
	"sync"
)

func CheckErr(err error) {
	if err != nil {
		panic(err)
	}
}

func PushTickerTypesIntoDB(insertIntoDB *structs.TickerTypeResponse, db *pg.DB) error {
	flattenedInsertIntoDB := structs.TickerTypesFlattenPayloadBeforeInsert(insertIntoDB)
	_, err := db.Model(&flattenedInsertIntoDB).Insert()
	if err != nil {
		panic(err)
	}
	println("Inserted all TickerTypes into the DB...")
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
