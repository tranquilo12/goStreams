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
	"lightning/utils/db"

	"github.com/spf13/cobra"
)

// createTablesCmd represents the createTables command
var createTablesCmd = &cobra.Command{
	Use:   "createTables",
	Short: "For a clean database, run this command to create the tables in default database 'postgres'",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("createTables called")

		user, _ := cmd.Flags().GetString("user")
		if user == "" {
			user = "postgres"
		}

		password, _ := cmd.Flags().GetString("password")
		if password == "" {
			panic("Cmon, pass a password!")
		}

		database, _ := cmd.Flags().GetString("database")
		if database == "" {
			database = "postgres"
		}

		host, _ := cmd.Flags().GetString("host")
		if host == "" {
			host = "127.0.0.1"
		}

		port, _ := cmd.Flags().GetString("port")
		if port == "" {
			port = "5432"
		}

		db.ExecCreateAllTablesModels(user, password, database, host, port)
	},
}

func init() {
	rootCmd.AddCommand(createTablesCmd)

	// Here you will define your flags and configuration settings.
	createTablesCmd.Flags().StringP("user", "u", "", "Postgres username")
	createTablesCmd.Flags().StringP("password", "P", "", "Postgres password")
	createTablesCmd.Flags().StringP("database", "d", "", "Postgres database name")
	createTablesCmd.Flags().StringP("host", "H", "127.0.0.1", "Postgres host (default localhost)")
	createTablesCmd.Flags().StringP("port", "p", "5432", "Postgres port (default 5432)")

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// createTablesCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// createTablesCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
