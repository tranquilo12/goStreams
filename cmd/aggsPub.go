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
		// Check if memory profiling is enabled
		memProfile, _ := cmd.Flags().GetBool("memprofile")

		// Get the context
		fmt.Println("aggsPub called")
		ctx := context.TODO()

		// If memory profiling is enabled, start the profiler
		if memProfile {
			fmt.Println("memprofile flag set, Profiling memory...")
			config.MemProfiler(ctx)
		}

		//// Update the urls table with the data already present in aggs
		//isEmpty := db.QDBCheckAggsLenPG(ctx)
		//if !isEmpty {
		//	db.QDBCheckAggsUrlsPG(ctx)
		//}

		// Get the logger here
		logger := config.GetLogger()

		// Fetch all urls that have not been pulled yet
		logger.Info("Fetching all urls that have not been pulled yet...")
		urls := db.QDBFetchUrls(ctx)

		// Download all data and push the data into kafka
		err := publisher.AggKafkaWriter(urls, "aggs", memProfile, logger)
		db.CheckErr(err)

		//db.Combined(ctx, logger)
	},
}

func init() {
	// Here you will define your flags and configuration settings.
	rootCmd.AddCommand(aggsPubCmd)
	aggsPubCmd.Flags().BoolP(
		"memprofile",
		"m",
		false,
		"To enable memory profiling, set this flag to true",
	)
}
