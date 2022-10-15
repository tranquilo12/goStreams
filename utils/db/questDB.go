package db

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	qdb "github.com/questdb/go-questdb-client"
	"lightning/utils/structs"
	"time"
)

// CheckErr checks for errors and panics if there is one
func CheckErr(err error) {
	if err != nil {
		fmt.Println(err)
	}
}

// QDBConnectPG creates a Postgres conn to QuestDB.
func QDBConnectPG(ctx context.Context) *pgxpool.Pool {
	pool, _ := pgxpool.Connect(ctx, "postgresql://admin:quest@localhost:8812/")
	return pool
}

// QDBInsertTickersILP to QuestDB.
// Ensure the sender is deferred closed before this function is called.
func QDBInsertTickersILP(ctx context.Context, ticker structs.TickersStruct) {
	// Connect to QDB and get sender
	sender, _ := qdb.NewLineSender(ctx)

	// Send all the data within the ticker
	for _, t := range ticker.Results {
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

		// Make sure that the messages are sent over the network.
		err = sender.Flush(ctx)
		CheckErr(err)
	}

	// Close the sender here
	sender.Close()
}

// QDBFetchUniqueTickersPG just takes whichever query that requests data and returns the result
// CAN ONLY BE USED TO FETCH ONE COLUMN
func QDBFetchUniqueTickersPG(ctx context.Context) []string {
	// Connect to QDB
	conn := QDBConnectPG(ctx)
	defer conn.Close()

	// Query the database
	query := "SELECT DISTINCT ticker FROM 'tickers' ORDER BY ticker asc;"
	rows, err := conn.Query(ctx, query)
	CheckErr(err)

	// Close the rows
	defer rows.Close()

	// Iterate through the rows and append the results to the slice
	var results []string
	for rows.Next() {
		var s string
		err = rows.Scan(&s)
		CheckErr(err)
		results = append(results, s)
	}

	// Delete the rows and return the results
	rows = nil

	return results
}

// QDBFetchAllUrls returns all the urls from the database, based on the retry flag, and the limit flag
func QDBFetchAllUrls(ctx context.Context, retry bool, limit int) []string {
	conn := QDBConnectPG(ctx)
	defer conn.Close()

	query := ""

	// Query the database
	switch limit {
	case -1:
		query = fmt.Sprintf("SELECT url FROM 'urls' WHERE retry = %t ORDER BY ticker, start asc;", retry)
	default:
		query = fmt.Sprintf("SELECT url FROM 'urls' WHERE retry = %t ORDER BY ticker, start asc LIMIT %d;", retry, limit)
	}

	// Finally make the query here
	rows, err := conn.Query(ctx, query)
	defer rows.Close()
	CheckErr(err)

	var results []string
	for rows.Next() {
		// Create a new url.URL and scan the row into it
		var s string

		// Scan the row, into the string
		err = rows.Scan(&s)
		CheckErr(err)

		results = append(results, s)
	}

	return results
}

// QDBFetchUrlsByTicker returns the urls for a specific ticker, no limits.
func QDBFetchUrlsByTicker(ctx context.Context, ticker string) []string {
	conn := QDBConnectPG(ctx)
	defer conn.Close()

	// Query the database
	query := fmt.Sprintf("SELECT url FROM 'urls' WHERE ticker = '%s' ORDER BY start asc;", ticker)
	rows, err := conn.Query(ctx, query)
	CheckErr(err)

	// Close the rows
	defer rows.Close()

	// Iterate through the rows and append the results to the slice
	var results []string
	for rows.Next() {
		var s string
		err = rows.Scan(&s)
		CheckErr(err)
		results = append(results, s)
	}

	// Delete the rows and return the results
	rows = nil

	return results
}
