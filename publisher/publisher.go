package publisher

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	qdb "github.com/questdb/go-questdb-client"
	"github.com/schollz/progressbar/v3"
	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
	"go.uber.org/ratelimit"
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

// DownloadFromPolygonIO downloads the prices from PolygonIO
func DownloadFromPolygonIO(
	client *http.Client,
	logger *logrus.Logger,
	u url.URL,
	res *structs.AggregatesBarsResponse,
) error {
	// Create a new client
	resp, err := client.Get(u.String())
	if err != nil {
		logger.WithFields(logrus.Fields{
			"url":   u.String(),
			"error": err,
		}).Info("Error in downloading data from PolygonIO")
		return nil
	}

	// Defer the closing of the body
	defer resp.Body.Close()

	// Decode the response
	if resp.StatusCode == http.StatusOK {
		err = json.NewDecoder(resp.Body).Decode(&res)
	}
	return err
}

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

	caCert, err := os.ReadFile(serviceCAPath)
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

// AggChannelWriter writes the aggregates to Kafka
func AggChannelWriter(
	urls []string,
	topic string,
	memProfile bool,
	logger *logrus.Logger,
) error {
	// use WaitGroup to make things more smooth with goroutines
	var wg sync.WaitGroup

	// create a buffer of the waitGroup, of the same length as urls
	wg.Add(len(urls))

	// Get Write client
	//writeConn := CreateKafkaWriterConn(topic)
	//defer writeConn.Close()

	// Create ui progress bar, formatted
	bar := progressbar.Default(int64(len(urls)))
	defer bar.Close()

	// Get the http client
	httpClient := config.GetHttpClient()

	// Get the channel that will be used to write to QuestDB
	c := make(chan structs.AggregatesBarsResponse, len(urls))
	defer close(c)

	// Make a goroutine that will accept data from a channel and push to questDB
	ctx := context.Background()
	wg.Add(1)
	go func() {
		defer wg.Done()

		sender, _ := qdb.NewLineSender(ctx)
		defer sender.Close()

		// Get the values from the channel
		for res := range c {
			for _, v := range res.Results {
				err := sender.Table("aggs").
					Symbol("ticker", res.Ticker).
					StringColumn("timespan", "minute").
					Int64Column("multiplier", int64(1)).
					Float64Column("open", v.O).
					Float64Column("high", v.H).
					Float64Column("low", v.L).
					Float64Column("close", v.C).
					Float64Column("volume", v.V).
					Float64Column("vw", v.Vw).
					Float64Column("n", float64(v.N)).
					At(ctx, time.UnixMilli(int64(v.T)).UnixNano())
				db.CheckErr(err)
			}

			// Make sure the sender is flushed
			err := sender.Flush(ctx)
			db.CheckErr(err)
		}
	}()

	// Max allow 500 requests per second
	prev := time.Now()
	rateLimiter := ratelimit.New(300)

	// Iterate over every url
	for _, u := range urls {
		// rate limit the requests
		now := rateLimiter.Take()

		// Create the goroutine
		go channelWriter(c, logger, httpClient, u, &wg, bar)

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
	logger *logrus.Logger,
	httpClient *http.Client,
	u string,
	kafkaWriter *kafka.Writer,
	wg *sync.WaitGroup,
	bar *progressbar.ProgressBar,
	memProfile bool,
) {
	// Makes sure wg closes
	defer wg.Done()

	// All messages
	var messages []kafka.Message

	// Convert the u(string) to a *url.URL
	FinalUrl, err := url.Parse(u)
	db.CheckErr(err)

	// Download the data from PolygonIO
	var res structs.AggregatesBarsResponse
	err = DownloadFromPolygonIO(httpClient, logger, *FinalUrl, &res)
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

	// Write the messages to Kafka, if memProfile is false
	if !memProfile {
		err = kafkaWriter.WriteMessages(ctx, messages...)
		db.CheckErr(err)
	}

	// Progress bar update
	bar.Add(1)

}

func channelWriter(
	chan1 chan structs.AggregatesBarsResponse,
	logger *logrus.Logger,
	httpClient *http.Client,
	u string,
	wg *sync.WaitGroup,
	bar *progressbar.ProgressBar,
) {
	// Makes sure wg closes
	defer wg.Done()

	// Convert the u(string) to a *url.URL
	FinalUrl, err := url.Parse(u)
	db.CheckErr(err)

	// Download the data from PolygonIO
	var res structs.AggregatesBarsResponse
	err = DownloadFromPolygonIO(httpClient, logger, *FinalUrl, &res)
	db.CheckErr(err)

	// Send the data to QDB, if response is not empty
	if res.Results != nil {
		chan1 <- res
	}

	// Progress bar update
	err = bar.Add(1)
	db.CheckErr(err)
}
