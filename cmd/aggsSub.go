// Package cmd /*
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
	"fmt"
	_ "github.com/segmentio/kafka-go/snappy"
	"github.com/spf13/cobra"
	"lightning/subscriber"
	"lightning/utils/config"
	"lightning/utils/db"
)

// aggsSub2Cmd represents the aggsPub2 command
var aggsSubCmd = &cobra.Command{
	Use:   "aggsSub",
	Short: "Helps pull data from the Kafka topic to the QuestDB database",
	Long: `
		This command helps pull data from the Kafka topic to the QuestDB database.
		Future versions will include a command line interface to the Kafka topic.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("aggsSub called")
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

		fmt.Println("-	Starting to read from Kafka topic and pushing to QuestDB...")
		subscriber.WriteFromKafkaToQuestDB("aggs", urls)
	},
}

func init() {
	rootCmd.AddCommand(aggsSubCmd)
	aggsSubCmd.Flags().StringP("dbtype", "d", "ec2db", "One of two... ec2db or localdb")
	aggsSubCmd.Flags().StringP("timespan", "T", "", "Timespan (minute, hour, day...)")
	aggsSubCmd.Flags().StringP("from", "f", "", "From which date? (format = %Y-%m-%d)")
	aggsSubCmd.Flags().StringP("to", "t", "", "To which date? (format = %Y-%m-%d)")
	aggsSubCmd.Flags().IntP("adjusted", "a", 1, "Adjusted? (1/0)")
	aggsSubCmd.Flags().IntP("limit", "L", 50000, "Limit?")
}
