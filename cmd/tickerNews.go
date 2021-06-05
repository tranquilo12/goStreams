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
	"lightning/utils/config"
	"lightning/utils/db"
)

// tickerNewsCmd represents the refreshTickers command
var tickerNewsCmd = &cobra.Command{
	Use:   "refreshTickerNews2",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("refreshTickers called")

		// get database conn
		DBParams := db.ReadPostgresDBParamsFromCMD(cmd)
		postgresDB := db.GetPostgresDBConn(&DBParams)
		defer postgresDB.Close()

		// read the ticker, from and to dates
		newsParams := db.ReadTickerNewsParamsFromCMD(cmd)

		apiKey := config.SetPolygonCred("other")
		url := db.MakeTickerNews2Query(apiKey, newsParams.Ticker, newsParams.From)
		Chan1 := db.MakeAllTickerNews2Requests(url)
		err := db.PushTickerNews2IntoDB(Chan1, postgresDB)
		if err != nil {
			panic(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(tickerNewsCmd)

	tickerNewsCmd.Flags().StringP("user", "u", "", "Postgres username")
	tickerNewsCmd.Flags().StringP("password", "P", "", "Postgres password")
	tickerNewsCmd.Flags().StringP("database", "d", "", "Postgres database name")
	tickerNewsCmd.Flags().StringP("host", "H", "127.0.0.1", "Postgres host (default localhost)")
	tickerNewsCmd.Flags().StringP("port", "p", "5432", "Postgres port (default 5432)")
	tickerNewsCmd.Flags().StringP("ticker", "T", "", "Which ticker? (eg. AAPL)")
	tickerNewsCmd.Flags().StringP("from", "f", "", "From which date? (format = %Y-%m-%d)")
	tickerNewsCmd.Flags().StringP("to", "t", "", "To which date? (format = %Y-%m-%d)")
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// refreshTickersCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// refreshTickersCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
