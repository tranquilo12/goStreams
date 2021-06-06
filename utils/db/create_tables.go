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
		(*structs.TickerNews2)(nil),
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
	query1 := "CREATE UNIQUE INDEX IF NOT EXISTS aggregates_bars_t_multiplier_timespan_ticker_uind on aggregates_bars (t, multiplier, timespan, ticker);"
	_, err = db.Exec(query1)
	if err != nil {
		panic(err)
	}

	query2 := "CREATE UNIQUE INDEX IF NOT EXISTS ticker_vxes_ticker_market_last_updated_utc_uind on ticker_vxes(ticker, market, last_updated_utc);"
	_, err = db.Exec(query2)
	if err != nil {
		panic(err)
	}

	query3 := `CREATE TABLE IF NOT EXISTS wsb_comments
			(
				total_awards_received           int,
				approved_at_utc                 int,
				comment_type                    int,
				awarders                        jsonb,
				mod_reason_by                   text,
				banned_by                       text,
				ups                             int,
				author_flair_type               text,
				removal_reason                  text,
				link_id                         text,
				author_flair_template_id        text,
				likes                           int,
				user_reports                    jsonb,
				saved                           bool,
				id                              text,
				banned_at_utc                   int,
				mod_reason_title                text,
				gilded                          int,
				archived                        bool,
				no_follow                       bool,
				author                          text,
				can_mod_post                    bool,
				send_replies                    bool,
				parent_id                       text,
				score                           int,
				author_fullname                 text,
				report_reasons                  jsonb,
				approved_by                     text,
				all_awardings                   jsonb,
				subreddit_id                    text,
				body                            text,
				edited                          int,
				downs                           int,
				author_flair_css_class          text,
				is_submitter                    bool,
				collapsed                       bool,
				author_flair_richtext           jsonb,
				author_patreon_flair            text,
				body_html                       text,
				gildings                        jsonb,
				collapsed_reason                text,
				associated_award                text,
				stickied                        bool,
				author_premium                  bool,
				subreddit_type                  text,
				can_gild                        bool,
				top_awarded_type                text,
				author_flair_text_color         text,
				score_hidden                    bool,
				permalink                       text,
				num_reports                     int,
				locked                          bool,
				name                            text,
				created                         int,
				author_flair_text               text,
				treatment_tags                  jsonb,
				created_utc                     int,
				subreddit_name_prefixed         text,
				controversiality                int,
				depth                           int,
				author_flair_background_color   text,
				collapsed_because_crowd_control bool,
				mod_reports                     jsonb,
				mod_note                        text,
				distinguished                   text
	);`
	_, err = db.Exec(query3)
	if err != nil {
		panic(err)
	}

	query4 := `CREATE UNIQUE INDEX IF NOT EXISTS wsb_comments_id_uind on wsb_comments(id);`
	_, err = db.Exec(query4)
	if err != nil {
		panic(err)
	}
	return nil
}
