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
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/schollz/progressbar/v3"
	"go.uber.org/ratelimit"
	"lightning/publisher"
	"lightning/subscriber"
	"lightning/utils/db"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
)

// S3DeleteObjectAPI defines the interface for the DeleteObject function.
// We use this interface to test the function using a mocked service.
type S3DeleteObjectAPI interface {
	DeleteObject(ctx context.Context,
		params *s3.DeleteObjectInput,
		optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
}

// DeleteItem deletes an object from an Amazon Simple Storage Service (Amazon S3) bucket
// Inputs:
//     c is the context of the method call, which includes the AWS Region
//     api is the interface that defines the method call
//     input defines the input arguments to the service call.
// Output:
//     If success, a DeleteObjectOutput object containing the result of the service call and nil
//     Otherwise, an error from the call to DeleteObject
func DeleteItem(c context.Context, api S3DeleteObjectAPI, input *s3.DeleteObjectInput) (*s3.DeleteObjectOutput, error) {
	return api.DeleteObject(c, input)
}

// deleteObjsCmd represents the deleteObjs command
var deleteObjsCmd = &cobra.Command{
	Use:   "deleteObjs",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("deleteObjs called")
		var toDelete []string
		dbType, _ := cmd.Flags().GetString("dbtype")
		if dbType == "" {
			dbType = "ec2db"
		}

		// Get agg parameters from cli
		aggParams := db.ReadAggregateParamsFromCMD(cmd)
		insertDate := aggParams.ForceInsertDate
		insertDate = strings.Replace(insertDate, "-", "/", -1)
		prefix := fmt.Sprintf("aggs/%s/", insertDate)

		fmt.Printf("Fetching all objects from %s with prefix %s ...\n", "polygonio-all", prefix)
		allObjs := subscriber.ListAllBucketObjsS3("polygonio-all", prefix)
		for _, url := range *allObjs {
			if url[:16] == prefix {
				toDelete = append(toDelete, url)
			}
		}

		// Create s3Client
		s3Client := publisher.CreateS3Client()

		// use WaitGroup to make things more smooth with goroutines
		var wg sync.WaitGroup

		// create a buffer of the waitGroup, of the same length as urls
		wg.Add(len(toDelete))

		bucket := "polygonio-all"

		// create a rate limiter to stop over-requesting
		prev := time.Now()
		rateLimiter := ratelimit.New(aggParams.Limit)

		fmt.Printf("Deleting all objs found...\n")
		bar := progressbar.Default(int64(len(toDelete)))
		for _, delKey := range toDelete {
			now := rateLimiter.Take()

			go func(k string) {
				input := &s3.DeleteObjectInput{
					Bucket: aws.String(bucket),
					Key:    aws.String(k),
				}

				_, err := DeleteItem(context.TODO(), s3Client, input)
				if err != nil {
					fmt.Println("Got an error deleting item:")
					fmt.Println(err)
					return
				}

				err = bar.Add(1)
				if err != nil {
					return
				}

				wg.Done()
			}(delKey)

			now.Sub(prev)
			prev = now
		}
	},
}

func init() {
	rootCmd.AddCommand(deleteObjsCmd)

	// Here you will define your flags and configuration settings.
	deleteObjsCmd.Flags().StringP("dbtype", "d", "ec2db", "One of two... ec2db or localdb")
	deleteObjsCmd.Flags().StringP("timespan", "T", "", "Timespan (minute, hour, day...)")
	deleteObjsCmd.Flags().StringP("from", "f", "", "From which date? (format = %Y-%m-%d)")
	deleteObjsCmd.Flags().StringP("to", "t", "", "To which date? (format = %Y-%m-%d)")
	deleteObjsCmd.Flags().IntP("mult", "m", 2, "Multiplier to use with Timespan")
	deleteObjsCmd.Flags().IntP("limit", "l", 300, "Rate limit to pull from polygonio")
	deleteObjsCmd.Flags().IntP("withLinearDates", "w", 1, "Usually 1, if appending datasets day-to-day, but if for backup, use 0.")
	deleteObjsCmd.Flags().StringP("forceInsertDate", "F", "", "Force an insert date, to overwrite past data?")

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// deleteObjsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// deleteObjsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
