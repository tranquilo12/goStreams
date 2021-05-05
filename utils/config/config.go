package config

import (
	"encoding/csv"
	"gopkg.in/ini.v1"
	"log"
	"os"
)

type DbParams struct {
	Host     string
	Port     string
	User     string
	Password string
	Dbname   string
}

type RedisParams struct {
	Host          string
	Port          string
	Password      string
	Db            string
	SocketTimeout string
}

func SetDBParams(params *DbParams, user string) error {
	pwd, err := os.Getwd()
	config, err := ini.Load(pwd + "/config.ini")
	if err != nil {
		return err
	}

	params.Host = config.Section("DB").Key("host").String()
	params.Port = config.Section("SSH").Key("ssh_local_bind_port").String()
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
