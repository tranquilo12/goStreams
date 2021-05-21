package subscriber

import (
	"fmt"
	"github.com/adjust/rmq/v3"
	"github.com/go-pg/pg/v10"
	"lightning/utils/db"
	"lightning/utils/structs"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

//func NewBatchConsumers(db *pg.DB, timespan string, multiplier int) *structs.NewConsumerStruct {
//	res := structs.AggregatesBarsResponse{}
//	conn := db.Conn()
//	return &structs.NewConsumerStruct{
//		AggBarsResponse: res,
//		Timespan:        timespan,
//		Multiplier:      multiplier,
//		DBConn:          conn,
//	}
//}

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

	// create a buffer of the waitGroup, of the same length as urls
	//wg.Add(1000000)

	cleaner := rmq.NewCleaner(queueConnection)

	i := 0
	for range time.Tick(time.Second) {
		i += 1
		wg.Add(1)
		go func() {
			defer wg.Done()
			name := fmt.Sprintf("consumer %d", i)

			name, err := taskQueue.AddBatchConsumer(name, 100, time.Second, NewConsumers(pgDB, timespan, multiplier))
			if err != nil {
				panic(err)
			}

			//if _, err := taskQueue.AddConsumer(name, NewConsumers(pgDB, timespan, multiplier)); err != nil {
			//	panic(err)
			//}

			_, err = cleaner.Clean()
			if err != nil {
				panic(err)
			}

		}()
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
