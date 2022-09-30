package config

import (
	"context"
	"flag"
	log "github.com/sirupsen/logrus"
	"gopkg.in/ini.v1"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime/pprof"
	"time"
)

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

// getConfigPath Get the config file path
func getConfigPath() string {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	configPath := filepath.Join(wd, "config.ini")
	return configPath
}

// getLogfilePath Get the log file path
func getLogfilePath() string {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	logfilePath := filepath.Join(wd, "logs", "log.txt")
	return logfilePath
}

// GetLogger Get the logger
func GetLogger() *log.Logger {
	// Create a new instance of the logger and ensure fields.
	var logger = log.New()
	logger.SetFormatter(&log.JSONFormatter{})
	logger.WithFields(log.Fields{
		"app": "lightning",
		"url": "",
		"err": "",
	})

	// Make sure to create the log file, or append to it
	logfilePath := getLogfilePath()
	logfile, err := os.OpenFile(logfilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}

	// Set the output to the log file
	logger.SetOutput(logfile)

	return logger
}

// SetPolygonCred Function that reads the config.ini file within the directory, and returns the API Key.
// param user has options 'me' and 'other'.
func SetPolygonCred(user string) string {
	// Get the config file path
	configPath := getConfigPath()

	// Load the config file
	config, err := ini.Load(configPath)
	if err != nil {
		panic(err)
	}

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

// GetHttpClient Get a modified http client with the correct timeout.
func GetHttpClient() *http.Client {
	timeout := 120 * time.Second

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
		ForceAttemptHTTP2:   false,
		DialContext:         dialer.DialContext,
		DisableKeepAlives:   true,
	}

	// Create a new http client and return it
	return &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}
}

// MemProfiler Entire block is for profiling memory, exposing localhost:6060
func MemProfiler(ctx context.Context) {
	var memprofile = flag.String("memprofile", "", "write memory profile to this file")
	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			panic(err)
		}

		err = pprof.WriteHeapProfile(f)
		if err != nil {
			panic(err)
		}

		pprof.SetGoroutineLabels(ctx)
		if err != nil {
			panic(err)
		}

		return
	}

	// Start the fgprof server
	//http.DefaultServeMux.Handle("/debug/fgprof", fgprof.Handler())

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
}
