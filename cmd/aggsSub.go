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
	"fmt"
	"github.com/spf13/cobra"
	"lightning/subscriber"
	"lightning/utils/config"
	"lightning/utils/db"
	"strconv"
)

// aggsSubCmd represents the aggs command
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
		dbType, _ := cmd.Flags().GetString("dbtype")
		if dbType == "" {
			dbType = "ec2db"
		}

		// Get agg parameters from cli
		aggParams := db.ReadAggregateParamsFromCMD(cmd)

		// Get the redis params from the config.ini file
		redisParams := config.RedisParams{}
		err := config.SetRedisCred(&redisParams)

		// Create a redis client
		port, err := strconv.Atoi(redisParams.Port)
		Check(err)

		// Get the apiKey from the config.ini file
		apiKey := config.SetPolygonCred("me")

		// Get all the tickers from the redis db
		pool := db.GetRedisPool(port, redisParams.Host)
		defer pool.Close()

		// Get all the tickers from the redis db
		tickers := db.GetAllTickersFromRedis(pool)

		// Make all urls from the tickers
		urls := db.MakeAllStocksAggsUrls(
			tickers,
			aggParams.Timespan,
			aggParams.From,
			aggParams.To,
			apiKey,
			aggParams.WithLinearDates,
			aggParams.Adjusted,
			aggParams.Gap,
		)

		// Download all data and push the data into redis
		err = subscriber.AggDownloader(urls, aggParams.ForceInsertDate, aggParams.Adjusted, pool)
		Check(err)
	},
}

func init() {
	// Here you will define your flags and configuration settings.
	rootCmd.AddCommand(aggsSubCmd)
	aggsSubCmd.Flags().StringP("dbtype", "d", "ec2db", "One of two... ec2db or localdb")
	aggsSubCmd.Flags().StringP("timespan", "T", "", "Timespan (minute, hour, day...)")
	aggsSubCmd.Flags().StringP("from", "f", "", "From which date? (format = %Y-%m-%d)")
	aggsSubCmd.Flags().StringP("to", "t", "", "To which date? (format = %Y-%m-%d)")
	aggsSubCmd.Flags().IntP("mult", "m", 2, "Multiplier to use with Timespan")
	aggsSubCmd.Flags().IntP("withLinearDates", "l", 1, "With linear dates? (1/0)")
	aggsSubCmd.Flags().IntP("adjusted", "a", 1, "Adjusted? (1/0)")
	aggsSubCmd.Flags().IntP("gap", "g", 24, "Gap?")
	aggsSubCmd.Flags().StringP("forceInsertDate", "i", "", "Force insert date?")
	aggsSubCmd.Flags().IntP("useRedis", "r", 1, "Use redis?(1/0)")
	aggsSubCmd.Flags().IntP("limit", "L", 50000, "Limit?")
}
