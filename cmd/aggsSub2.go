// Package cmd /*
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
	"strconv"
)

func Check(err error) {
	if err != nil {
		panic(err)
	}
}

// aggsSub2Cmd represents the aggsPub2 command
var aggsSub2Cmd = &cobra.Command{
	Use:   "aggsPub2",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("aggsPub2 called")
		_ = db.ReadAggParamsFromCMD2(cmd)

		// Possibly get all the redis parameters from the .ini file.
		var redisParams config.RedisParams
		err := config.SetRedisCred(&redisParams)
		Check(err)

		// Create a redis client
		port, err := strconv.Atoi(redisParams.Port)
		pool := db.GetRedisPool(port, redisParams.Host)
		Check(err)

		// Push the data to redis
		err = db.PushAggIntoFFSCont(pool)
		Check(err)
	},
}

func init() {
	rootCmd.AddCommand(aggsSub2Cmd)

	// Here you will define your flags and configuration settings.
	aggsSub2Cmd.Flags().StringP("dbtype", "d", "ec2db", "One of two... ec2db or localdb")

	// Get agg parameters from console

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// aggsSub2Cmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// aggsSub2Cmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
