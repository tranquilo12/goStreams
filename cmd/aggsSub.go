package cmd

import (
	"context"
	"fmt"
	_ "github.com/segmentio/kafka-go/snappy"
	"github.com/spf13/cobra"
	"lightning/subscriber"
	"lightning/utils/db"
)

// aggsSub2Cmd represents the aggsPub2 command
var aggsSubCmd = &cobra.Command{
	Use:   "aggsSub",
	Short: "Helps pull data from the Kafka topic to the QuestDB database",
	Long: `
		This command helps pull data from the Kafka topic to the QuestDB database.
		Future versions will include a command line interface to the Kafka topic.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("aggsSub called")
		ctx := context.TODO()

		// Fetch all urls that have not been pulled yet
		urls := db.QDBFetchUrls(ctx)

		fmt.Println("-	Starting to read from Kafka topic and pushing to QuestDB...")
		subscriber.WriteFromKafkaToQuestDB("aggs", urls)
	},
}

func init() {
	rootCmd.AddCommand(aggsSubCmd)
}
