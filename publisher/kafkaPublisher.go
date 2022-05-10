package publisher

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/segmentio/kafka-go"
	"io/ioutil"
	"log"
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

	w := &kafka.Writer{
		Addr:         kafka.TCP("kafka-trial-shriram-c8ec.aivencloud.com:26032"),
		Topic:        topic,
		RequiredAcks: kafka.RequireAll,
		Async:        true,
		Balancer:     &kafka.Hash{},
		Transport: &kafka.Transport{
			TLS: tlsConfig,
		},
	}

	return w
}
