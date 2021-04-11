/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package main

import (
	"fmt"
	"github.com/akamensky/argparse"
	"lightning/utils/db"
	"lightning/utils/structs"
	"time"

	//url2 "net/url"
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
	prog := parser.String("p", "program", &argparse.Options{Required: true, Help: "TickerTypes(tt) or Aggregates(ag) or TickersVx(vx)?"})
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

	if *prog == "vx" {
		estLoc, _ := time.LoadLocation("America/New_York")
		startDate := time.Date(2021, 1, 1, 0, 0, 0, 0, estLoc)
		endDate := time.Date(2021, 1, 1, 0, 0, 0, 0, estLoc)

		urls := db.MakeAllTickersVxSourceQueries(apiKey, startDate, endDate)
		unexpandedChan := db.MakeAllTickersVxRequests(urls)

		err := db.PushTickerVxIntoDB(unexpandedChan, postgresDB)
		if err != nil {
			panic(err)
		}
	}

	if *prog == "ti" {
		urls := db.MakeAllTickersQuery(apiKey, 1000)
		unexpandedChan := db.MakeAllTickersRequests(urls)
		err := db.PushTickerIntoDB(unexpandedChan, postgresDB)
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

	if *prog == "tb" {
		db.ExecCreateAllTablesModels()
	}

}
