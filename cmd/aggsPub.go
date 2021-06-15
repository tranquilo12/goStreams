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
	"lightning/publisher"
	"lightning/utils/config"
	"lightning/utils/db"
	"lightning/utils/structs"
)

// aggsPubCmd represents the aggs command
var aggsPubCmd = &cobra.Command{
	Use:   "aggsPub",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("aggsPub called")
		dbType, _ := cmd.Flags().GetString("dbtype")
		if dbType == "" {
			dbType = "ec2db"
		}

		// get database conn
		DBParams := structs.DBParams{}
		err := config.SetDBParams(&DBParams, dbType)
		if err != nil {
			panic(err)
		}

		postgresDB := db.GetPostgresDBConn(&DBParams)
		defer postgresDB.Close()

		// Get agg parameters from cli
		aggParams := db.ReadAggregateParamsFromCMD(cmd)

		// Possibly get all the redis parameters from the .ini file.
		var redisParams config.RedisParams
		err = config.SetRedisCred(&redisParams)
		if err != nil {
			panic(err)
		}

		redisEndpoint := config.GetRedisParams("ELASTICCACHE")
		redisClient := db.GetRedisClient(7000, redisEndpoint)

		//var tickers = db.GetAllTickers(postgresDB, aggParams.Timespan)
		var tickers = db.GetAllTickersFromRedis(redisClient)
		urls := db.MakeAllStocksAggsQueries(tickers, aggParams.Timespan, aggParams.From, aggParams.To, apiKey)
		err = publisher.AggPublisher(urls)
		if err != nil {
			fmt.Println("Something wrong with AggPublisher...")
			panic(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(aggsPubCmd)
	// Here you will define your flags and configuration settings.
	aggsPubCmd.Flags().StringP("dbtype", "d", "ec2db", "One of two... ec2db or localdb")
	aggsPubCmd.Flags().StringP("timespan", "T", "", "Timespan (minute, hour, day...)")
	aggsPubCmd.Flags().StringP("from", "f", "", "From which date? (format = %Y-%m-%d)")
	aggsPubCmd.Flags().StringP("to", "t", "", "To which date? (format = %Y-%m-%d)")
	aggsPubCmd.Flags().IntP("mult", "m", 2, "Multiplier to use with Timespan")
}
