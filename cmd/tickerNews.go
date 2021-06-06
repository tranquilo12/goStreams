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
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
	"lightning/utils/config"
	"lightning/utils/db"
	"lightning/utils/structs"
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
		fmt.Println("refreshTickerNews2 called")

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

		// read the ticker, from and to dates
		newsParams := db.ReadTickerNewsParamsFromCMD(cmd)

		apiKey := config.SetPolygonCred("other")

		bar := progressbar.Default(int64(len(newsParams.Tickers)))
		for _, ticker := range newsParams.Tickers {
			url := db.MakeTickerNews2Query(apiKey, ticker, newsParams.From)
			Chan1 := db.MakeAllTickerNews2Requests(url)

			err = db.PushTickerNews2IntoDB(Chan1, postgresDB)
			if err != nil {
				panic(err)
			}

			err = bar.Add(1)
			if err != nil {
				panic(err)
			}
		}

	},
}

func init() {
	rootCmd.AddCommand(tickerNewsCmd)
	// Here you will define your flags and configuration settings.
	tickerNewsCmd.Flags().StringP("dbtype", "d", "ec2db", "One of two... ec2db or localdb")
	tickerNewsCmd.Flags().StringSliceP("ticker", "T", []string{}, "Which ticker? (eg. AAPL)")
	tickerNewsCmd.Flags().StringP("from", "f", "", "From which date? (format = %Y-%m-%d)")
	tickerNewsCmd.Flags().StringP("to", "t", "", "To which date? (format = %Y-%m-%d)")

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// refreshTickersCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// refreshTickersCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
