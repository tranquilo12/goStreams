package subscriber

import (
	"encoding/json"
	"fmt"
	"github.com/adjust/rmq/v3"
	"github.com/go-pg/pg/v10"
	"github.com/streadway/amqp"
	"lightning/utils/db"
	"lightning/utils/structs"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func NewConsumers(db *pg.DB, timespan string, multiplier int) *structs.NewConsumerStruct {
	res := structs.AggregatesBarsResponse{}
	conn := db.Conn()
	return &structs.NewConsumerStruct{
		AggBarsResponse: res,
		Timespan:        timespan,
		Multiplier:      multiplier,
		DBConn:          conn,
	}
}

func AggSubscriber(DBParams *structs.DBParams, timespan string, multiplier int) error {
	var err error
	var errChan chan error

	// get the PG connection here
	pgDB := db.GetPostgresDBConn(DBParams)

	// also get the redis connection
	client := db.GetRedisClient(7000)
	queueConnection, err := rmq.OpenConnectionWithRedisClient("AGG", client, errChan)
	if err != nil {
		fmt.Println("Something wrong with this queueConnection...")
	}

	taskQueue, err := queueConnection.OpenQueue("AGG")
	if err != nil {
		fmt.Printf("Please, something wrong with the taskQueue...")
	}

	err = taskQueue.StartConsuming(1000, time.Second)
	if err != nil {
		fmt.Printf("Please, something wrong with the StartConsuming...")
	}

	// use WaitGroup to make things more smooth with goroutines
	var wg sync.WaitGroup

	cleaner := rmq.NewCleaner(queueConnection)

	i := 0
	for range time.Tick(time.Second) {
		i += 1
		if i < 10 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				name := fmt.Sprintf("consumer %d", i)
				name, err := taskQueue.AddBatchConsumer(name, 1000, time.Second, NewConsumers(pgDB, timespan, multiplier))
				if err != nil {
					panic(err)
				}

				_, err = cleaner.Clean()
				if err != nil {
					panic(err)
				}
			}()
		} else {
			break
		}
	}

	// just wait for all of them to be done
	wg.Wait()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT)
	defer signal.Stop(signals)

	<-signals // wait for signal
	go func() {
		<-signals // hard exit on second signal (in case shutdown gets stuck)
		os.Exit(1)
	}()

	<-queueConnection.StopAllConsuming()

	return err
}

func AggSubscriberRMQ(DBParams *structs.DBParams, timespan string, multiplier int) error {

	AmqpServerUrl := "amqp://guest:guest@localhost:5672"
	connectRabbitMQ, err := amqp.Dial(AmqpServerUrl)
	if err != nil {
		//panic(err)
		return err
	}
	defer connectRabbitMQ.Close()

	channelRabbitMQ, err := connectRabbitMQ.Channel()
	if err != nil {
		//panic(err)
		return err
	}
	defer channelRabbitMQ.Close()

	messages, err := channelRabbitMQ.Consume(
		"AGG",
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		//panic(err)
		return err
	}

	// Make a channel to receive messages into infinite loop.
	forever := make(chan bool)

	// use WaitGroup to make things more smooth with goroutines
	var wg sync.WaitGroup

	// create a buffer of the waitGroup, of the same length as urls
	wg.Add(20000)

	// get the PG connection here
	pgDB := db.GetPostgresDBConn(DBParams)
	defer pgDB.Close()

	go func() {
		defer wg.Done()
		for message := range messages {

			// For example, show received message in a console.
			res := structs.AggregatesBarsResponse{}
			err := json.Unmarshal(message.Body, &res)
			if err != nil {
				panic(err)
			}

			conn := pgDB.Conn()
			aggs := structs.AggBarFlattenPayloadBeforeInsert(res, timespan, multiplier)
			if len(aggs) > 0 {
				_, err := conn.Model(&aggs).OnConflict("(t, multiplier, timespan, ticker) DO NOTHING").Insert()
				if err != nil {
					panic(err)
				} else {
					fmt.Printf(" > Inserted ticker: %s\n", res.Ticker)
				}
			}

			//if err := message.Ack(false); err != nil {
			//	log.Printf("Error acknowledging message : %s", err)
			//} else {
			//	log.Printf("Acknowledged message")
			//}

			err = conn.Close()
			if err != nil {
				panic(err)
			}
		}
	}()
	wg.Wait()

	<-forever
	return nil
}
