package publisher

import (
	"encoding/json"
	"fmt"
	"github.com/streadway/amqp"
	"go.uber.org/ratelimit"
	"net/http"
	"sync"

	"lightning/utils/structs"
	"net/url"
	"time"
)

func AggPublisherRMQ(urls []*url.URL) error {
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
		"AGG",
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
		target := new(structs.AggregatesBarsResponse)

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
