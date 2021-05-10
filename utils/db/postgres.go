package db

import (
	"fmt"
	"github.com/go-pg/pg/v10"
	"github.com/spf13/cobra"
	"lightning/utils/config"
	"lightning/utils/structs"
)

// ReadPostgresDBParamsFromCMD A function that reads in parameters related to the postgres DB.
func ReadPostgresDBParamsFromCMD(cmd *cobra.Command) structs.DBParams {
	user, _ := cmd.Flags().GetString("user")
	if user == "" {
		user = "postgres"
	}

	password, _ := cmd.Flags().GetString("password")
	if password == "" {
		panic("Cmon, pass a password!")
	}

	dbname, _ := cmd.Flags().GetString("database")
	if dbname == "" {
		dbname = "postgres"
	}

	host, _ := cmd.Flags().GetString("host")
	if host == "" {
		host = "127.0.0.1"
	}

	port, _ := cmd.Flags().GetString("port")
	if port == "" {
		port = "5432"
	}

	res := structs.DBParams{
		User:     user,
		Password: password,
		Dbname:   dbname,
		Host:     host,
		Port:     port,
	}
	return res
}

func GetAggParams(cmd *cobra.Command) config.AggCliParams {

	timespan, _ := cmd.Flags().GetString("timespan")
	if timespan == "" {
		timespan = "min"
	}

	from_, _ := cmd.Flags().GetString("from")
	if from_ == "" {
		from_ = "2021-01-01"
	}

	to_, _ := cmd.Flags().GetString("to")
	if to_ == "" {
		to_ = "2021-03-01"
	}

	// make multiplier 1 always
	multiplier, _ := cmd.Flags().GetInt("mult")
	if multiplier == 2 {
		multiplier = 1
	}

	res := config.AggCliParams{
		Timespan:   timespan,
		From:       from_,
		To:         to_,
		Multiplier: multiplier,
	}

	return res
}

// GetPostgresDBConn Makes sure the connection object to the postgres instance is returned.
func GetPostgresDBConn(DBParams *structs.DBParams) *pg.DB {
	addr := fmt.Sprintf("%s:%s", DBParams.Host, DBParams.Port)
	var postgresDB = pg.Connect(&pg.Options{
		Addr:     addr,
		User:     DBParams.User,
		Password: DBParams.Password,
		Database: DBParams.Dbname,
		PoolSize: 100,
	})
	return postgresDB
}

// ExecCreateAllTablesModels Makes sure CreateAllTablesModels() is called and all table models are made.
func ExecCreateAllTablesModels(user string, password string, database string, host string, port string) {
	err := CreateAllTablesModel(user, password, database, host, port)
	if err != nil {
		panic(err)
	}
}
