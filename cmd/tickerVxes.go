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
	"lightning/utils/db"
	"time"

	"github.com/spf13/cobra"
)

const (
	TimeLayout = "2006-01-02"
)

// tickerVxesCmd represents the tickerVxes command
var tickerVxesCmd = &cobra.Command{
	Use:   "tickerVxes",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("tickerVxes called")

		timespan, _ := cmd.Flags().GetString("timespan")
		if timespan == "" {
			timespan = "min"
		}

		from_, _ := cmd.Flags().GetString("from")
		if from_ == "" {
			from_ = "2021-01-01"
		}

		to_, _ := cmd.Flags().GetString("to")
		if to_ == "" {
			to_ = "2021-03-01"
		}

		// make multiplier 1 always
		multiplier, _ := cmd.Flags().GetInt("mult")
		if multiplier == 2 {
			multiplier = 1
		}

		// get database conn
		DBParams := db.ReadPostgresDBParamsFromCMD(cmd)
		postgresDB := db.GetPostgresDBConn(&DBParams)
		defer postgresDB.Close()

		estLoc, _ := time.LoadLocation("America/New_York")
		to, _ := time.ParseInLocation(TimeLayout, to_, estLoc)
		from, _ := time.ParseInLocation(TimeLayout, from_, estLoc)

		urls := db.MakeAllTickersVxSourceQueries(apiKey, from, to)
		unexpandedChan := db.MakeAllTickersVxRequests(urls)

		err := db.PushTickerVxIntoDB(unexpandedChan, postgresDB)
		if err != nil {
			panic(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(tickerVxesCmd)

	// Here you will define your flags and configuration settings.
	tickerVxesCmd.Flags().StringP("user", "u", "", "Postgres username")
	tickerVxesCmd.Flags().StringP("password", "P", "", "Postgres password")
	tickerVxesCmd.Flags().StringP("database", "d", "", "Postgres database name")
	tickerVxesCmd.Flags().StringP("host", "H", "127.0.0.1", "Postgres host (default localhost)")
	tickerVxesCmd.Flags().StringP("port", "p", "5432", "Postgres port (default 5432)")

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// tickerVxesCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// tickerVxesCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
