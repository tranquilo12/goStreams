package subscriber

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/segmentio/kafka-go"
	_ "github.com/segmentio/kafka-go/snappy"
	"lightning/utils/db"
	"log"
	"os"
	"path/filepath"
	"time"
)

const (
	batchSize = int(10e6) // 10MB
)

// CreateKafkaReaderConn creates a new kafka subscriber connection
func CreateKafkaReaderConn(topic string) *kafka.Reader {
	// Load User's home directory
	dirname, err := os.UserHomeDir()
	db.CheckErr(err)

	// Load the client cert
	serviceCertPath := filepath.Join(dirname, ".avn", "service.cert")
	serviceKeyPath := filepath.Join(dirname, ".avn", "service.key")
	serviceCAPath := filepath.Join(dirname, ".avn", "ca.pem")

	// Get the key and cert files
	keypair, err := tls.LoadX509KeyPair(serviceCertPath, serviceKeyPath)
	caCert, err := os.ReadFile(serviceCAPath)
	if err != nil {
		log.Println(err)
	}

	// Get the CA cert pool
	caCertPool := x509.NewCertPool()
	ok := caCertPool.AppendCertsFromPEM(caCert)
	if !ok {
		log.Println(err)
	}

	// Create the tls config
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{keypair},
		RootCAs:      caCertPool,
	}

	// Get a new dialer
	dialer := &kafka.Dialer{
		Timeout:   90 * time.Second,
		TLS:       tlsConfig,
		DualStack: true,
	}

	// Create the reader
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        []string{"kafka-trial-shriram-c8ec.aivencloud.com:26032"},
		Topic:          topic,
		Dialer:         dialer,
		MinBytes:       batchSize * 10,  // 500MB
		MaxBytes:       batchSize * 100, // 1000MB
		CommitInterval: time.Second * 10,
	})
	return r
}

//
//// WriteFromKafkaToQuestDB writes data from kafka to questdb
//func WriteFromKafkaToQuestDB(topic string, urls []string) {
//	// Create a new context
//	ctx := context.TODO()
//
//	// Create a new questdb client
//	readerConns := CreateKafkaReaderConn(topic)
//	defer readerConns.Close()
//
//	// Create new progress bar
//	progLen := int64(len(urls))
//	bar := progressbar.Default(-1)
//
//	// Create a new channel
//	ch := make(chan structs.NewAggStruct, progLen)
//	defer close(ch)
//
//	// Create waitGroup for goroutines, buffered
//	var wg sync.WaitGroup
//	wg.Add(int(progLen))
//
//	// Connect to QDB and get sender
//	sender, _ := qdb.NewLineSender(ctx)
//	defer sender.Close()
//
//	// Create a new go routine to insert data into the channel
//	go func() {
//		for {
//			// Get the message
//			m, err := readerConns.ReadMessage(ctx)
//			if err != nil {
//				log.Println("Error reading message: ", err)
//				break
//			}
//
//			// Get the data from the message
//			v := structs.AggregatesBarsResults{}
//			err = json.Unmarshal(m.Value, &v)
//			db.CheckErr(err)
//
//			// Add the message to the channel
//			ch <- structs.NewAggStruct{Ticker: string(m.Key), AggBarsResults: v}
//		}
//	}()
//
//	// Create a new go routine
//	go func(wg1 *sync.WaitGroup) {
//		for {
//			// Complete the waitGroup
//			defer wg1.Done()
//
//			// Get the message from the channel
//			data := <-ch
//
//			// Send the data to QDB
//			err := sender.Table("aggs").
//				Symbol("ticker", data.Ticker).
//				StringColumn("timespan", "minute").
//				Int64Column("multiplier", int64(1)).
//				Float64Column("open", data.AggBarsResults.O).
//				Float64Column("high", data.AggBarsResults.H).
//				Float64Column("low", data.AggBarsResults.L).
//				Float64Column("close", data.AggBarsResults.C).
//				Float64Column("volume", data.AggBarsResults.V).
//				Float64Column("vw", data.AggBarsResults.Vw).
//				Float64Column("n", float64(data.AggBarsResults.N)).
//				At(ctx, time.UnixMilli(int64(data.AggBarsResults.T)).UnixNano())
//			if err != nil {
//				panic(err)
//			}
//
//			// Make sure the sender is flushed
//			sender.Flush(ctx)
//
//			// Increment the progress bar
//			bar.Add(1)
//		}
//	}(&wg)
//
//	// Close the channel
//	//close(ch)
//
//	// Wait for the waitGroup to finish
//	wg.Wait()
//
//	// Wait for the progress bar to finish
//	_ = bar.Finish()
//	bar.Close()
//}
