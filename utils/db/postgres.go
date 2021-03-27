// File that contains multiple functions that can be called by main.go
package db

import "github.com/go-pg/pg/v10"

// make sure the connection to postgres is fine
func GetPostgresDB() *pg.DB {
	var postgresDB = pg.Connect(&pg.Options{
		Network:               "",
		Addr:                  "127.0.0.1:5432",
		Dialer:                nil,
		OnConnect:             nil,
		User:                  "postgres",
		Password:              "rogerthat",
		Database:              "TimeScaleDB",
		ApplicationName:       "",
		TLSConfig:             nil,
		DialTimeout:           0,
		ReadTimeout:           0,
		WriteTimeout:          0,
		MaxRetries:            0,
		RetryStatementTimeout: false,
		MinRetryBackoff:       0,
		MaxRetryBackoff:       0,
		PoolSize:              0,
		MinIdleConns:          0,
		MaxConnAge:            0,
		PoolTimeout:           0,
		IdleTimeout:           0,
		IdleCheckFrequency:    0,
	})
	return postgresDB
}

// make sure all table models are made
func ExecCreateAllTablesModels() {
	err := CreateAllTablesModel()
	if err != nil {
		panic(err)
	}
}
