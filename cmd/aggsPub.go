package cmd

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
	"lightning/publisher"
	"lightning/utils/db"
)

// Get new pbar
func getNewPbar(p *mpb.Progress, total int, name string) *mpb.Bar {
	// create a single bar, which will inherit container's width
	bar := p.New(int64(total),
		// BarFillerBuilder with custom style
		mpb.BarStyle().Lbound("╢").Filler("▌").Tip("▌").Padding("░").Rbound("╟"),
		mpb.PrependDecorators(
			// display our name with one space on the right
			decor.Name(name, decor.WC{W: len(name) + 1, C: decor.DidentRight}),
			decor.CountersNoUnit("%d / %d", decor.WCSyncWidth),
			// replace ETA decorator with "done" message, OnComplete event
			decor.OnComplete(
				decor.AverageETA(decor.ET_STYLE_GO, decor.WC{W: 4, C: decor.DSyncSpace}), "Done!",
			),
		),
		mpb.AppendDecorators(decor.Percentage(decor.WC{W: 5})),
	)
	return bar
}

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
		fmt.Println("Downloading data into 'aggs' table...")
		ctx := context.TODO()

		// Fetch all urls, ticker by ticker, so fetch all the tickers first
		tickers := db.QDBFetchUniqueTickersPG(ctx)

		// initialize progress container, with custom width
		p := mpb.New(mpb.WithWidth(64))
		bar := getNewPbar(p, len(tickers), "Downloading :")

		fmt.Printf("Fetching all data for each ticker...")
		for _, ticker := range tickers {
			// Get the urls for this ticker
			urls := db.QDBFetchUrlsByTicker(ctx, ticker)

			// Download all agg data and push the data into QuestDB
			err := publisher.AggChannelWriter(urls)
			db.CheckErr(err)

			// increment bar
			bar.Increment()
		}

		// wait for our bar to complete and flush
		p.Wait()
	},
}

func init() {
	// Here you will define your flags and configuration settings.
	rootCmd.AddCommand(aggsPubCmd)
}
