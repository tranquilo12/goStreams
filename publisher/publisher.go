package publisher

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gosuri/uiprogress"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/segmentio/kafka-go"
	"go.uber.org/ratelimit"
	"lightning/utils/structs"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

func Check(err error) {
	if err != nil {
		panic(err)
	}
}

func DownloadFromS3(bucket string, key string) *manager.WriteAtBuffer {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithSharedConfigProfile("default"), config.WithRegion("eu-central-1"))
	if err != nil {
		panic(err)
	}

	// Define a strategy that will buffer 1Mib into memory
	downloader := manager.NewDownloader(s3.NewFromConfig(cfg), func(u *manager.Downloader) {
		u.BufferProvider = manager.NewPooledBufferedWriterReadFromProvider(1 * 1024 * 1024)
	})

	buff := &manager.WriteAtBuffer{}
	_, err = downloader.Download(context.TODO(), buff,
		&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		})
	if err != nil {
		panic(err)
	}
	return buff
}

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

func CreateAggKey2(url string) string {
	splitUrl := strings.Split(url, "/")
	ticker := splitUrl[6]
	from_ := splitUrl[10]
	to_ := strings.Split(splitUrl[11], "?")[0]
	return fmt.Sprintf("%s/%s/%s", from_, to_, ticker)
}

// DownloadFromPolygonIO downloads the prices from PolygonIO
func DownloadFromPolygonIO(u url.URL, res *structs.AggregatesBarsResponse) structs.InfluxDBAggBarsResults {
	// Create a new client
	resp, err := http.Get(u.String())
	Check(err)

	// Defer the closing of the body
	defer resp.Body.Close()

	// Decode the response
	err = json.NewDecoder(resp.Body).Decode(&res)
	return structs.InfluxDBAggBarsResults{InsertThis: res.Results, Key: res.Ticker}
}

func AggDownloader(urls []*url.URL, client influxdb2.Client) error {
	// use WaitGroup to make things more smooth with goroutines
	var wg sync.WaitGroup

	// create a buffer of the waitGroup, of the same length as urls
	wg.Add(len(urls))

	// Max allow 1000 requests per second
	prev := time.Now()
	rateLimiter := ratelimit.New(500)

	// Get Write client
	writeAPI := client.WriteAPI("lightning", "Lightning")

	// Start the UI progress bar
	uiprogress.Start()

	// Create ui progress bar, formatted
	bar1 := uiprogress.AddBar(len(urls)).AppendCompleted().PrependElapsed()
	bar1.PrependFunc(func(b *uiprogress.Bar) string {
		return fmt.Sprintf("Pushing data into db (%d/%d)", b.Current(), len(urls))
	})

	// Iterate over every url
	for _, u := range urls {
		// rate limit the requests
		now := rateLimiter.Take()

		go func(urls *url.URL) {
			// Defer the waitGroup
			defer wg.Done()

			// Download the data from PolygonIO
			oneKey := DownloadFromPolygonIO(
				*urls,
				&structs.AggregatesBarsResponse{},
			)

			// Write the data to InfluxDB
			for _, v := range oneKey.InsertThis {
				// Convert the data to influx points
				p := influxdb2.NewPoint(
					"aggregates",
					map[string]string{"ticker": oneKey.Key},
					map[string]interface{}{
						"open": v.O, "high": v.H, "low": v.L, "close": v.C, "vWap": v.Vw, "volume": v.V,
					},
					time.Unix(int64(v.T)/1000, 0),
				)

				// Progress bar2 update
				bar1.Incr()

				// Write the point
				writeAPI.WritePoint(p)

				// Flush write API
				writeAPI.Flush()
			}
		}(u)
		// Rate limit the requests, so note the time
		now.Sub(prev)
		prev = now
	}
	// Wait for all the goroutines to finish
	wg.Wait()

	// Stop the UI progress bar
	uiprogress.Stop()

	return nil
}

func AggKafkaWriter(urls []*url.URL) error {
	// use WaitGroup to make things more smooth with goroutines
	var wg sync.WaitGroup

	// create a buffer of the waitGroup, of the same length as urls
	wg.Add(len(urls))

	// Max allow 1000 requests per second
	prev := time.Now()
	rateLimiter := ratelimit.New(1000)

	// Get Write client
	writeConn := CreateKafkaWriterConn("agg")

	// Start the UI progress bar
	uiprogress.Start()

	// Create ui progress bar, formatted
	bar1 := uiprogress.AddBar(len(urls)).AppendCompleted().PrependElapsed()
	bar1.PrependFunc(func(b *uiprogress.Bar) string {
		return fmt.Sprintf("Pushing data into kafka (%d/%d)", b.Current(), len(urls))
	})

	// Iterate over every url
	for _, u := range urls {
		// rate limit the requests
		now := rateLimiter.Take()

		go func(urls *url.URL) {
			// All messages
			var messages []kafka.Message

			// Defer the waitGroup
			defer wg.Done()

			// Download the data from PolygonIO
			oneKey := DownloadFromPolygonIO(
				*urls,
				&structs.AggregatesBarsResponse{},
			)

			// Write the data to InfluxDB
			for _, v := range oneKey.InsertThis {
				// Convert the data to influx points
				val, err := json.Marshal(v)
				Check(err)

				// Create the key
				key := []byte(oneKey.Key)

				// Create the messages
				messages = append(
					messages,
					kafka.Message{
						Key:   key,
						Value: val,
						Time:  time.Time{},
					},
				)
			}

			// Write the messages to Kafka
			err := writeConn.WriteMessages(context.Background(), messages...)
			Check(err)

			// Progress bar2 update
			bar1.Incr()
		}(u)
		// Rate limit the requests, so note the time
		now.Sub(prev)
		prev = now
	}
	// Wait for all the goroutines to finish
	wg.Wait()

	// Stop the kafka writer
	err := writeConn.Close()
	Check(err)

	// Stop the UI progress bar
	uiprogress.Stop()

	return nil
}

func ListAllBucketObjsS3(bucket string, prefix string) *[]string {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithSharedConfigProfile("default"), config.WithRegion("eu-central-1"))
	if err != nil {
		panic(err)
	}

	client := s3.NewFromConfig(cfg)

	// Set the parameters based on teh CLI flag inputs.
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
