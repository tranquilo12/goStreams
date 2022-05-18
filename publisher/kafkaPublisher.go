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
	"net/http"
	"net/url"
	"sync"
)

const (
	ServiceCertPath = "/Users/shriramsunder/.avn/service.cert"
	ServiceKeyPath  = "/Users/shriramsunder/.avn/service.key"
	ServiceCAPath   = "/Users/shriramsunder/.avn/ca.pem"
)

// CreateKafkaWriterConn creates a new kafka producer connection
func CreateKafkaWriterConn(topic string) *kafka.Writer {
	// Get the key and cert files
	keypair, err := tls.LoadX509KeyPair(ServiceCertPath, ServiceKeyPath)
	caCert, err := ioutil.ReadFile(ServiceCAPath)
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
	//var messages []kafka.Message

	// Download the data from PolygonIO
	var res structs.AggregatesBarsResponse
	err := DownloadFromPolygonIO(httpClient, *url, &res)
	db.Check(err)

	for _, v := range res.Results {
		// Convert the data to influx points
		_, err := json.Marshal(v)
		db.Check(err)

		// Create the messages
		//messages = append(
		//	messages,
		//	kafka.Message{
		//		Key:   []byte(res.Ticker),
		//		Value: val,
		//		Time:  time.Time{},
		//	},
		//)
	}

	// Write the messages to Kafka
	//err = kafkaWriter.WriteMessages(ctx, messages...)
	//Check(err)

	// Progress bar update
	bar.Add(1)

	// Close the wg
	wg.Done()
}
