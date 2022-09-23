package cmd

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
		db.PushAllUrlsToTable(
			ctx,
			tickers,
			params.Timespan,
			params.From,
			params.To,
			apiKey,
			params.Adjusted,
		)

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
}
