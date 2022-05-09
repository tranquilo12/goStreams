package db

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/domain"
	"lightning/utils/config"
	"lightning/utils/structs"
	"net"
	"net/http"
	"strings"
	"time"
)

// GetInfluxDBClient returns a new influx client, depending on the withOptions flag
func GetInfluxDBClient(withOptions bool) influxdb2.Client {
	params := new(structs.InfluxDBStruct)
	err := config.SetInfluxDBCred(params)
	Check(err)
	if withOptions == true {

		// Create HTTP client,
		// For efficient reuse of HTTP resources among multiple clients,
		// create an HTTP client and use Options.SetHTTPClient() for setting it to all clients:
		httpClient := &http.Client{
			Timeout: time.Second * time.Duration(600),
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout: 600 * time.Second,
				}).DialContext,
				TLSHandshakeTimeout: 600 * time.Second,
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
				MaxIdleConns:        1000,
				MaxIdleConnsPerHost: 999,
				IdleConnTimeout:     600 * time.Second,
			},
		}

		// Create client
		return influxdb2.NewClientWithOptions(
			params.Url,
			params.ApiKey,
			influxdb2.
				DefaultOptions().
				SetBatchSize(50).
				SetHTTPClient(httpClient).
				SetHTTPRequestTimeout(600),
		)
	} else {
		return influxdb2.NewClient(params.Url, params.ApiKey)
	}
}

// PushTickerVxIntoInfluxDB pushes the ticker vx into influxdb
func PushTickerVxIntoInfluxDB(insertIntoInfluxDB <-chan []structs.TickerVx, client influxdb2.Client) {
	// use WaitGroup to make things more smooth with channels
	var allTickers []string

	// for each insertIntoDB that follows...spin off another go routine
	for val, ok := <-insertIntoInfluxDB; ok; val, ok = <-insertIntoInfluxDB {
		if ok && val != nil {
			for _, v := range val {
				allTickers = append(allTickers, v.Ticker)
			}
		}
	}

	// Get write client
	writeAPI := client.WriteAPI("lightning", "Lightning")

	// Prepare a point
	p := influxdb2.NewPoint(
		"allTickers",
		map[string]string{"tag": "tickers"},
		map[string]interface{}{"tickers": allTickers},
		time.Now(),
	)

	// Write the point
	writeAPI.WritePoint(p)

	// Force all unwritten data to be written to be sent.
	writeAPI.Flush()

	// Close the client
	client.Close()
}

// GetAllTickersFromInfluxDB returns a slice of strings of all the tickers in the influxDB database
func GetAllTickersFromInfluxDB(client influxdb2.Client) []string {
	var res []string

	// Get read client
	queryAPI := client.QueryAPI("lightning")

	// Form the query
	query := `from(bucket:"Lightning") |> range(start: -30d) |> filter(fn: (r) => r._measurement == "allTickers") |> filter(fn: (r) => r._field == "tickers")`

	// Execute the query
	result, err := queryAPI.Query(context.Background(), query)
	Check(err)

	// Check for errors
	if err == nil {
		for result.Next() {
			r := result.Record().Values()["_value"].(string)
			res = strings.Split(strings.TrimSuffix(strings.TrimPrefix(r, "["), "]"), " ")
		}
		if result.Err() != nil {
			fmt.Printf("query parsing error: %s\n", result.Err().Error())
		}
	} else {
		panic(err)
	}
	return res
}

// DeleteFromInfluxDB Just delete something from the influxDB database
func DeleteFromInfluxDB(client influxdb2.Client, measurement string, from_ string, to_ string) {
	// Get client's organization ID
	orgID := domain.Organization{Id: aws.String("50f21dbeb2c94acc")}

	// Get Bucket ID
	bucketID := domain.Bucket{OrgID: orgID.Id, Name: "Lightning", Id: aws.String("3a675d3b8f306259")}

	// Get Delete client
	deleteAPI := client.DeleteAPI()

	// measurement
	m := fmt.Sprintf("_measurement=%s", measurement)

	// from and to time conversion
	layout := "2006-01-02"
	fromTime, err := time.Parse(layout, from_)
	Check(err)

	toTime, err := time.Parse(layout, to_)
	Check(err)

	// Delete the point
	err = deleteAPI.Delete(
		context.Background(),
		&orgID,
		&bucketID,
		fromTime,
		toTime,
		m,
	)
	Check(err)
}
