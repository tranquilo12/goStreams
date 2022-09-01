package cmd

/*
Copyright © 2021 Shriram Sunder <shriram.sunder121091@gmail.com>

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

import (
	"context"
	"flag"
	"fmt"
	"github.com/spf13/cobra"
	"lightning/publisher"
	"lightning/utils/config"
	"lightning/utils/db"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime/pprof"
)

// aggsPubCmd represents the aggs command
var aggsPubCmd = &cobra.Command{
	Use:   "aggsPub",
	Short: "Helps pull data from Polygon-io and into a Kafka topic",
	Long: `
		This command pulls data from Polygon-io and into a Kafka topic.
        Future enhancements will include a command line interface to
		interact with the Kafka topic.
`,
	Run: func(cmd *cobra.Command, args []string) {
		var memprofile = flag.String("memprofile", "", "write memory profile to this file")
		if *memprofile != "" {
			f, err := os.Create(*memprofile)
			if err != nil {
				fmt.Println(err)
			}
			err = pprof.WriteHeapProfile(f)
			if err != nil {
				return
			}
			f.Close()
			return
		}

		go func() {
			log.Println(http.ListenAndServe("localhost:6060", nil))
		}()

		fmt.Println("aggsPub called")
		ctx := context.TODO()

		// Get agg parameters from cli
		aggParams := db.ReadAggregateParamsFromCMD(cmd)

		// Get the apiKey from the config.ini file
		apiKey := config.SetPolygonCred("loving_aryabhata_key")

		// Now get all the tickers from QDB
		tickers := db.QDBFetchUniqueTickersPG(ctx)

		//Make all urls from the tickers
		urls := db.MakeAllStocksAggsUrls(
			tickers,
			aggParams.Timespan,
			aggParams.From,
			aggParams.To,
			apiKey,
			aggParams.Adjusted,
		)

		// Download all data and push the data into kafka
		err := publisher.AggKafkaWriter(urls, "aggs")
		db.CheckErr(err)
	},
}

func init() {
	// Here you will define your flags and configuration settings.
	rootCmd.AddCommand(aggsPubCmd)
	aggsPubCmd.Flags().StringP("dbtype", "d", "ec2db", "One of two... ec2db or localdb")
	aggsPubCmd.Flags().StringP("timespan", "T", "", "Timespan (minute, hour, day...)")
	aggsPubCmd.Flags().StringP("from", "f", "", "From which date? (format = %Y-%m-%d)")
	aggsPubCmd.Flags().StringP("to", "t", "", "To which date? (format = %Y-%m-%d)")
	aggsPubCmd.Flags().IntP("adjusted", "a", 1, "Adjusted? (1/0)")
	aggsPubCmd.Flags().IntP("limit", "L", 50000, "Limit?")
}
