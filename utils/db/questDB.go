package db

import (
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
	qdb "github.com/questdb/go-questdb-client"
	"lightning/utils/structs"
	"time"
)

func CheckErr(err error) {
	if err != nil {
		panic(err)
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

func QDBFetchUrls(ctx context.Context) []string {
	// Connect to QDB
	conn := QDBConnectPG(ctx)
	defer conn.Close()

	// Query the database
	query := "SELECT url FROM 'urls' WHERE done = false ORDER BY ticker asc;"
	rows, err := conn.Query(ctx, query)
	defer rows.Close()
	CheckErr(err)

	var results []string
	for rows.Next() {
		// Create a new url.URL and scan the row into it
		var s string
		//var u *url.URL

		// Scan the row, into the string
		err = rows.Scan(&s)
		CheckErr(err)

		// Parse the url
		//u, err = url.Parse(s)
		//CheckErr(err)

		results = append(results, s)
	}

	// Delete the rows and return the results
	rows = nil

	return results
}

// QDBCheckAggsUrlsPG Checks if the data in aggs is already pulled from the urls table
func QDBCheckAggsUrlsPG(ctx context.Context) {
	conn := QDBConnectPG(ctx)
	defer conn.Close()

	subquery1 := "SELECT ticker, timestamp FROM aggs LATEST on timestamp PARTITION BY ticker"
	subquery2 := "SELECT q.* FROM q JOIN urls ON (ticker) WHERE `timestamp` <= end"
	query := "WITH q AS (" + subquery1 + "), q_url AS (" + subquery2 + ") UPDATE urls u SET done = true FROM q_url WHERE u.ticker = q_url.ticker;"

	_, err := conn.Exec(ctx, query)
	CheckErr(err)
}

// QDBCheckAggsLenPG checks if the length of the aggs table is 0
func QDBCheckAggsLenPG(ctx context.Context) bool {
	conn := QDBConnectPG(ctx)
	defer conn.Close()

	query := "SELECT count(*) FROM aggs;"
	rows, err := conn.Query(ctx, query)
	defer rows.Close()
	CheckErr(err)

	var count int
	for rows.Next() {
		err = rows.Scan(&count)
		CheckErr(err)
	}

	if count == 0 {
		return true
	}

	return false
}

// QDBGetUrlsLenPG get the length of the urls table
func QDBGetUrlsLenPG(ctx context.Context) int64 {
	conn := QDBConnectPG(ctx)
	defer conn.Close()

	query := "SELECT count(*) FROM urls;"
	rows, err := conn.Query(ctx, query)
	defer rows.Close()
	CheckErr(err)

	var count int
	for rows.Next() {
		err = rows.Scan(&count)
		CheckErr(err)
	}

	return int64(count)
}
