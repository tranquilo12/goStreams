package publisher

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/schollz/progressbar/v3"
	"github.com/streadway/amqp"
	"go.uber.org/ratelimit"
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

//type S3Result struct {
//	Output *s3.PutObjectOutput
//	Err    error
//}

func CreateAggKey(url string) string {
	splitUrl := strings.Split(url, "/")
	ticker := splitUrl[6]
	multiplier := splitUrl[8]
	timespan := splitUrl[9]

	from_ := splitUrl[10]
	fromYear := strings.Split(from_, "-")[0]
	fromMon := strings.Split(from_, "-")[1]
	fromDay := strings.Split(from_, "-")[2]

	today := time.Now().Format("2006-01-02")
	todayYear := strings.Split(today, "-")[0]
	todayMon := strings.Split(today, "-")[1]
	todayDay := strings.Split(today, "-")[2]

	newKey := fmt.Sprintf("aggs/%s/%s/%s/%s/%s/%s/%s/%s/%s/data.json", todayYear, todayMon, todayDay, timespan, multiplier, fromYear, fromMon, fromDay, ticker)
	return newKey
}

func createS3Client() *s3.Client {
	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		//config.WithSharedConfigProfile("default"),
		config.WithRegion("eu-central-1"),
	)
	if err != nil {
		panic(err)
	}

	s3Client := s3.NewFromConfig(cfg)
	return s3Client
}

func UploadToS3(bucket string, key string, body []byte) error {
	//cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithSharedConfigProfile("default"), config.WithRegion("eu-central-1"))
	//if err != nil {
	//	panic(err)
	//}
	//
	//// Define a strategy that will buffer 1Mib into memory
	//uploader := manager.NewUploader(s3.NewFromConfig(cfg), func(u *manager.Uploader) {
	//	u.BufferProvider = manager.NewBufferedReadSeekerWriteToPool(1 * 1024 * 1024)
	//})

	s3Client := createS3Client()

	_, err := s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(body),
		ContentType: aws.String("application/json"),
	})
	if err != nil {
		panic(err)
	}

	//_, err = uploader.Upload(context.TODO(), &s3.PutObjectInput{
	//	Bucket: aws.String(bucket),
	//	Key:    aws.String(key),
	//	Body:   bytes.NewReader(body),
	//})
	//if err != nil {
	//	panic(err)
	//}

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
	rateLimiter := ratelimit.New(300)

	s3Client := createS3Client()

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
				defer f.Close()
				logger := log.New(f, "URL-ERROR: ", log.LstdFlags)
				logger.Println(err.Error())
			} else {
				// create the key
				messageKey := CreateAggKey(u.String())

				// Marshal target to bytes
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
					panic(err)
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

//func AggPublisherS3(urls []*url.URL) error {
//
//	// use WaitGroup to make things more smooth with goroutines
//	var wg sync.WaitGroup
//
//	// create a buffer of the waitGroup, of the same length as urls
//	wg.Add(len(urls))
//
//	// create a rate limiter to stop over-requesting
//	prev := time.Now()
//	rateLimiter := ratelimit.New(30)
//
//	// Create new session
//	svc := db.CreateNewS3Session()
//
//	for _, u := range urls {
//		now := rateLimiter.Take()
//		//target := new(structs.AggregatesBarsResponse)
//
//		go func(u *url.URL) {
//			defer wg.Done()
//			resp, err := http.Get(u.String())
//			if err != nil {
//				fmt.Println("Error retrieving URL: ", err)
//				panic(err)
//			} else {
//				err = db.UploadAggToS3(svc, u, resp)
//				if err != nil {
//					panic(err)
//				}
//
//				// create the key
//				//messageKey := CreateAggKey(u)
//
//				// Marshal target to bytes
//				//err = json.NewDecoder(resp.Body).Decode(&target)
//				//taskBytes, err := json.Marshal(target)
//				//if err != nil {
//				//	fmt.Println("Error retrieving URL: ", err)
//				//}
//
//			}
//		}(u)
//
//		now.Sub(prev)
//		prev = now
//	}
//
//	wg.Wait()
//
//	return nil
//}
