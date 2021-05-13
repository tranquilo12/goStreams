package subscriber

import (
	"fmt"
	"github.com/adjust/rmq/v3"
	"github.com/go-pg/pg/v10"
	"lightning/utils/db"
	"lightning/utils/structs"
	"os"
	"os/signal"
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

	err = taskQueue.StartConsuming(10, time.Second)
	if err != nil {
		fmt.Printf("Please, something wrong with the StartConsuming...")
	}

	for i := 0; i < 10; i++ {
		name := fmt.Sprintf("consumer %d", i)
		if _, err := taskQueue.AddConsumer(name, NewConsumers(pgDB, timespan, multiplier)); err != nil {
			panic(err)
		}
	}

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
