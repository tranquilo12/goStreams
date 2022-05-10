package subscriber

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/schollz/progressbar/v3"
	"github.com/segmentio/kafka-go"
	"io/ioutil"
	"lightning/publisher"
	"lightning/utils/db"
	"lightning/utils/structs"
	"log"
	"time"
)

const (
	batchSize = int(10e6) // 10MB
)

// CreateKafkaReaderConn creates a new kafka subscriber connection
func CreateKafkaReaderConn(topic string, groupID string) *kafka.Reader {

	// Get the key and cert files
	keypair, err := tls.LoadX509KeyPair(publisher.ServiceCertPath, publisher.ServiceKeyPath)
	caCert, err := ioutil.ReadFile(publisher.ServiceCAPath)
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

	// Create the readers
	//var readers []*kafka.Reader
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{"kafka-trial-shriram-c8ec.aivencloud.com:26032"},
		GroupID:  groupID,
		Topic:    topic,
		Dialer:   dialer,
		MinBytes: 1e6,
		MaxBytes: batchSize,
	})
	return r
}

// WriteFromKafkaToInfluxDB writes the data from kafka to influxdb
func WriteFromKafkaToInfluxDB(kafkaReaders *kafka.Reader, influxDBClient influxdb2.Client) {
	// Get Write influxDBClient
	writeAPI := influxDBClient.WriteAPI("lightning", "Lightning")
	defer influxDBClient.Close()

	// Get a progress bar
	bar := progressbar.Default(-1)

	for {
		// Get the message
		m, err := kafkaReaders.ReadMessage(context.Background())
		db.Check(err)

		// Get the ticker from the message
		ticker := string(m.Key)

		// Get the data from the message
		v := structs.AggregatesBarsResults{}
		err = json.Unmarshal(m.Value, &v)
		db.Check(err)

		// Convert the data to influx points
		p := influxdb2.NewPoint(
			"aggregates",
			map[string]string{"ticker": ticker},
			map[string]interface{}{
				"open": v.O, "high": v.H, "low": v.L, "close": v.C, "vWap": v.Vw, "volume": v.V,
			},
			time.Unix(int64(v.T)/1000, 0),
		)

		// Write messages to influxdb
		writeAPI.WritePoint(p)

		// Flush write API
		writeAPI.Flush()

		// Update the progress bar
		err = bar.Add(1)
		db.Check(err)
	}

}
