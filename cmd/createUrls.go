package cmd

/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

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
	"context"
	"fmt"
	"lightning/utils/config"
	"lightning/utils/db"

	"github.com/spf13/cobra"
)

// createUrlsCmd represents the createUrls command
var createUrlsCmd = &cobra.Command{
	Use:   "createUrls",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("createUrls called...")
		ctx := context.TODO()

		// Get agg parameters from cli, so you can populate the urls table
		params := db.ReadAggregateParamsFromCMD(cmd)

		// Get all tickers from QDB
		tickers := db.QDBFetchUniqueTickersPG(ctx)

		// Get the apiKey from the config.ini file
		apiKey := config.SetPolygonCred("loving_aryabhata_key")

		// Push all the urls to the urls table
		db.PushAllUrlsToTable(ctx, tickers, params.Timespan, params.From, params.To, apiKey, params.Adjusted)

		// Mark as done...
		fmt.Println("createUrls finished...")
	},
}

func init() {
	rootCmd.AddCommand(createUrlsCmd)
	createUrlsCmd.Flags().StringP("dbtype", "d", "ec2db", "One of two... ec2db or localdb")
	createUrlsCmd.Flags().StringP("timespan", "T", "", "Timespan (minute, hour, day...)")
	createUrlsCmd.Flags().StringP("from", "f", "", "From which date? (format = %Y-%m-%d)")
	createUrlsCmd.Flags().StringP("to", "t", "", "To which date? (format = %Y-%m-%d)")
	createUrlsCmd.Flags().IntP("adjusted", "a", 1, "Adjusted? (1/0)")
	createUrlsCmd.Flags().IntP("limit", "L", 50000, "Limit?")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// createUrlsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// createUrlsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
