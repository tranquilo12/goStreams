package cmd

/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

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
	"github.com/adjust/rmq/v3"
	"lightning/utils/config"
	"lightning/utils/db"
	"log"
	"net/http"

	"github.com/spf13/cobra"
)

type Handler struct {
	connection rmq.Connection
}

func NewHandler(connection rmq.Connection) *Handler {
	return &Handler{connection: connection}
}

func (handler *Handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	layout := request.FormValue("layout")
	refresh := request.FormValue("refresh")

	queues, err := handler.connection.GetOpenQueues()
	if err != nil {
		panic(err)
	}

	stats, err := handler.connection.CollectStats(queues)
	if err != nil {
		panic(err)
	}

	log.Printf("queue stats\n%s", stats)
	_, err = fmt.Fprint(writer, stats.GetHtml(layout, refresh))
	if err != nil {
		panic(err)
	}
}

// showQCmd represents the showQ command
var showQCmd = &cobra.Command{
	Use:   "showQ",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		var errChan chan error
		fmt.Println("showQ called")

		// also get the redis connection
		redisEndpoint := config.GetRedisParams("ELASTICCACHE")
		redisClient := db.GetRedisClient(7000, redisEndpoint)
		queueConnection, err := rmq.OpenConnectionWithRedisClient("AGG", redisClient, errChan)
		if err != nil {
			fmt.Println("Something wrong with this queueConnection...")
		}

		http.Handle("/overview", NewHandler(queueConnection))
		fmt.Printf("Handler listening on http://localhost:3333/overview\n")
		if err := http.ListenAndServe(":3333", nil); err != nil {
			panic(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(showQCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// showQCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// showQCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
