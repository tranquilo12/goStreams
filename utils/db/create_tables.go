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

// CreateAllTablesModel Currently does a lot of things other than just create models
// #TODO Needs to be split up into multiple functions
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
			panic(err)
		}
	}(db)

	err := createSchema(db)
	if err != nil {
		panic(err)
	}

	// #TODO change the way the password is accessed.
	newUserQuery := `
					 DO
					 $$BEGIN
					 IF EXISTS (SELECT FROM pg_roles WHERE rolname = 'grafanareader') THEN
					     EXECUTE 'REVOKE ALL PRIVILEGES ON ALL TABLES IN SCHEMA public FROM grafanareader;
							      REVOKE ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public FROM grafanareader;
							      REVOKE ALL PRIVILEGES ON ALL FUNCTIONS IN SCHEMA public FROM grafanareader;
								  REVOKE USAGE ON SCHEMA public FROM grafanareader;
								  DROP USER grafanareader;';
					 END IF;
					 END$$;

					 CREATE USER grafanareader WITH PASSWORD 'Complicated@Password@Here';
					 GRANT USAGE ON SCHEMA public TO grafanareader;
					 GRANT SELECT ON public.aggregates_bars TO grafanareader;
					 GRANT SELECT ON public.ticker_vxes TO grafanareader;
					 GRANT SELECT ON public.ticker_news TO grafanareader;
					 GRANT SELECT ON public.snapshot_all_tickers TO grafanareader;
					 GRANT SELECT ON public.snapshot_gainers_losers TO grafanareader;
					 GRANT SELECT ON public.snapshot_one_tickers TO grafanareader;
					 GRANT SELECT ON public.previous_closes TO grafanareader;
					 GRANT SELECT ON public.markets TO grafanareader;
					 GRANT SELECT ON public.daily_open_closes TO grafanareader;
					 GRANT SELECT ON public.locales TO grafanareader;`
	_, err = db.Exec(newUserQuery)
	if err != nil {
		panic(err)
	}

	// Create a unique index on one of the tables.
	query1 := "create unique index if not exists aggregates_bars_t_vw_multiplier_timespan_ticker_uind on aggregates_bars (t, vw, multiplier, timespan, ticker, o, h, l, c);"
	_, err = db.Exec(query1)
	if err != nil {
		panic(err)
	}

	query2 := "create unique index if not exists ticker_vxes_ticker_market_last_updated_utc_uind on ticker_vxes(ticker, market, last_updated_utc);"
	_, err = db.Exec(query2)
	if err != nil {
		panic(err)
	}

	return nil
}
