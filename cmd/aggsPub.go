package cmd

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"lightning/publisher"
	"lightning/utils/config"
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

		// Get the logger here
		logger := config.GetLogger()

		// Fetch all urls that have not been pulled yet
		logger.Info("Fetching all urls that have not been pulled yet...")
		urls := db.QDBFetchUrls(ctx, false, 1000)

		// Download all data and push the data into kafka
		err := publisher.AggChannelWriter(urls, logger)
		db.CheckErr(err)
	},
}

func init() {
	// Here you will define your flags and configuration settings.
	rootCmd.AddCommand(aggsPubCmd)
}
