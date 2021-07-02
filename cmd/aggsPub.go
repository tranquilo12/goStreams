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
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"lightning/publisher"
	"lightning/subscriber"
	"lightning/utils/config"
	"lightning/utils/db"
	"strings"
	"time"
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

		// Get agg parameters from cli
		aggParams := db.ReadAggregateParamsFromCMD(cmd)

		// Possibly get all the redis parameters from the .ini file.
		var redisParams config.RedisParams
		err := config.SetRedisCred(&redisParams)
		if err != nil {
			panic(err)
		}

		var redisEndpoint string
		if dbType == "ELASTICCACHE" {
			redisEndpoint = config.GetRedisParams("ELASTICCACHE")
		} else {
			redisEndpoint = "localhost"
		}

		fmt.Printf("Getting redis client and tickers from redis...")
		redisClient := db.GetRedisClient(6379, redisEndpoint)
		redisTickers := db.GetAllTickersFromRedis(redisClient)
		today := time.Now().Format("2006-01-02")

		fmt.Printf("Getting All agg tickers from s3...\n")
		s3Tickers := publisher.GetAggTickersFromS3(today, aggParams.Timespan, aggParams.Multiplier, aggParams.From, aggParams.To)

		fmt.Printf("Getting the difference between tickers in redis and s3...\n")
		tickers := db.GetDifferenceBtwTickersInRedisAndS3(*redisTickers, *s3Tickers)

		fmt.Printf("Making all stocks aggs queries...\n")
		urls := db.MakeAllStocksAggsQueries(tickers, aggParams.Timespan, aggParams.From, aggParams.To, apiKey, aggParams.WithLinearDates)

		fmt.Printf("Publishing all values to the db...\n")
		err = publisher.AggPublisher(urls, aggParams.Limit)
		if err != nil {
			fmt.Println("Something wrong with AggPublisher...")
			panic(err)
		}

		fmt.Printf("Pushing current status of all data in s3 to s3 as currentDataStatus.json...\n")
		var s3tickers []string
		m1 := make(map[string][][]string)
		m2 := make(map[string]map[string][][]string)

		fmt.Printf("Fetching all objects from %s with prefix %s ...\n", "polygonio-all", "aggs")
		allObjs := subscriber.ListAllBucketObjsS3("polygonio-all", "aggs")

		fmt.Printf("Parsing all objects from %s with prefix %s ...\n", "polygonio-all", "aggs")
		for _, ele := range *allObjs {
			splitEle := strings.Split(ele, "/")
			insertDate := strings.Join(splitEle[1:4], "-")
			startDate := strings.Join(splitEle[6:9], "-")
			endDate := strings.Join(splitEle[9:12], "-")
			timespan := splitEle[4]
			s3tickers = append(s3tickers, splitEle[12])
			m1[insertDate] = append(m1[insertDate], []string{startDate, endDate})
			m1[insertDate] = publisher.Unique2dStr(m1[insertDate])
			m2[timespan] = m1
		}

		sendToS3, err := json.Marshal(m2)
		if err != nil {
			panic(err)
		}

		fmt.Printf("Pushing all objects from %s with prefix %s to s3 as %s...\n", "polygonio-all", "aggs", "currentDataStatus.json")
		err = publisher.UploadToS3("polygonio-all", "currentDataStatus.json", sendToS3)
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
	aggsPubCmd.Flags().IntP("limit", "l", 300, "Rate limit to pull from polygonio")
	aggsPubCmd.Flags().IntP("withLinearDates", "w", 1, "Usually 1, if appending datasets day-to-day, but if for backup, use 0.")
}
