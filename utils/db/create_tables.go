package db

import (
	"fmt"
	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
	"lightning/utils/structs"
)

// createSchema creates database schema for User and Story models.
func createSchema(db *pg.DB) error {
	models := []interface{}{
		(*structs.DailyOpenClose)(nil),
		(*structs.Tickers)(nil),
		(*structs.TickerVx)(nil), // rename in db to Tickers_VX
		(*structs.TickerType)(nil),
		(*structs.TickerDetails)(nil),
		(*structs.TickerNews)(nil),
		(*structs.Markets)(nil),
		(*structs.Locales)(nil),
		(*structs.StockSplits)(nil),
		(*structs.StockDividends)(nil),
		(*structs.StockFinancials)(nil),
		(*structs.MarketHolidays)(nil),
		(*structs.MarketStatus)(nil),
		(*structs.StockExchanges)(nil),
		(*structs.ConditionsMapping)(nil),
		(*structs.CryptoExchanges)(nil),
		(*structs.AggregatesBars)(nil),
		(*structs.GroupedDailyBars)(nil),
		(*structs.PreviousClose)(nil),
		(*structs.SnapshotAllTickers)(nil),
		(*structs.SnapshotOneTicker)(nil),
		(*structs.SnapshotGainersLosers)(nil),
	}

	for _, model := range models {
		err := db.Model(model).CreateTable(&orm.CreateTableOptions{
			Temp:        false,
			IfNotExists: true,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func CreateAllTablesModel(user string, password string, database string, host string, port string) error {

	addr := fmt.Sprintf("%s:%s", host, port)
	db := pg.Connect(&pg.Options{
		Addr:     addr,
		User:     user,
		Password: password,
		Database: database,
	})
	defer func(db *pg.DB) {
		err := db.Close()
		if err != nil {

		}
	}(db)

	err := createSchema(db)
	if err != nil {
		panic(err)
	}

	return nil
}
