package publisher

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	qdb "github.com/questdb/go-questdb-client"
	"github.com/schollz/progressbar/v3"
	"github.com/segmentio/kafka-go"
	"go.uber.org/ratelimit"
	"io/ioutil"
	"lightning/utils/config"
	"lightning/utils/db"
	"log"
	"os"
	"path/filepath"

	//"lightning/utils/db"

	"lightning/utils/structs"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// CreateKafkaWriterConn creates a new kafka producer connection
func CreateKafkaWriterConn(topic string) *kafka.Writer {
	// Load User's home directory
	dirname, err := os.UserHomeDir()
	db.CheckErr(err)

	// Load the client cert
	serviceCertPath := filepath.Join(dirname, ".avn", "service.cert")
	serviceKeyPath := filepath.Join(dirname, ".avn", "service.key")
	serviceCAPath := filepath.Join(dirname, ".avn", "ca.pem")

	// Get the key and cert files
	keypair, err := tls.LoadX509KeyPair(serviceCertPath, serviceKeyPath)
	db.CheckErr(err)

	caCert, err := ioutil.ReadFile(serviceCAPath)
	db.CheckErr(err)

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

	// Create a writer connection
	w := &kafka.Writer{
		Addr:         kafka.TCP("kafka-trial-shriram-c8ec.aivencloud.com:26032"),
		Topic:        topic,
		RequiredAcks: kafka.RequireOne,
		Async:        true,
		Balancer:     &kafka.LeastBytes{},
		Transport: &kafka.Transport{
			TLS: tlsConfig,
		},
		Compression: kafka.Lz4,
	}

	return w
}

//// CreateKafkaWriterAws For everything related to creating a connection to aws
//func CreateKafkaWriterAws(topic string) *kafka.Writer {
//	// Load User's home directory
//	dirname, err := os.UserHomeDir()
//	db.CheckErr(err)
//
//	// Load the client cert
//	serviceCAPath := filepath.Join(dirname, ".aws", "AWSCertificate.pem")
//	caCert, err := ioutil.ReadFile(serviceCAPath)
//	db.CheckErr(err)
//
//	// Get the CA cert pool
//	caCertPool := x509.NewCertPool()
//	ok := caCertPool.AppendCertsFromPEM(caCert)
//	if !ok {
//		log.Println(err)
//	}
//
//	// Create the tls config
//	//tlsConfig := &tls.Config{
//	//	RootCAs: caCertPool,
//	//}
//
//	w := kafka.Writer{
//		Topic: topic,
//		Addr: kafka.TCP(
//			"b-3.lightningclusterfinal.9iviow.c7.kafka.us-east-2.amazonaws.com:9092",
//			"b-2.lightningclusterfinal.9iviow.c7.kafka.us-east-2.amazonaws.com:9092",
//			"b-1.lightningclusterfinal.9iviow.c7.kafka.us-east-2.amazonaws.com:9092",
//		),
//		Balancer:               &kafka.LeastBytes{},
//		AllowAutoTopicCreation: true,
//		Compression:            kafka.Lz4,
//		RequiredAcks:           kafka.RequireOne,
//		Transport:              &kafka.Transport{TLS: &tls.Config{InsecureSkipVerify: true}},
//		WriteTimeout:           10 * time.Second,
//		Async:                  true,
//	}
//	return &w
//}

// AggKafkaWriter writes the aggregates to Kafka
func AggKafkaWriter(urls []*url.URL, topic string) error {
	// use WaitGroup to make things more smooth with goroutines
	var wg sync.WaitGroup

	// create a buffer of the waitGroup, of the same length as urls
	wg.Add(len(urls))

	// Max allow 500 requests per second
	prev := time.Now()
	rateLimiter := ratelimit.New(3000)

	// Get Write client
	writeConn := CreateKafkaWriterConn(topic)
	defer writeConn.Close()

	// Create ui progress bar, formatted
	bar := progressbar.Default(int64(len(urls)))
	defer bar.Close()

	// Get the http client
	httpClient := config.GetHttpClient()

	// Iterate over every url
	for _, u := range urls {
		// rate limit the requests
		now := rateLimiter.Take()

		// Create the goroutine
		go KafkaWriter(context.Background(), httpClient, u, writeConn, &wg, bar)

		// Rate limit the requests, so note the time
		now.Sub(prev)
		prev = now
	}

	// Wait for all the goroutines to finish
	wg.Wait()

	return nil
}

func KafkaWriter(
	ctx context.Context,
	httpClient *http.Client,
	url *url.URL,
	kafkaWriter *kafka.Writer,
	wg *sync.WaitGroup,
	bar *progressbar.ProgressBar,
) {
	// Makes sure wg closes
	defer wg.Done()

	// All messages
	var messages []kafka.Message

	// Download the data from PolygonIO
	var res structs.AggregatesBarsResponse
	err := DownloadFromPolygonIO(httpClient, *url, &res)
	db.CheckErr(err)

	for _, v := range res.Results {
		// Convert the data to influx points
		val, err := json.Marshal(v)
		db.CheckErr(err)

		// Create the messages
		messages = append(
			messages,
			kafka.Message{
				Key:   []byte(res.Ticker),
				Value: val,
				Time:  time.Time{},
			},
		)
	}

	// Write the messages to Kafka
	err = kafkaWriter.WriteMessages(ctx, messages...)
	db.CheckErr(err)

	// Progress bar update
	bar.Add(1)
}

// QDBQueryAndInsertAggILP to QuestDB.
func QDBQueryAndInsertAggILP(
	ctx context.Context,
	pbar *progressbar.ProgressBar,
	timespan string,
	aggBar structs.AggregatesBarsResponse,
	multiplier int,
) error {
	// Connect to QDB and get sender
	sender, _ := qdb.NewLineSender(ctx)

	// For each of these results, push!
	for _, agg := range aggBar.Results {
		err := sender.Table("aggs").
			Symbol("ticker", aggBar.Ticker).
			StringColumn("timespan", timespan).
			Int64Column("multiplier", int64(multiplier)).
			Float64Column("open", agg.O).
			Float64Column("high", agg.H).
			Float64Column("low", agg.L).
			Float64Column("close", agg.C).
			Float64Column("volume", agg.V).
			Float64Column("vw", agg.Vw).
			Float64Column("n", float64(agg.N)).
			At(ctx, time.UnixMilli(int64(agg.T)).UnixNano())
		if err != nil {
			return err
		}
	}

	// Make sure that the messages are sent over the network.
	err := sender.Flush(ctx)
	if err != nil {
		return err
	}

	// Progress bar update
	pbar.Add(1)

	// close sender
	sender.Close()

	return nil
}

// QDBPushAllAggIntoDB Entire pipeline of querying all tickers and then pushing it to the db
func QDBPushAllAggIntoDB(ctx context.Context, urls []*url.URL, timespan string, multiplier int) {
	// Use a WaitGroup to make things simpler.
	// Create a buffer of the WaitGroup
	var wg sync.WaitGroup
	wg.Add(len(urls))

	// Done channel
	var doneCh chan bool

	// Max 300 requests per second
	prev := time.Now()
	rateLimiter := ratelimit.New(300)

	// Get the http client
	httpClient := config.GetHttpClient()

	// Init a progress pbar here
	progressbar.OptionSetWidth(500)
	//pbar := progressbar.Default(int64(len(urls)), "Downloading...")
	pbar := progressbar.NewOptions(len(urls),
		progressbar.OptionSetDescription("Downloading..."),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionSetWidth(10),
		progressbar.OptionThrottle(65*time.Millisecond),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionFullWidth(),
		progressbar.OptionSpinnerType(14),
		progressbar.OptionSetRenderBlankState(true),
		progressbar.OptionOnCompletion(func() {
			_, _ = fmt.Fprint(os.Stderr, "\n")
			doneCh <- true
		}),
	)

	// Iterate over all these urls, and insert them into the db!
	var err error
	for _, u := range urls {
		// Rate limit
		now := rateLimiter.Take()

		// Create a goroutine that will take care of the querying and insert
		go func() {
			err = Retry(10, 2, func() error {
				// First query, then insert. If anything goes wrong with this go-routine, it should start with querying it again.
				var aggBar structs.AggregatesBarsResponse
				err = DownloadFromPolygonIO(httpClient, *u, &aggBar)
				if err != nil {
					return err
				}

				err = QDBQueryAndInsertAggILP(ctx, pbar, timespan, aggBar, multiplier)
				return err
			})
		}()

		// Rate limit, recalculate
		now.Sub(prev)
		prev = now
	}

	// Wait for all of them to finish.
	wg.Wait()

	// Just close the progressbar
	pbar.Close()

	// Done
	<-doneCh
}

// Retry A "decorator" function that wraps around every func that needs a retry.
func Retry(attempts int, sleep time.Duration, f func() error) (err error) {
	for i := 0; i < attempts; i++ {
		if i > 0 {
			log.Println("Retrying after error: ", err)
			time.Sleep(sleep)
		}
		err = f()
		if err == nil {
			return nil
		}
	}
	return fmt.Errorf("after %d attempts, last error: %s", attempts, err)
}
