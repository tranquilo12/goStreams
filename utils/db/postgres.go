package db

import (
	"fmt"
	"github.com/go-pg/pg/v10"
)

// GetPostgresDB Makes sure the connection object to the postgres instance is returned.
func GetPostgresDB(user string, password string, database string, host string, port string) *pg.DB {
	addr := fmt.Sprintf("%s:%s", host, port)
	var postgresDB = pg.Connect(&pg.Options{
		Addr:     addr,
		User:     user,
		Password: password,
		Database: database,
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
