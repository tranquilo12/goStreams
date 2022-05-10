// Package cmd /*
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
	"github.com/spf13/cobra"
	"lightning/subscriber"
	"lightning/utils/db"
)

func Check(err error) {
	if err != nil {
		panic(err)
	}
}

// aggsSub2Cmd represents the aggsPub2 command
var aggsSub2Cmd = &cobra.Command{
	Use:   "aggsSub2",
	Short: "Helps pull data from the Kafka topic to the InfluxDB database",
	Long: `
		This command helps pull data from the Kafka topic to the InfluxDB database.
		Future versions will include a command line interface to the Kafka topic.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("aggsSub2 called")
		readerConns := subscriber.CreateKafkaReaderConn("agg", "g1")

		fmt.Println("Getting influxDB client...")
		influxDBClient := db.GetInfluxDBClient(true)
		defer influxDBClient.Close()

		fmt.Println("Starting to read from Kafka topic and pushing to InfluxDB...")
		subscriber.WriteFromKafkaToInfluxDB(readerConns, influxDBClient)
	},
}

func init() {
	rootCmd.AddCommand(aggsSub2Cmd)

	// Here you will define your flags and configuration settings.
	//aggsSub2Cmd.Flags().StringP("dbtype", "d", "ec2db", "One of two... ec2db or localdb")

	// Get agg parameters from console

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// aggsSub2Cmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// aggsSub2Cmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
