package cmd

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
	"lightning/publisher"
	"lightning/utils/db"
	"math/rand"
	_ "net/http/pprof"
	"sync"
	"time"
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

		// Fetch all urls, ticker by ticker, so fetch all the tickers first
		tickers := db.QDBFetchUniqueTickersPG(ctx)

		// Get new multiple progress bar, with waitgroup
		numBars := 2
		var pbarWg sync.WaitGroup
		p := mpb.New(mpb.WithWaitGroup(&pbarWg))

		// Total number of progress bars = 2, and the
		for i := 0; i < numBars; i++ {
			task := fmt.Sprintf("Task #%d", i)
			queue := make([]*mpb.Bar, 2)

			queue[0] = p.AddBar(rand.Int63n(201)+100,
				mpb.PrependDecorators(
					decor.Name(task, decor.WC{W: len(task) + 1, C: decor.DidentRight}),
					decor.Name("Downloading", decor.WCSyncSpaceR),
					decor.CountersNoUnit("%d / %d", decor.WCSyncWidth),
				),
				mpb.AppendDecorators(decor.Percentage(decor.WC{W: 5})),
			)

			queue[1] = p.AddBar(rand.Int63n(101)+100,
				mpb.BarQueueAfter(queue[0], false), // this bar is queued
				mpb.BarFillerClearOnComplete(),
				mpb.PrependDecorators(
					decor.Name(task, decor.WC{W: len(task) + 1, C: decor.DidentRight}),
					decor.OnComplete(decor.Name("\x1b[31mInserting\x1b[0m", decor.WCSyncSpaceR), "done!"),
					decor.OnComplete(decor.EwmaETA(decor.ET_STYLE_MMSS, 0, decor.WCSyncWidth), ""),
				),
				mpb.AppendDecorators(
					decor.OnComplete(decor.Percentage(decor.WC{W: 5}), ""),
				),
			)

			go func() {
				for _, b := range queue {
					completeIteration(ctx, b, tickers)
				}
			}()
		}

		// Wait for all bars to complete and flush
		p.Wait()
	},
}

func completeIteration(ctx context.Context, bar *mpb.Bar, tickers []string) {
	for !bar.Completed() {
		for _, ticker := range tickers {
			start := time.Now()
			// Get the urls for this ticker
			urls := db.QDBFetchUrlsByTicker(ctx, ticker)

			// Download all agg data and push the data into QuestDB
			err := publisher.AggChannelWriter(urls)
			db.CheckErr(err)
			// we need to call EwmaIncrement to fulfill ewma decorator's contract
			bar.EwmaIncrInt64(rand.Int63n(5)+1, time.Since(start))
		}
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	rootCmd.AddCommand(aggsPubCmd)
}
