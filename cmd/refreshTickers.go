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
	"github.com/go-pg/pg/v10"
	"github.com/spf13/cobra"
	"lightning/utils/config"
	"lightning/utils/db"
	"lightning/utils/structs"
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

		// get database conn
		DBParams := structs.DBParams{}
		err := config.SetDBParams(&DBParams, "ec2db")
		db.Check(err)

		// Get the DB connection
		postgresDB := db.GetPostgresDBConn(&DBParams)
		defer func(postgresDB *pg.DB) {
			err := postgresDB.Close()
			if err != nil {
				panic(err)
			}
		}(postgresDB)

		// Get the Polygon API key
		apiKey := config.SetPolygonCred("other")

		// Make all the Ticker Vx queries
		fmt.Printf("-- Making all Ticker VX Queries...\n")
		url := db.MakeTickerVxQuery(apiKey)

		// Push all the Ticker Vx queries to a channel
		Chan1 := db.MakeAllTickersVxRequests(url)

		// Get the ticker data from the channel, and push the data to redis
		fmt.Printf("-- Pushing all Ticker VXs to influxDB...\n")
		influxDBClient := db.GetInfluxDBClient(true)
		db.PushTickerVxIntoInfluxDB(Chan1, influxDBClient)
	},
}

func init() {
	rootCmd.AddCommand(refreshTickersCmd)
}
