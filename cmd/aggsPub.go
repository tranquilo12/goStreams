package cmd

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"lightning/publisher"
	"lightning/utils/db"
	_ "net/http/pprof"
)

// aggsPubCmd represents the aggs command
var aggsPubCmd = &cobra.Command{
	Use:   "aggsPub",
	Short: "Helps pull data from Polygon-io and into a Kafka topic",
	Long: `
		This command pulls data from Polygon-io and into a Kafka topic.
        Future enhancements will include a command line interface to
		interact with the Kafka topic.
`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get the context
		fmt.Println("aggsPub called")
		ctx := context.TODO()

		// Fetch all urls, ticker by ticker, so fetch all the tickers first
		tickers := db.QDBFetchUniqueTickersPG(ctx)

		for i, ticker := range tickers {
			fmt.Printf("Fetching all data for ticker: %s, %d/%d \n", ticker, i, len(tickers))

			// Get the urls for this ticker
			urls := db.QDBFetchUrlsByTicker(ctx, ticker)

			// Download all agg data and push the data into QuestDB
			err := publisher.AggChannelWriter(urls)
			db.CheckErr(err)
		}
	},
}

func init() {
	// Here you will define your flags and configuration settings.
	rootCmd.AddCommand(aggsPubCmd)
}
