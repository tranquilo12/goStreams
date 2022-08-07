package db

import (
	"context"
	"github.com/jackc/pgx/v4"
	qdb "github.com/questdb/go-questdb-client"
	"lightning/utils/structs"
)

// QDBConnectPG creates a Postgres conn to QuestDB.
func QDBConnectPG(ctx context.Context) *pgx.Conn {
	conn, _ := pgx.Connect(ctx, "postgresql://admin:quest@localhost:8812/")
	return conn
}

// QDBConnectILP to QuestDB and return a sender.
func QDBConnectILP(ctx context.Context) (*qdb.LineSender, error) {
	sender, err := qdb.NewLineSender(ctx)
	CheckErr(err)
	return sender, err
}

// QDBInsertTickersILP to QuestDB.
// Ensure the sender is deferred closed before this function is called.
func QDBInsertTickersILP(ctx context.Context, sender *qdb.LineSender, ticker structs.TickersStruct) {
	// Send all the data within the ticker
	for _, t := range ticker.Results {
		// Push ticker
		err := sender.
			Table("tickers").
			Symbol("ticker", t.Ticker).
			StringColumn("name", t.Name).
			StringColumn("market", t.Market).
			StringColumn("locale", t.Locale).
			StringColumn("primary_exchange", t.PrimaryExchange).
			StringColumn("type", t.Type).
			BoolColumn("active", t.Active).
			StringColumn("currency_name", t.CurrencyName).
			StringColumn("cik", t.Cik).
			StringColumn("composite_figi", t.CompositeFigi).
			StringColumn("share_class_figi", t.ShareClassFigi).
			StringColumn("last_updated_utc", t.LastUpdatedUtc.String()).
			AtNow(ctx)
		CheckErr(err)
	}

	// Make sure that the messages are sent over the network.
	err := sender.Flush(ctx)
	CheckErr(err)
}

// QDBInsertAggILP to QuestDB.
func QDBInsertAggILP(ctx context.Context, sender *qdb.LineSender, aggBar structs.NewAggStruct) {
	for _, agg := range aggBar.AggBarsResponse.Results {
		// Push aggregates
		err := sender.
			Table("aggregates").
			Symbol("ticker", aggBar.AggBarsResponse.Ticker).
			Symbol("timespan", aggBar.Timespan).
			Symbol("multiplier", string(rune(aggBar.Multiplier))).
			Float64Column("timestamp", agg.T).
			Float64Column("open", agg.O).
			Float64Column("high", agg.H).
			Float64Column("low", agg.L).
			Float64Column("close", agg.C).
			Float64Column("volume", agg.V).
			Float64Column("vw", agg.Vw).
			Int64Column("n", int64(agg.N)).
			AtNow(ctx)
		CheckErr(err)
	}

	// Make sure that the messages are sent over the network.
	err := sender.Flush(ctx)
	CheckErr(err)
}
