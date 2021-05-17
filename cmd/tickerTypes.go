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
	"lightning/utils/db"
)

const (
	// make this more secure
	apiKey = "9AheK9pypnYOf_DU6TGpydCK6IMEVkIw"
)

// tickertypesCmd represents the tickertypes command
var tickertypesCmd = &cobra.Command{
	Use:   "tickerTypes",
	Short: "Update all ticker types in the database",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("tickertypes called")

		// get database conn
		DBParams := db.ReadPostgresDBParamsFromCMD(cmd)
		postgresDB := db.GetPostgresDBConn(&DBParams)
		defer postgresDB.Close()

		insertIntoDB := db.MakeTickerTypesRequest(apiKey)

		err := db.PushTickerTypesIntoDB(insertIntoDB, postgresDB)
		if err != nil {
			panic(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(tickertypesCmd)

	// Here you will define your flags and configuration settings.
	tickertypesCmd.Flags().StringP("user", "u", "", "Postgres username")
	tickertypesCmd.Flags().StringP("password", "P", "", "Postgres password")
	tickertypesCmd.Flags().StringP("database", "d", "", "Postgres database name")
	tickertypesCmd.Flags().StringP("host", "H", "127.0.0.1", "Postgres host (default localhost)")
	tickertypesCmd.Flags().StringP("port", "p", "5432", "Postgres port (default 5432)")

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// tickertypesCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// tickertypesCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
