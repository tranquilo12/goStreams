package config

import (
	"gopkg.in/ini.v1"
	"lightning/utils/structs"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

//const configPath = "/Users/shriramsunder/GolandProjects/goStreams/config.ini"

// AggCliParams Struct for parsing cli params
type AggCliParams struct {
	Timespan        string
	From            string
	To              string
	Multiplier      int
	Limit           int
	WithLinearDates int
	ForceInsertDate string
	UseRedis        int
	Adjusted        int
	Gap             int
}

// NewsCliParams Struct for parsing cli params
type NewsCliParams struct {
	Tickers []string
	From    string
	To      string
}

// RedisParams Struct for Redis db parameters.
type RedisParams struct {
	Host          string
	Port          string
	Password      string
	Db            string
	SocketTimeout string
}

func Check(err error) {
	if err != nil {
		panic(err)
	}
}

func getConfigPath() string {
	wd, err := os.Getwd()
	Check(err)
	configPath := filepath.Join(wd, "config.ini")
	return configPath
}

// SetDBParams Function that reads the config.ini file within the directory, setting the Postgres db parameters.
// param user has options 'grafana' and 'postgres'.
func SetDBParams(params *structs.DBParams, section string) error {
	configPath := getConfigPath()
	config, err := ini.Load(configPath)
	Check(err)

	section = strings.ToUpper(section)
	params.Host = config.Section(section).Key("host").String()
	params.Port = config.Section(section).Key("port").String()
	params.Dbname = config.Section(section).Key("name").String()
	params.User = config.Section(section).Key("user").String()
	params.Password = config.Section(section).Key("password").String()

	return err
}

// GetElasticCacheEndpoint Function that reads the config.ini file within the directory, setting the Elastic cache endpoint.
func GetElasticCacheEndpoint(section string) string {
	configPath := getConfigPath()
	config, err := ini.Load(configPath)
	Check(err)

	section = strings.ToUpper(section)
	endpoint := config.Section(section).Key("primaryEndpoint").String()
	return endpoint
}

// SetPolygonCred Function that reads the config.ini file within the directory, and returns the API Key.
// param user has options 'me' and 'other'.
func SetPolygonCred(user string) string {
	// Get the config file path
	configPath := getConfigPath()

	// Load the config file
	config, err := ini.Load(configPath)
	Check(err)

	// Get the API Key depending upon the username
	var appId string
	switch user {
	case "me":
		appId = config.Section("POLYGON").Key("reverent_visvesvaraya_key").String()
	case "other":
		appId = config.Section("POLYGON").Key("loving_aryabhata_key").String()
	default:
		appId = config.Section("POLYGON").Key("reverent_visvesvaraya_key").String()
	}
	return appId
}

// SetRedisCred Function that reads the config.ini file within the directory, setting the Redis db parameters.
func SetRedisCred(params *RedisParams) error {
	configPath := getConfigPath()
	config, err := ini.Load(configPath)
	Check(err)

	params.Host = config.Section("REDIS").Key("host").String()
	params.Port = config.Section("REDIS").Key("port").String()
	params.Password = config.Section("REDIS").Key("password").String()
	params.Db = config.Section("REDIS").Key("db").String()
	params.SocketTimeout = config.Section("REDIS").Key("socket_timeout").String()
	return err
}

// SetInfluxDBCred Function that reads the config.ini file within the directory, setting the Influx db parameters.
func SetInfluxDBCred(params *structs.InfluxDBStruct) error {
	configPath := getConfigPath()
	config, err := ini.Load(configPath)
	Check(err)

	params.Url = config.Section("INFLUXDB").Key("url").String()
	params.Bucket = config.Section("INFLUXDB").Key("bucket").String()
	params.Org = config.Section("INFLUXDB").Key("org").String()
	params.ApiKey = config.Section("INFLUXDB").Key("apikey2").String()
	params.PersonalApiKey = config.Section("INFLUXDB").Key("personalApiKey2").String()

	return err
}

// GetHttpClient Get a modified http client with the correct timeout.
func GetHttpClient() *http.Client {
	timeout := time.Duration(900 * time.Second)

	dialer := &net.Dialer{
		Timeout:   timeout,
		KeepAlive: timeout,
	}

	// Create a new transport
	transport := &http.Transport{
		MaxIdleConns:        1000,
		MaxIdleConnsPerHost: 1000,
		IdleConnTimeout:     timeout,
		MaxConnsPerHost:     1000,
		ForceAttemptHTTP2:   true,
		DialContext:         dialer.DialContext,
	}

	// Create a new http client and return it
	return &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}
}
