package publisher

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"github.com/schollz/progressbar/v3"
	"github.com/segmentio/kafka-go"
	"go.uber.org/ratelimit"
	"io/ioutil"
	"lightning/utils/config"
	"lightning/utils/db"
	"lightning/utils/structs"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// CreateKafkaWriterConn creates a new kafka producer connection
func CreateKafkaWriterConn(topic string) *kafka.Writer {
	// Load User's home directory
	dirname, err := os.UserHomeDir()
	db.Check(err)

	// Load the client cert
	serviceCertPath := filepath.Join(dirname, ".avn", "service.cert")
	serviceKeyPath := filepath.Join(dirname, ".avn", "service.key")
	serviceCAPath := filepath.Join(dirname, ".avn", "ca.pem")

	// Get the key and cert files
	keypair, err := tls.LoadX509KeyPair(serviceCertPath, serviceKeyPath)
	db.Check(err)

	caCert, err := ioutil.ReadFile(serviceCAPath)
	db.Check(err)

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

// AggKafkaWriter writes the aggregates to Kafka
func AggKafkaWriter(urls []*url.URL) error {
	// use WaitGroup to make things more smooth with goroutines
	var wg sync.WaitGroup

	// create a buffer of the waitGroup, of the same length as urls
	wg.Add(len(urls))

	// Max allow 500 requests per second
	prev := time.Now()
	rateLimiter := ratelimit.New(300)

	// Get Write client
	writeConn := CreateKafkaWriterConn("agg")
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
	db.Check(err)

	for _, v := range res.Results {
		// Convert the data to influx points
		val, err := json.Marshal(v)
		db.Check(err)

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
	db.Check(err)

	// Progress bar update
	bar.Add(1)
}
