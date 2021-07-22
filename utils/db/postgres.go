package db

import (
	"fmt"
	"github.com/go-pg/pg/v10"
	"github.com/spf13/cobra"
	"lightning/utils/config"
	"lightning/utils/structs"
	"time"
)

const dateFmt = "2006-01-02"

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

func ReadAggregateParamsFromCMD(cmd *cobra.Command) config.AggCliParams {

	timespan, _ := cmd.Flags().GetString("timespan")
	if timespan == "" {
		panic("Cmon provide some context, which --timespan??")
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

	limit, _ := cmd.Flags().GetInt("limit")
	if limit < 0 {
		limit = 100
	}

	useRedis, _ := cmd.Flags().GetInt("useRedis")
	adjusted, _ := cmd.Flags().GetInt("adjusted")
	withLinearDates, _ := cmd.Flags().GetInt("withLinearDates")

	forceInsertDate, _ := cmd.Flags().GetString("forceInsertDate")
	if forceInsertDate == "" {
		currDate := time.Now()
		forceInsertDate = currDate.Format(dateFmt)
	}

	res := config.AggCliParams{
		Timespan:        timespan,
		From:            from_,
		To:              to_,
		Multiplier:      multiplier,
		Limit:           limit,
		WithLinearDates: withLinearDates,
		ForceInsertDate: forceInsertDate,
		UseRedis:        useRedis,
		Adjusted:        adjusted,
	}

	return res
}

// ReadTickerNewsParamsFromCMD reads parameters like ticker, startDate, endDate
func ReadTickerNewsParamsFromCMD(cmd *cobra.Command) config.NewsCliParams {
	ticker, _ := cmd.Flags().GetStringSlice("ticker")
	if len(ticker) == 0 {
		panic("Cmon provide some context, which ticker(s)??")
	}

	from_, _ := cmd.Flags().GetString("from")
	if from_ == "" {
		from_ = "2021-01-01"
	}

	to_, _ := cmd.Flags().GetString("to")
	if to_ == "" {
		to_ = "2021-03-01"
	}

	res := config.NewsCliParams{
		Tickers: ticker,
		From:    from_,
		To:      to_,
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
func ExecCreateAllTablesModels(DBParams *structs.DBParams) {
	err := CreateAllTablesModel(DBParams.User, DBParams.Password, DBParams.Dbname, DBParams.Host, DBParams.Port)
	if err != nil {
		panic(err)
	}
}
