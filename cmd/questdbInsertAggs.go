package cmd

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"lightning/publisher"
	"lightning/utils/config"
	"lightning/utils/db"
)

// questdbInsertAggsCmd represents the questdbInsertAggs command
var questdbInsertAggsCmd = &cobra.Command{
	Use:   "questdbInsertAggs",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("-	questdbInsertAggs called...")
		ctx := context.TODO()

		// Get agg parameters from cli
		qdbAggParams := db.ReadAggregateParamsFromCMD(cmd)

		// Create the table if not exists... as well as the constraints...
		//db.QDBCreateAggTable(ctx)

		// Get the apiKey from the config.ini file
		apiKey := config.SetPolygonCred("loving_aryabhata_key")

		// Now get all the tickers from QDB
		tickers := db.QDBFetchUniqueTickersPG(ctx)

		//Make all urls from the tickers
		urls := db.MakeAllStocksAggsUrls(
			tickers,
			qdbAggParams.Timespan,
			qdbAggParams.From,
			qdbAggParams.To,
			apiKey,
			qdbAggParams.Adjusted,
		)

		// Push into questDB
		publisher.QDBPushAllAggIntoDB(ctx, urls, qdbAggParams.Timespan, qdbAggParams.Multiplier)
	},
}

func init() {
	rootCmd.AddCommand(questdbInsertAggsCmd)
	questdbInsertAggsCmd.Flags().StringP("timespan", "T", "", "Timespan (minute, hour, day...)")
	questdbInsertAggsCmd.Flags().StringP("from", "f", "", "From which date? (format = %Y-%m-%d)")
	questdbInsertAggsCmd.Flags().StringP("to", "t", "", "To which date? (format = %Y-%m-%d)")
	questdbInsertAggsCmd.Flags().IntP("adjusted", "a", 1, "Adjusted? (1/0)")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// questdbInsertAggsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// questdbInsertAggsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
