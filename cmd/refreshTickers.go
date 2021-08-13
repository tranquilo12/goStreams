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
	"lightning/utils/config"
	"lightning/utils/db"
	"lightning/utils/structs"
	"strings"
)

// refreshTickersCmd represents the refreshTickers command
var refreshTickersCmd = &cobra.Command{
	Use:   "refreshTickers",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("refreshTickers called")

		dbType, _ := cmd.Flags().GetString("dbtype")
		if dbType == "" {
			dbType = "ec2db"
		}

		//// get database conn
		DBParams := structs.DBParams{}
		err := config.SetDBParams(&DBParams, dbType)
		if err != nil {
			panic(err)
		}

		postgresDB := db.GetPostgresDBConn(&DBParams)
		defer postgresDB.Close()

		var redisEndpoint string
		if dbType == "ELASTICCACHE" {
			redisEndpoint = config.GetRedisParams("ELASTICCACHE")
		} else {
			redisEndpoint = "localhost"
		}

		redisClient := db.GetRedisClient(6379, redisEndpoint)
		apiKey := config.SetPolygonCred("other")

		fmt.Printf("-- Making all Ticker VX Queries...\n")
		url := db.MakeTickerVxQuery(apiKey)
		Chan1 := db.MakeAllTickersVxRequests(url)

		fmt.Printf("-- Pushing all Ticker VXs to redis...\n")
		err = db.PushTickerVxIntoRedis(Chan1, redisClient)
		if err != nil {
			panic(err)
		}

		allTickersKey := "allTickers"
		allTickers, err := redisClient.Get(allTickersKey).Result()
		if err != nil {
			panic(err)
		}

		fmt.Printf("-- Pushing all Ticker Details to pgdb...\n")
		allUrls := db.MakeAllTickerDetailsQueries(apiKey, strings.Split(allTickers, string(',')))
		err = db.MakeAllTickerDetailsRequestsAndPushToDB(allUrls, postgresDB)

	},
}

func init() {
	rootCmd.AddCommand(refreshTickersCmd)

	refreshTickersCmd.Flags().StringP("dbtype", "d", "ec2db", "One of two... ec2db or localdb")
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// refreshTickersCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// refreshTickersCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
