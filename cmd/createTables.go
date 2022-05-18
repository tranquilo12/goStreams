package cmd

/*
Copyright © 2021 Shriram Sunder <shriram.sunder121091@gmail.com>

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

// createTablesCmd represents the createTables command
var createTablesCmd = &cobra.Command{
	Use:   "createTables",
	Short: "For a clean database, run this command to create the tables in default database 'postgres'",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("createTables called")

		dbType, _ := cmd.Flags().GetString("dbtype")
		if dbType == "" {
			dbType = "ec2db"
		}

		// get database conn
		DBParams := structs.DBParams{}
		err := config.SetDBParams(&DBParams, dbType)
		if err != nil {
			panic(err)
		}

		db.ExecCreateAllTablesModels(&DBParams)
	},
}

func init() {
	rootCmd.AddCommand(createTablesCmd)
	createTablesCmd.Flags().StringP("dbtype", "d", "", "One of two... ec2db or localdb")
}
