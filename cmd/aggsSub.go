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
	Short: "aggsPub will help 'Publish' 'Aggregates' to any database that you point it to (S3/postgres).",
	Long: `aggsPub: This command will help publish aggregates to any database that you point it towards. 
After inserting data into S3, it will make a file called "currentDataStatus.json", placed at the head of the S3 path, 
that will contain a succinct method of determining which dates have been inserted.

Accepts flags like: 
	1.  forceInsertDate (type: str) (called by: --forceInsertDate) = 0 or 1; If this flag is set to 0, today's date
         will be used as the first index in the flat file structure-like index that's used in S3. 
	2.  useRedis (type: int) (called by: --useRedis) = 0 or 1; If this flag is set to 0, redis will not be used as a 
         middle-man, between polygon-io and the database. It was (will-be) important when queues need to be implemented, 
         i.e if things will go back to Postgres.
	3.  withLinearDates (type: int) (called by: --withLinearDates) = 0 or 1; If this flag is set to 0, the urls that are called will not be
        split into (start_day, start_day + 1... end_date) but rather (start_date, end_date).
	4.  dbtype (type: str) (called by: --dbtype) = "ec2db"; Either keep this param empty or make it ec2db; (Will be removed in later versions).
	5.  timespan (type: str) (called by: --timespan) = minute/hour/day/week/month/quarter/year, found in (https://polygon.io/docs/get_v2_aggs_ticker__stocksTicker__range__multiplier___timespan___from___to__anchor).
	6.  multiplier (type: int) (called by: --mult) = usually 1.
	7.  from (type: str) (called by: --from) = the start date in %Y-%m-%d format.
	8.  to (type: str) (called by: --to) = the end date in %Y-%m-%d format.
	9.  adjusted (type: int) (called by: --adj) = 0 or 1; If this flag is set 0, only adjusted data will be called. 
	10. limit (type: int) (called by: --limit) = 300; The rate limit by which data will be pulled (from polygon-io) and inserted into S3.
`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("aggsPub called")
		dbType, _ := cmd.Flags().GetString("dbtype")
		if dbType == "" {
			dbType = "ec2db"
		}

		// Get agg parameters from console
		aggParams := db.ReadAggregateParamsFromCMD(cmd)

		// Possibly get all the redis parameters from the .ini file.
		var redisParams config.RedisParams
		err := config.SetRedisCred(&redisParams)
		if err != nil {
			panic(err)
		}

		var redisEndpoint string
		if dbType == "ELASTICCACHE" {
			redisEndpoint = config.GetElasticCacheEndpoint("ELASTICCACHE")
		} else {
			redisEndpoint = "localhost"
		}

		var polygonTickers []string
		if aggParams.UseRedis == 1 {
			fmt.Printf("Getting redis client and tickers from redis...")
			pool := db.GetRedisPool(6379, redisEndpoint)
			// conn will close within GetAllTickersFromRedis
			polygonTickers = db.GetAllTickersFromRedis(pool)
		} else {
			polygonTickers = db.GetAllTickersFromPolygonioDirectly()
		}
		forceInsertDate := aggParams.ForceInsertDate

		var insertDate string
		if forceInsertDate == "" {
			insertDate = time.Now().Format("2006-01-02")
		} else {
			insertDate = aggParams.ForceInsertDate
		}

		fmt.Printf("Getting All agg tickers from s3...\n")
		s3Tickers := subscriber.GetAggTickersFromS3(
			insertDate,
			aggParams.Timespan,
			aggParams.Multiplier,
			aggParams.From,
			aggParams.To,
			aggParams.Adjusted,
		)

		fmt.Printf("Getting the difference between tickers in redis and s3...\n")
		tickers := db.GetDifferenceBtwTickersInMemAndS3(polygonTickers, *s3Tickers)

		fmt.Printf("Making all stocks aggs queries...\n")
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

		fmt.Printf("Publishing all values to the db...\n")
		err = subscriber.AggPublisher(urls, aggParams.Limit, forceInsertDate, aggParams.Adjusted)
		if err != nil {
			fmt.Println("Something wrong with AggPublisher...")
			panic(err)
		}

		if aggParams.Adjusted == 1 {
			fmt.Printf("Pushing current status of all data in s3 to s3 as currentDataStatusAdj.json...\n")
			var s3tickers []string
			m1 := make(map[string][][]string)
			m2 := make(map[string]map[string][][]string)
			fmt.Printf("Fetching all objects from %s with prefix %s ...\n", "polygonio-all", "aggs/adj")
			allObjs := publisher.ListAllBucketObjsS3("polygonio-all", "aggs/adj")

			fmt.Printf("Parsing all objects from %s with prefix %s ...\n", "polygonio-all", "aggs/adj")
			for _, ele := range *allObjs {
				splitEle := strings.Split(ele, "/")
				insertDate := strings.Join(splitEle[2:5], "-")
				startDate := strings.Join(splitEle[7:10], "-")
				endDate := strings.Join(splitEle[10:13], "-")
				timespan := splitEle[5]
				s3tickers = append(s3tickers, splitEle[13])
				m1[insertDate] = append(m1[insertDate], []string{startDate, endDate})
				m1[insertDate] = subscriber.Unique2dStr(m1[insertDate])
				m2[timespan] = m1
			}

			sendToS3, err := json.Marshal(m2)
			if err != nil {
				panic(err)
			}

			fmt.Printf("Pushing all objects from %s with prefix %s to s3 as %s...\n", "polygonio-all", "aggs/adj", "currentDataStatusAdj.json")
			err = subscriber.UploadToS3("polygonio-all", "currentDataStatusAdj.json", sendToS3)
		} else {
			fmt.Printf("Pushing current status of all data in s3 to s3 as currentDataStatus.json...\n")
			var s3tickers []string
			m1 := make(map[string][][]string)
			m2 := make(map[string]map[string][][]string)
			fmt.Printf("Fetching all objects from %s with prefix %s ...\n", "polygonio-all", "aggs")
			allObjs := publisher.ListAllBucketObjsS3("polygonio-all", "aggs")

			fmt.Printf("Parsing all objects from %s with prefix %s ...\n", "polygonio-all", "aggs")
			for _, ele := range *allObjs {
				splitEle := strings.Split(ele, "/")
				insertDate := strings.Join(splitEle[1:4], "-")
				startDate := strings.Join(splitEle[6:9], "-")
				endDate := strings.Join(splitEle[9:12], "-")
				timespan := splitEle[4]
				s3tickers = append(s3tickers, splitEle[12])
				m1[insertDate] = append(m1[insertDate], []string{startDate, endDate})
				m1[insertDate] = subscriber.Unique2dStr(m1[insertDate])
				m2[timespan] = m1
			}

			sendToS3, err := json.Marshal(m2)
			if err != nil {
				panic(err)
			}

			fmt.Printf("Pushing all objects from %s with prefix %s to s3 as %s...\n", "polygonio-all", "aggs", "currentDataStatus.json")
			err = subscriber.UploadToS3("polygonio-all", "currentDataStatus.json", sendToS3)

		}
	},
}

func init() {
	// Here you will define your flags and configuration settings.
	rootCmd.AddCommand(aggsPubCmd)
	aggsPubCmd.Flags().StringP("dbtype", "d", "ec2db", "One of two... ec2db or localdb")
	aggsPubCmd.Flags().StringP("timespan", "T", "", "Timespan (minute, hour, day...)")
	aggsPubCmd.Flags().StringP("from", "f", "", "From which date? (format = %Y-%m-%d)")
	aggsPubCmd.Flags().StringP("to", "t", "", "To which date? (format = %Y-%m-%d)")
	aggsPubCmd.Flags().IntP("mult", "m", 2, "Multiplier to use with Timespan")
	aggsPubCmd.Flags().IntP("limit", "l", 300, "Rate limit to pull from polygonio")
	aggsPubCmd.Flags().IntP("withLinearDates", "w", 1, "Usually 1, if appending datasets day-to-day, but if for backup, use 0.")
	aggsPubCmd.Flags().StringP("forceInsertDate", "F", "", "Force an insert date, to overwrite past data?")
	aggsPubCmd.Flags().IntP("useRedis", "u", 0, "Should you use redis?")
	aggsPubCmd.Flags().IntP("adjusted", "a", 1, "Adjusted or unadjusted?")
}
