package publisher

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"github.com/schollz/progressbar/v3"
	"github.com/segmentio/kafka-go"
	"io/ioutil"
	"lightning/utils/db"
	"lightning/utils/structs"
	"log"
	"math"
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

func KafkaWriter(
	ctx context.Context,
	httpClient *http.Client,
	url *url.URL,
	kafkaWriter *kafka.Writer,
	wg *sync.WaitGroup,
	bar *progressbar.ProgressBar,
) {
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

		// Convert float to second, and then to the time
		sec, dec := math.Modf(v.T)
		t := time.Unix(int64(sec), int64(dec*(1e9)))

		// Create the messages
		messages = append(
			messages,
			kafka.Message{
				Key:   []byte(res.Ticker),
				Value: val,
				Time:  t,
			},
		)
	}

	// Write the messages to Kafka
	err = kafkaWriter.WriteMessages(ctx, messages...)
	db.Check(err)

	// Progress bar update
	bar.Add(1)

	// Close the wg
	wg.Done()
}
