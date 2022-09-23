package config

import (
	"flag"
	"gopkg.in/ini.v1"
	"lightning/utils/db"
	"log"
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

func getConfigPath() string {
	wd, err := os.Getwd()
	db.CheckErr(err)
	configPath := filepath.Join(wd, "config.ini")
	return configPath
}

// SetPolygonCred Function that reads the config.ini file within the directory, and returns the API Key.
// param user has options 'me' and 'other'.
func SetPolygonCred(user string) string {
	// Get the config file path
	configPath := getConfigPath()

	// Load the config file
	config, err := ini.Load(configPath)
	db.CheckErr(err)

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
	timeout := 10 * time.Second

	dialer := &net.Dialer{
		Timeout:   timeout,
		KeepAlive: timeout,
	}

	// Create a new transport
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     timeout,
		MaxConnsPerHost:     100,
		ForceAttemptHTTP2:   true,
		DialContext:         dialer.DialContext,
	}

	// Create a new http client and return it
	return &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}
}

// MemProfiler Entire block is for profiling memory, exposing localhost:6060
func MemProfiler() {
	var memprofile = flag.String("memprofile", "", "write memory profile to this file")
	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		db.CheckErr(err)

		err = pprof.WriteHeapProfile(f)
		db.CheckErr(err)

		return
	}

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
}
