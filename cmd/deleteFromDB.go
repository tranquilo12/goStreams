package cmd

/*
Copyright © 2022 Shriram Sunder <shriram.sunder121091@gmail.com>

*/

import (
	"fmt"
	"github.com/spf13/cobra"
	"lightning/utils/db"
)

// deleteFromDBCmd represents the deleteFromDB command
var deleteFromDBCmd = &cobra.Command{
	Use:   "deleteFromDB",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {

		fmt.Println("deleteFromDB called")
		measurement, _ := cmd.Flags().GetString("measurement")
		if measurement == "" {
			measurement = "aggregates"
		}

		// Get From/To date
		from_, _ := cmd.Flags().GetString("from")
		if from_ == "" {
			from_ = "2021-01-01"
		}

		to_, _ := cmd.Flags().GetString("to")
		if to_ == "" {
			to_ = "2022-06-01"
		}

		fmt.Println("Get influxDB client...")
		influxDBClient := db.GetInfluxDBClient(true)

		fmt.Printf("Deleting %s (measurement) from influxDB...", measurement)
		db.DeleteFromInfluxDB(influxDBClient, measurement, from_, to_)
	},
}

func init() {
	rootCmd.AddCommand(deleteFromDBCmd)
	deleteFromDBCmd.Flags().StringP("measurement", "m", "", "Which measurement to delete?")
	deleteFromDBCmd.Flags().StringP("from", "f", "", "From which date? (format = %Y-%m-%d)")
	deleteFromDBCmd.Flags().StringP("to", "t", "", "To which date? (format = %Y-%m-%d)")
}
