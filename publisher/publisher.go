package publisher

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/schollz/progressbar/v3"
	"go.uber.org/ratelimit"
	"lightning/subscriber"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"lightning/utils/structs"
	"net/url"
	"time"
)

func CreateS3Client() *s3.Client {
	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithRegion("eu-central-1"),
	)
	if err != nil {
		panic(err)
	}

	s3Client := s3.NewFromConfig(cfg)
	return s3Client
}

func UploadToS3(bucket string, key string, body []byte) error {
	s3Client := CreateS3Client()

	_, err := s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(body),
		ContentType: aws.String("application/json"),
	})
	if err != nil {
		panic(err)
	}

	return nil
}

// S3ListObjectsAPI defines the interface for the ListObjectsV2 function. Tests the function using a mocked service.
type S3ListObjectsAPI interface {
	ListObjectsV2(ctx context.Context,
		params *s3.ListObjectsV2Input,
		optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error)
}

// ListObjects retrieves the objects in an Amazon Simple Storage Service (Amazon S3) bucket
// Inputs:
//     c is the context of the method call, which includes the AWS Region
//     api is the interface that defines the method call
//     input defines the input arguments to the service call.
// Output:
//     If success, a ListObjectsV2Output object containing the result of the service call and nil
//     Otherwise, nil and an error from the call to ListObjectsV2
func ListObjects(c context.Context, api S3ListObjectsAPI, input *s3.ListObjectsV2Input) (*s3.ListObjectsV2Output, error) {
	return api.ListObjectsV2(c, input)
}

func GetAggTickersFromS3(insertDate string, timespan string, multiplier int, from_ string, to_ string, adjusted int) *[]string {
	var results []string

	s3Client := CreateS3Client()

	from_ = strings.Replace(from_, "-", "/", -1)
	to_ = strings.Replace(to_, "-", "/", -1)
	insertDate = strings.Replace(insertDate, "-", "/", -1)
	newKey := fmt.Sprintf("aggs/%s/%s/%d/%s/%s", insertDate, timespan, multiplier, from_, to_)

	if adjusted == 1 {
		newKey = fmt.Sprintf("aggs/adj/%s/%s/%d/%s/%s", insertDate, timespan, multiplier, from_, to_)
	}

	input := &s3.ListObjectsV2Input{
		Bucket: aws.String("polygonio-all"),
		Prefix: aws.String(newKey),
	}

	resp, err := ListObjects(context.TODO(), s3Client, input)
	if err != nil {
		fmt.Println("Got error retrieving list of objects:")
		fmt.Println(err)
	}

	for _, item := range resp.Contents {
		splt := strings.Split(*item.Key, "/")
		tkr := splt[len(splt)-2]
		results = append(results, tkr)
	}

	nextContToken := resp.NextContinuationToken
	for {
		if nextContToken != nil {
			input2 := &s3.ListObjectsV2Input{
				Bucket:            aws.String("polygonio-all"),
				Prefix:            aws.String(newKey),
				ContinuationToken: nextContToken,
			}

			resp2, err := ListObjects(context.TODO(), s3Client, input2)
			if err != nil {
				panic(err)
			}

			for _, item := range resp2.Contents {
				splt := strings.Split(*item.Key, "/")
				tkr := splt[len(splt)-2]
				results = append(results, tkr)
			}

			nextContToken = resp2.NextContinuationToken
		} else {
			break
		}
	}

	return &results
}

func AggPublisher(urls []*url.URL, limit int, forceInsertDate string, adjusted int) error {

	// use WaitGroup to make things more smooth with goroutines
	var wg sync.WaitGroup

	// create a buffer of the waitGroup, of the same length as urls
	wg.Add(len(urls))

	// create a rate limiter to stop over-requesting
	prev := time.Now()
	rateLimiter := ratelimit.New(limit)

	s3Client := CreateS3Client()

	bar := progressbar.Default(int64(len(urls)))
	for _, u := range urls {
		now := rateLimiter.Take()
		target := new(structs.AggregatesBarsResponse)

		go func(u *url.URL) {
			resp, err := http.Get(u.String())

			if err != nil {
				fmt.Println("Error retrieving URL (writing to file ./urlErrors.log): ", err.Error())
				f, err := os.OpenFile("urlErrors.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					log.Println(err)
				}
				defer func(f *os.File) {
					err := f.Close()
					if err != nil {
						panic(err)
					}
				}(f)
				logger := log.New(f, "URL-ERROR: ", log.LstdFlags)
				logger.Println(err.Error)
			} else {
				// create the key
				messageKey := subscriber.CreateAggKey(u.String(), forceInsertDate, adjusted)

				// Marshal targets to bytes
				err = json.NewDecoder(resp.Body).Decode(&target)
				taskBytes, err := json.Marshal(target)
				if err != nil {
					fmt.Println("Error retrieving URL: ", err)
				}

				_, err = s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
					Bucket:      aws.String("polygonio-all"),
					Key:         aws.String(messageKey),
					Body:        bytes.NewReader(taskBytes),
					ContentType: aws.String("application/json"),
				})
				if err != nil {
					println(err)
				}

				err = bar.Add(1)
				if err != nil {
					return
				}

				time.Sleep(5 * time.Millisecond)
			}
			wg.Done()
		}(u)

		now.Sub(prev)
		prev = now
	}

	wg.Wait()

	return nil
}

func Unique2dStr(strSlice [][]string) [][]string {
	k := make(map[string][]string)
	txs := make([][]string, 0, len(k))
	for _, r := range strSlice {
		combo := r[0] + "-" + r[1]
		k[combo] = r
	}
	for _, tx := range k {
		txs = append(txs, tx)
	}
	return txs
}
