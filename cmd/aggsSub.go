package cmd

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

import (
	"fmt"
	"github.com/spf13/cobra"
	"lightning/subscriber"
)

// aggsPubCmd represents the aggs command
var aggsSubCmd = &cobra.Command{
	Use:   "aggsSub",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("aggsSub called")

		// get database conn
		//DBParams := db.ReadPostgresDBParamsFromCMD(cmd)
		//postgresDB := db.GetPostgresDBConn(&DBParams)
		//defer postgresDB.Close()
		//
		//// Get agg parameters from cli
		//aggParams := db.GetAggParams(cmd)

		// Possibly get all the redis parameters from the .ini file.
		//var redisParams config.RedisParams
		//err := config.SetRedisCred(&redisParams)
		//if err != nil {
		//	panic(err)
		//}

		// Get a pool of redis connections
		//var redisPool *redis.Pool
		//redisPool = db.GetRedisPool()

		// Get a New Re-Json Handler who's client will be set later within AggPublisher.
		//rh := rejson.NewReJSONHandler()

		//var tickers = []string{"AAPL", "GME"}
		//urls := db.MakeAllStocksAggsQueries(tickers, aggParams.Timespan, aggParams.From, aggParams.To, apiKey)
		//err = publisher.AggPublisher(redisPool, rh, urls, true)
		err := subscriber.AggSubscriber()
		//err = publisher.AggPublisher(urls)
		if err != nil {
			fmt.Println("Something wrong with AggPublisher...")
			panic(err)
		}

		//unexpandedChan := db.MakeAllAggRequests(urls, timespan, multiplier)

		// insert all the data quickly!
		//err := db.PushGiantPayloadIntoDB1(unexpandedChan, postgresDB)
		//if err != nil {
		//	panic(err)
		//}
	},
}

func init() {
	rootCmd.AddCommand(aggsSubCmd)

	// Here you will define your flags and configuration settings.
	aggsSubCmd.Flags().StringP("user", "u", "", "Postgres username")
	aggsSubCmd.Flags().StringP("password", "P", "", "Postgres password")
	aggsSubCmd.Flags().StringP("database", "d", "", "Postgres database name")
	aggsSubCmd.Flags().StringP("host", "H", "127.0.0.1", "Postgres host (default localhost)")
	aggsSubCmd.Flags().StringP("port", "p", "5432", "Postgres port (default 5432)")
	aggsSubCmd.Flags().StringP("timespan", "T", "", "Timespan (minute, hour, day...)")
	aggsSubCmd.Flags().StringP("from", "f", "", "From which date? (format = %Y-%m-%d)")
	aggsSubCmd.Flags().StringP("to", "t", "", "To which date? (format = %Y-%m-%d)")
	aggsSubCmd.Flags().IntP("mult", "m", 2, "Multiplier to use with Timespan")

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// aggsPubCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// aggsPubCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
