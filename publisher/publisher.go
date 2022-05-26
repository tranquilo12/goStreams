package publisher

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/schollz/progressbar/v3"
	"go.uber.org/ratelimit"
	config2 "lightning/utils/config"
	"lightning/utils/db"
	"lightning/utils/structs"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// CreateAggKey creates the key for the aggregated data
func CreateAggKey(url string, forceInsertDate string, adjusted int) string {
	splitUrl := strings.Split(url, "/")
	ticker := splitUrl[6]
	multiplier := splitUrl[8]
	timespan := splitUrl[9]

	from_ := splitUrl[10]
	fromYear := strings.Split(from_, "-")[0]
	fromMon := strings.Split(from_, "-")[1]
	fromDay := strings.Split(from_, "-")[2]

	to_ := splitUrl[11]
	toYear := strings.Split(to_, "-")[0]
	toMon := strings.Split(to_, "-")[1]
	toDay := strings.Split(to_, "-")[2]
	toDay = strings.Split(toDay, "?")[0]

	insertDate := forceInsertDate
	insertDateYear := strings.Split(insertDate, "-")[0]
	insertDateMon := strings.Split(insertDate, "-")[1]
	insertDateDay := strings.Split(insertDate, "-")[2]

	newKey := fmt.Sprintf("aggs/%s/%s/%s/%s/%s/%s/%s/%s/%s/%s/%s/%s/data.json", insertDateYear, insertDateMon, insertDateDay, timespan, multiplier, fromYear, fromMon, fromDay, toYear, toMon, toDay, ticker)
	if adjusted == 1 {
		newKey = fmt.Sprintf("aggs/adj/%s/%s/%s/%s/%s/%s/%s/%s/%s/%s/%s/%s/data.json", insertDateYear, insertDateMon, insertDateDay, timespan, multiplier, fromYear, fromMon, fromDay, toYear, toMon, toDay, ticker)
	}
	return newKey
}

// DownloadFromPolygonIO downloads the prices from PolygonIO
func DownloadFromPolygonIO(client *http.Client, u url.URL, res *structs.AggregatesBarsResponse) error {
	// Create a new client
	resp, err := client.Get(u.String())
	db.Check(err)

	// Defer the closing of the body
	defer resp.Body.Close()

	// Decode the response
	if resp.StatusCode == http.StatusOK {
		err = json.NewDecoder(resp.Body).Decode(&res)
		db.Check(err)
	}
	return err
}

// AggKafkaWriter writes the aggregates to Kafka
func AggKafkaWriter(urls []*url.URL) error {
	// use WaitGroup to make things more smooth with goroutines
	var wg sync.WaitGroup

	urls = urls[27299+72598+319843+872323+139078:]

	// create a buffer of the waitGroup, of the same length as urls
	wg.Add(len(urls))

	// Max allow 500 requests per second
	prev := time.Now()
	rateLimiter := ratelimit.New(200)

	// Get Write client
	writeConn := CreateKafkaWriterConn("agg")
	defer writeConn.Close()

	// Create ui progress bar, formatted
	bar := progressbar.Default(int64(len(urls)))
	defer bar.Close()

	// Get the http client
	httpClient := config2.GetHttpClient()

	// Iterate over every url
	for _, u := range urls {
		// rate limit the requests
		now := rateLimiter.Take()

		// Create the goroutine
		go KafkaWriter(context.Background(), httpClient, u, writeConn, &wg, bar)

		// Rate limit the requests, so note the time
		now.Sub(prev)
		prev = now
	}
	// Wait for all the goroutines to finish
	wg.Wait()

	return nil
}

// ListAllBucketObjsS3 lists all the objects in a bucket
func ListAllBucketObjsS3(bucket string, prefix string) *[]string {
	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithSharedConfigProfile("default"),
		config.WithRegion("eu-central-1"),
	)
	db.Check(err)

	client := s3.NewFromConfig(cfg)

	// Set the parameters based on the CLI flag inputs.
	params := &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
	}
	if len(prefix) != 0 {
		params.Prefix = aws.String(prefix)
	}

	p := s3.NewListObjectsV2Paginator(client, params, func(o *s3.ListObjectsV2PaginatorOptions) {
		if v := int32(20000); v != 0 {
			o.Limit = v
		}
	})

	var res []string
	var i int
	for p.HasMorePages() {
		i++
		page, err := p.NextPage(context.TODO())
		if err != nil {
			log.Fatalf("Failed to get page %v, %v", i, err)
		}

		for _, obj := range page.Contents {
			res = append(res, *obj.Key)
		}
	}
	return &res
}
