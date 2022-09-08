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

		// If memory profiling is enabled, start the profiler
		if memProfile {
			fmt.Println("memprofile flag set, Profiling memory...")
			config.MemProfiler()
		}

		// Get the context
		fmt.Println("aggsPub called")
		ctx := context.TODO()

		// Update the urls table with the data already present in aggs
		isEmpty := db.QDBCheckAggsLenPG(ctx)
		if !isEmpty {
			db.QDBCheckAggsUrlsPG(ctx)
		}

		// Fetch all urls that have not been pulled yet
		urls := db.QDBFetchUrls(ctx)

		// Download all data and push the data into kafka
		err := publisher.AggKafkaWriter(urls, "aggs", memProfile)
		db.CheckErr(err)
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
