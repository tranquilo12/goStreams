package subscriber

import (
	"encoding/json"
	"fmt"
	"github.com/streadway/amqp"
	"lightning/utils/db"
	"lightning/utils/structs"
	"sync"
)

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

func TickerNewsSubscriberRMQ(DBParams *structs.DBParams) error {

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
		"NEWS",
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
			res := structs.TickerNews2Response{}
			err := json.Unmarshal(message.Body, &res)
			if err != nil {
				panic(err)
			}

			conn := pgDB.Conn()
			aggs := structs.TickerNews2FlattenPayloadBeforeInsert(res)
			if len(aggs) > 0 {
				_, err := conn.Model(&aggs).OnConflict("(id) DO NOTHING").Insert()
				if err != nil {
					panic(err)
				} else {
					fmt.Printf(" > Inserted id: %s\n", res.RequestID)
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
