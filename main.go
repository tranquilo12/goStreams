package main

import (
	"fmt"
	"github.com/akamensky/argparse"
	"lightning/utils/db"
	"lightning/utils/structs"
	"os"
)

const (
	// make this more secure
	apiKey     = "9AheK9pypnYOf_DU6TGpydCK6IMEVkIw"
	timespan   = "minute"
	from_      = "2021-01-01"
	to_        = "2021-03-20"
	multiplier = 1
)

func main() {
	parser := argparse.NewParser("lightning", "Which program do you want to execute?")
	prog := parser.String("p", "program", &argparse.Options{Required: true, Help: "TickerTypes(tt) or Aggregates(ag)?"})
	err := parser.Parse(os.Args)
	if err != nil {
		// In case of error print error and print usage
		// This can also be done by passing -h or --help flags
		fmt.Print(parser.Usage(err))
	}

	// get database conn
	postgresDB := db.GetPostgresDB()
	defer postgresDB.Close()

	if *prog == "tt" {
		var insertIntoDB *structs.TickerTypeResponse
		insertIntoDB = db.MakeTickerTypesRequest(apiKey)

		err := db.PushTickerTypesIntoDB(insertIntoDB, postgresDB)
		if err != nil {
			panic(err)
		}
	}

	if *prog == "ag" {
		// get all ticker
		var tickers = []string{"AAPL", "GME"}

		urls := db.MakeAllStocksAggsQueries(tickers, timespan, from_, to_, apiKey)
		unexpandedChan := db.MakeAllAggRequests(urls, timespan, multiplier)

		// insert all the data quickly!
		err := db.PushGiantPayloadIntoDB1(unexpandedChan, postgresDB)
		if err != nil {
			panic(err)
		}
	}
}
