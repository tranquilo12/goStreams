package config

import (
	"encoding/csv"
	"gopkg.in/ini.v1"
	"lightning/utils/structs"
	"log"
	"os"
)

// AggCliParams Struct for parsing cli params
type AggCliParams struct {
	Timespan   string
	From       string
	To         string
	Multiplier int
}

// RedisParams Struct for Redis db parameters.
type RedisParams struct {
	Host          string
	Port          string
	Password      string
	Db            string
	SocketTimeout string
}

// SetDBParams Function that reads the config.ini file within the directory, setting the Postgres db parameters.
// param user has options 'grafana' and 'postgres'.
func SetDBParams(params *structs.DBParams, user string) error {
	pwd, err := os.Getwd()
	config, err := ini.Load(pwd + "/config.ini")
	if err != nil {
		return err
	}

	params.Host = config.Section("DB").Key("host").String()
	//params.Port = config.Section("SSH").Key("ssh_local_bind_port").String()
	params.Port = "6432"
	params.Dbname = config.Section("DB").Key("name").String()

	if user == "postgres" {
		params.User = config.Section("DB").Key("user").String()
		params.Password = config.Section("DB").Key("password").String()
	}

	if user == "grafana" {
		params.User = config.Section("GRAFANA_POSTGRES").Key("user").String()
		params.Password = config.Section("GRAFANA_POSTGRES").Key("password").String()
	}

	return err
}

// SetPolygonCred Function that reads the config.ini file within the directory, and returns the API Key.
// param user has options 'me' and 'other'.
func SetPolygonCred(user string) string {
	pwd, err := os.Getwd()
	config, err := ini.Load(pwd + "/config.ini")
	if err != nil {
		println(err)
	}

	var appId string
	if user == "me" {
		appId = config.Section("POLYGON").Key("reverent_visvesvaraya_key").String()
	}

	if user == "other" {
		appId = config.Section("POLYGON").Key("loving_aryabhata_key").String()
	}

	return appId
}

// SetRedisCred Function that reads the config.ini file within the directory, setting the Redis db parameters.
func SetRedisCred(params *RedisParams) error {
	config, err := ini.Load("config.ini")
	if err != nil {
		return err
	}

	params.Host = config.Section("REDIS").Key("host").String()
	params.Port = config.Section("REDIS").Key("port").String()
	params.Password = config.Section("REDIS").Key("password").String()
	params.Db = config.Section("REDIS").Key("db").String()
	params.SocketTimeout = config.Section("REDIS").Key("socket_timeout").String()

	return err
}

// ReadEquitiesList Function that reads the file /utils/config/equities_list.csv, to return the list of equities.
func ReadEquitiesList() []string {
	var res []string

	pwd, err := os.Getwd()
	f, err := os.Open(pwd + "/utils/config/equities_list.csv")
	if err != nil {
		log.Fatal("Unable to read input file "+"equities_list.csv", err)
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {

		}
	}(f)

	csvReader := csv.NewReader(f)
	records, err := csvReader.ReadAll()
	if err != nil {
		log.Fatal("Unable to parse file as CSV for "+"equities_list.csv", err)
	}

	for _, element := range records[1:] {
		if len(element) == 2 {
			res = append(res, element[1])
		}
	}

	return res
}
