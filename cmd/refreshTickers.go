package cmd

import (
	"context"
	"fmt"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
	"lightning/utils/config"
	"lightning/utils/db"
	"lightning/utils/structs"
)

// refreshTickersCmd represents the questDBPub command
var refreshTickersCmd = &cobra.Command{
	Use:   "refreshTickers",
	Short: "A command to refresh the tickers in the database",
	Long: `
	Refreshes the tickers in the database.
	First reads from the config.ini file to get the apiKey, 
	then fetches the ticker symbols from the polygon.io API.
`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("-	refreshTickers called...")

		// Get the context To-do (empty context)
		ctx := context.TODO()

		// Get a progress bar
		bar := progressbar.Default(30, "Inserting into qdb...")

		// Get the apiKey from the config.ini file
		apiKey := config.SetPolygonCred("loving_aryabhata_key")

		// Make a channel that will store all the results, of the flattened type
		TickerChan := make(chan structs.TickersStruct, 35)

		// For each of the tickers, send it to the db
		go func() {
			for t := range TickerChan {
				db.QDBInsertTickersILP(ctx, t)

				// Update the progress bar
				err := bar.Add(1)
				db.CheckErr(err)
			}
		}()

		// Get the channel that does all the fetching
		db.FetchAllTickers(apiKey, TickerChan)
		close(TickerChan)
	},
}

func init() {
	rootCmd.AddCommand(refreshTickersCmd)
}
