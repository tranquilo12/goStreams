package publisher

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/streadway/amqp"
	"go.uber.org/ratelimit"
	"net/http"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gosuri/uiprogress"
	"lightning/utils/structs"
	"net/url"
	"time"
)

func CreateAggKey(url string) string {
	splitUrl := strings.Split(url, "/")
	ticker := splitUrl[6]
	multiplier := splitUrl[8]
	timespan := splitUrl[9]
	from_ := splitUrl[10]
	//to_ := splitUrl[9]
	today := time.Now().Format("2006-01-02")
	newKey := fmt.Sprintf("aggs/inserted-on-%s/%s-%s/data-for-date-%s/%s", today, timespan, multiplier, from_, ticker)
	return newKey
}

func UploadToS3(bucket string, key string, body []byte) error {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithSharedConfigProfile("default"), config.WithRegion("eu-central-1"))
	if err != nil {
		panic(err)
	}

	// Define a strategy that will buffer 1Mib into memory
	uploader := manager.NewUploader(s3.NewFromConfig(cfg), func(u *manager.Uploader) {
		u.BufferProvider = manager.NewBufferedReadSeekerWriteToPool(1 * 1024 * 1024)
	})

	_, err = uploader.Upload(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(body),
	})
	if err != nil {
		panic(err)
	}
	return nil
}

func AggPublisher(urls []*url.URL) error {

	//AmqpServerUrl := "amqp://guest:guest@localhost:5672"
	//connectRabbitMQ, err := amqp.Dial(AmqpServerUrl)
	//if err != nil {
	//	return err
	//}
	//defer connectRabbitMQ.Close()
	//
	//channelRabbitMQ, err := connectRabbitMQ.Channel()
	//if err != nil {
	//	return err
	//}
	//defer channelRabbitMQ.Close()
	//
	//_, err = channelRabbitMQ.QueueDeclare(
	//	"AGG",
	//	true,
	//	false,
	//	false,
	//	false,
	//	nil,
	//)
	//if err != nil {
	//	panic(err)
	//}
	//
	////err = channelRabbitMQ.QueueBind(
	////	q.Name,
	////	"",
	////	"FANNEDOUT",
	////	false,
	////	nil,
	////)
	////if err != nil {
	////	return err
	////}
	//
	//err = channelRabbitMQ.Qos(10, 0, false)
	//if err != nil {
	//	return err
	//}

	// use WaitGroup to make things more smooth with goroutines
	var wg sync.WaitGroup

	// create a buffer of the waitGroup, of the same length as urls
	wg.Add(len(urls))

	// create a rate limiter to stop over-requesting
	prev := time.Now()
	rateLimiter := ratelimit.New(30)

	// create a new bar and prepend the task progress to the bar and fanout into 1k go routines
	//waitTime := time.Millisecond * 100
	count := len(urls)
	bar := uiprogress.AddBar(count).AppendCompleted().PrependElapsed()
	//bar.PrependFunc(func(b *uiprogress.Bar) string {
	//	return fmt.Sprintf("Task (%d/%d)", b.Current(), count)
	//})
	for i, u := range urls {
		now := rateLimiter.Take()
		target := new(structs.AggregatesBarsResponse)

		go func(u *url.URL, i int) {
			defer wg.Done()
			resp, err := http.Get(u.String())
			if err != nil {
				fmt.Println("Error retrieving URL: ", err)
				panic(err)
			} else {
				// create the key
				messageKey := CreateAggKey(u.String())

				// Marshal target to bytes
				err = json.NewDecoder(resp.Body).Decode(&target)
				taskBytes, err := json.Marshal(target)
				if err != nil {
					fmt.Println("Error retrieving URL: ", err)
				}

				err = UploadToS3("polygonio-all", messageKey, taskBytes)
				err = bar.Set(i)
				if err != nil {
					panic(err)
				}

				//// Create message, and publish
				//message := amqp.Publishing{
				//	ContentType: "application/json",
				//	Body:        taskBytes,
				//}
				//if err := channelRabbitMQ.Publish(
				//	"",
				//	"AGG",
				//	false,
				//	false,
				//	message,
				//); err != nil {
				//	panic(err)
				//}
			}
		}(u, i)

		now.Sub(prev)
		prev = now
	}

	wg.Wait()
	uiprogress.Stop()

	return nil
}

func TickerNewsPublisherRMQ(urls []*url.URL) error {
	AmqpServerUrl := "amqp://guest:guest@localhost:5672"
	connectRabbitMQ, err := amqp.Dial(AmqpServerUrl)
	if err != nil {
		return err
	}
	defer connectRabbitMQ.Close()

	channelRabbitMQ, err := connectRabbitMQ.Channel()
	if err != nil {
		return err
	}
	defer channelRabbitMQ.Close()

	_, err = channelRabbitMQ.QueueDeclare(
		"NEWS",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		panic(err)
	}

	//err = channelRabbitMQ.QueueBind(
	//	q.Name,
	//	"",
	//	"FANNEDOUT",
	//	false,
	//	nil,
	//)
	//if err != nil {
	//	return err
	//}

	err = channelRabbitMQ.Qos(10, 0, false)
	if err != nil {
		return err
	}

	// use WaitGroup to make things more smooth with goroutines
	var wg sync.WaitGroup

	// create a buffer of the waitGroup, of the same length as urls
	wg.Add(len(urls))

	// create a rate limiter to stop over-requesting
	prev := time.Now()
	rateLimiter := ratelimit.New(30)

	for _, u := range urls {
		now := rateLimiter.Take()
		target := new(structs.TickerNews2Response)

		go func(u *url.URL) {
			defer wg.Done()
			resp, err := http.Get(u.String())
			if err != nil {
				fmt.Println("Error retrieving URL: ", err)
				panic(err)
			} else {
				// create the key
				//messageKey := CreateAggKey(u)

				// Marshal target to bytes
				err = json.NewDecoder(resp.Body).Decode(&target)
				taskBytes, err := json.Marshal(target)
				if err != nil {
					fmt.Println("Error retrieving URL: ", err)
				}

				// Create message, and publish
				message := amqp.Publishing{
					ContentType: "application/json",
					Body:        taskBytes,
				}
				if err := channelRabbitMQ.Publish(
					"",
					"AGG",
					false,
					false,
					message,
				); err != nil {
					panic(err)
				}

			}
		}(u)

		now.Sub(prev)
		prev = now
	}

	wg.Wait()

	return nil
}
