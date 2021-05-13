package subscriber

import (
	"fmt"
	"github.com/adjust/rmq/v3"
	"lightning/utils/db"
	"lightning/utils/structs"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func NewConsumers() *structs.AggregatesBarsResponse {
	return &structs.AggregatesBarsResponse{}
}

func AggSubscriber() error {
	var err error
	var errChan chan error

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
		if _, err := taskQueue.AddConsumer(name, NewConsumers()); err != nil {
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
