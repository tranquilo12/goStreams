package db

import (
	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
	"lightning/utils/structs"
)

// createSchema creates database schema for User and Story models.
func createSchema(db *pg.DB) error {
	models := []interface{}{
		(*structs.DailyOpenClose)(nil),
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

func CreateAllTablesModel() error {

	db := pg.Connect(&pg.Options{
		Addr:     "127.0.0.1:5432",
		User:     "postgres",
		Password: "rogerthat",
		Database: "TimeScaleDB",
	})
	defer db.Close()

	err := createSchema(db)
	if err != nil {
		panic(err)
	}

	return nil
}
