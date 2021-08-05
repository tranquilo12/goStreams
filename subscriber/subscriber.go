package subscriber

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/schollz/progressbar/v3"
	"go.uber.org/ratelimit"
	"lightning/publisher"
	"lightning/utils/structs"
	"log"
	"net/url"
	"sync"
	"time"
)

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

func AggDownloader(urls []*url.URL, forceInsertDate string, adjusted int) chan structs.RedisAggBarsResults {

	insertIntoRedisChan := make(chan structs.RedisAggBarsResults, 100000)

	// use WaitGroup to make things more smooth with goroutines
	var wg sync.WaitGroup

	// create a buffer of the waitGroup, of the same length as urls
	wg.Add(len(urls))

	prev := time.Now()
	rateLimiter := ratelimit.New(500)

	bar := progressbar.Default(int64(len(urls)))
	for _, u := range urls {
		now := rateLimiter.Take()

		go func(url *url.URL) {
			defer wg.Done()
			messageKey := publisher.CreateAggKey(url.String(), forceInsertDate, adjusted)
			fromS3 := DownloadFromS3("polygonio-all", messageKey)

			// For example, show received message in a console.
			res := structs.AggregatesBarsResponse{}

			err := json.Unmarshal(fromS3.Bytes(), &res)
			if err != nil {
				panic(err)
			}

			oneKey := structs.RedisAggBarsResults{
				InsertThis: res.Results, Key: messageKey,
			}

			err = bar.Add(1)
			if err != nil {
				return
			}

			time.Sleep(5 * time.Millisecond)
			insertIntoRedisChan <- oneKey
		}(u)

		now.Sub(prev)
		prev = now
	}
	wg.Wait()
	close(insertIntoRedisChan)
	return insertIntoRedisChan
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
