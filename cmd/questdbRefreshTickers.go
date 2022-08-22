// Package cmd /*
package cmd

import (
	"context"
	"fmt"
	"github.com/schollz/progressbar/v3"
	"lightning/utils/config"
	"lightning/utils/db"
	"lightning/utils/structs"

	"github.com/spf13/cobra"
)

// questdbPubCmd represents the questdbPub command
var questdbPubCmd = &cobra.Command{
	Use:   "questdbRefreshTickers",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("-	questdbRefreshTickers called...")

		// Get the context To-do (empty context)
		ctx := context.TODO()

		// Get a progress bar
		bar := progressbar.Default(30, "Inserting into qdb...")

		// Get the apiKey from the config.ini file
		apiKey := config.SetPolygonCred("loving_aryabhata_key")

		// Make a channel that will store all the results, of the flattened type
		Tickerchan := make(chan structs.TickersStruct, 35)

		// For each of the tickers, send it to the db
		go func() {
			for t := range Tickerchan {
				db.QDBInsertTickersILP(ctx, t)

				// Update the progress bar
				err := bar.Add(1)
				db.CheckErr(err)
			}
		}()

		// Get the channel that does all the fetching
		db.FetchAllTickers(apiKey, Tickerchan)
		close(Tickerchan)
	},
}

func init() {
	rootCmd.AddCommand(questdbPubCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// questdbPubCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// questdbPubCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
