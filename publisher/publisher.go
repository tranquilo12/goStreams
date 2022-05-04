package publisher

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gomodule/redigo/redis"
	"github.com/schollz/progressbar/v3"
	"go.uber.org/ratelimit"
	"io"
	"lightning/utils/db"
	"lightning/utils/structs"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type Downloaded struct {
	Key  string
	Body []byte
}

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

// DownloadFromPolygonIO downloads the prices from PolygonIO
func DownloadFromPolygonIO(
	u url.URL,
	forceInsertDate string,
	adjusted int,
	res *structs.AggregatesBarsResponse,
) structs.RedisAggBarsResults {
	resp, err := http.Get(u.String())
	if err != nil {
		panic(err)
	} else {
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				panic(err)
			}
		}(resp.Body)
		messageKey := CreateAggKey(u.String(), forceInsertDate, adjusted)
		err = json.NewDecoder(resp.Body).Decode(&res)
		return structs.RedisAggBarsResults{InsertThis: res.Results, Key: messageKey}
	}
}

func AggDownloader(urls []*url.URL, forceInsertDate string, adjusted int, pool *redis.Pool) error {
	// use WaitGroup to make things more smooth with goroutines
	var wg sync.WaitGroup

	// create a buffer of the waitGroup, of the same length as urls
	wg.Add(len(urls))

	// Max allow 1000 requests per second
	prev := time.Now()
	rateLimiter := ratelimit.New(5000)

	bar := progressbar.Default(int64(len(urls)))
	for _, u := range urls {
		now := rateLimiter.Take()

		// Just sleep for 10 milliseconds, will add jitter
		time.Sleep(time.Millisecond * 20)

		go func(urls *url.URL, p *redis.Pool) {
			// Defer the waitGroup.Done() call until the end of the function
			defer wg.Done()

			// Download the data from PolygonIO
			oneKey := DownloadFromPolygonIO(
				*urls,
				forceInsertDate,
				adjusted,
				&structs.AggregatesBarsResponse{},
			)

			// Convert the data to JSONBytes
			resBytes, err := json.Marshal(oneKey.InsertThis)
			Check(err)

			// Set the key in Redis
			//args := []interface{}{oneKey.Key, resBytes}
			err = db.Set(p, oneKey.Key, resBytes)
			//_ = db.ProcessRedisCommand[[]string](p, "SET", args, false, "string")

			// Update the progress bar
			err = bar.Add(1)
			Check(err)

		}(u, pool)

		now.Sub(prev)
		prev = now
	}
	wg.Wait()
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
