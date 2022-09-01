package db

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4"
	qdb "github.com/questdb/go-questdb-client"
	"lightning/utils/structs"
	"time"
)

// QDBConnectPG creates a Postgres conn to QuestDB.
func QDBConnectPG(ctx context.Context) *pgx.Conn {
	conn, _ := pgx.Connect(ctx, "postgresql://admin:quest@localhost:8812/")
	return conn
}

// QDBInsertTickersILP to QuestDB.
// Ensure the sender is deferred closed before this function is called.
func QDBInsertTickersILP(ctx context.Context, ticker structs.TickersStruct) {
	// Connect to QDB and get sender
	sender, _ := qdb.NewLineSender(ctx)

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
			At(ctx, time.Now().UnixNano())
		CheckErr(err)
	}

	// Make sure that the messages are sent over the network.
	err := sender.Flush(ctx)
	CheckErr(err)

	// Close the sender here
	sender.Close()
}

// QDBCreateAggTable For just creating the base agg table
func QDBCreateAggTable(ctx context.Context) {
	println("-	Creating Aggregates table and constraints...")
	conn := QDBConnectPG(ctx)
	defer conn.Close(ctx)

	tx, err := conn.BeginTx(ctx, pgx.TxOptions{})
	CheckErr(err)

	// Create a table, using text-based query
	_, err = tx.Exec(ctx,
		"CREATE TABLE IF NOT EXISTS aggregates ("+
			"ticker SYMBOL, timespan SYMBOL, multiplier SYMBOL, timestamp TIMESTAMP, "+
			"open DOUBLE PRECISION, high DOUBLE PRECISION, low DOUBLE PRECISION, close DOUBLE PRECISION, volume DOUBLE PRECISION,"+
			"vw DOUBLE PRECISION, n INT), "+
			"index(ticker) timestamp(timestamp);",
	)
	CheckErr(err)

	if err := tx.Commit(ctx); err != nil {
		fmt.Printf("Failed to commit: %v\n", err)
	}

	//// Create the constraint here.
	//_, err = tx.Exec(ctx, "ALTER TABLE aggregates ALTER COLUMN ticker ADD INDEX;")
	//CheckErr(err)
	//
	//if err := tx.Commit(ctx); err != nil {
	//	fmt.Printf("Failed to commit: %v\n", err)
	//}

	println("-	Done..")
}

// QDBFetchUniqueTickersPG just takes whichever query that requests data and returns the result
// CAN ONLY BE USED TO FETCH ONE COLUMN
func QDBFetchUniqueTickersPG(ctx context.Context) []string {
	conn := QDBConnectPG(ctx)
	defer conn.Close(ctx)

	query := "SELECT ticker FROM 'tickers' WHERE ticker NOT IN (SELECT DISTINCT ticker FROM 'aggs');"
	//query := "SELECT DISTINCT ticker FROM 'tickers';"
	rows, err := conn.Query(ctx, query)
	defer rows.Close()
	CheckErr(err)

	var results []string
	for rows.Next() {
		var s string
		err = rows.Scan(&s)
		CheckErr(err)
		results = append(results, s)
	}

	return results
}
