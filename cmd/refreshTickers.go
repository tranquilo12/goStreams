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
	"github.com/spf13/cobra"
	"lightning/utils/config"
	"lightning/utils/db"
	"lightning/utils/structs"
)

const (
	layout = "2006-01-02"
)

// refreshTickersCmd represents the refreshTickers command
var refreshTickersCmd = &cobra.Command{
	Use:   "refreshTickers",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("refreshTickers called")
		var dbParams structs.DBParams
		err := config.SetDBParams(&dbParams, "postgres")
		if err != nil {
			panic(err)
		}

		// Get a db conn that will be used by the function 'PushTickerIntoDB'
		dbConn := db.GetPostgresDBConn(&dbParams)

		apiKey := config.SetPolygonCred("other")
		url := db.MakeTickerVxQuery(apiKey)
		Chan1 := db.MakeAllTickersVxRequests(url)
		err = db.PushTickerVxIntoDB(Chan1, dbConn)
		if err != nil {
			panic(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(refreshTickersCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// refreshTickersCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// refreshTickersCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}