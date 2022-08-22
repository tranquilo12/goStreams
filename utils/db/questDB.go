package db

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4"
	qdb "github.com/questdb/go-questdb-client"
	"github.com/schollz/progressbar/v3"
	"go.uber.org/ratelimit"
	"lightning/publisher"
	"lightning/utils/config"
	"lightning/utils/structs"
	"log"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"
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

	rows, err := conn.Query(ctx, "SELECT DISTINCT ticker FROM tickers WHERE market = 'stocks';")
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

// QDBQueryAndInsertAggILP to QuestDB.
func QDBQueryAndInsertAggILP(ctx context.Context, httpClient *http.Client, pbar *progressbar.ProgressBar, url *url.URL, timespan string, multiplier int) error {
	// Connect to QDB and get sender
	sender, _ := qdb.NewLineSender(ctx)

	// First query, then insert. If anything goes wrong with this go-routine, it should start with querying it again.
	var aggBar structs.AggregatesBarsResponse
	err := publisher.DownloadFromPolygonIO(httpClient, *url, &aggBar)
	CheckErr(err)

	// For each of these results, push!
	for _, agg := range aggBar.Results {
		err := sender.Table("aggs").
			Symbol("ticker", aggBar.Ticker).
			StringColumn("timespan", timespan).
			Int64Column("multiplier", int64(multiplier)).
			Float64Column("open", agg.O).
			Float64Column("high", agg.H).
			Float64Column("low", agg.L).
			Float64Column("close", agg.C).
			Float64Column("volume", agg.V).
			Float64Column("vw", agg.Vw).
			Float64Column("n", float64(agg.N)).
			At(ctx, time.UnixMilli(int64(agg.T)).UnixNano())
		if err != nil {
			return err
		}
	}

	// Make sure that the messages are sent over the network.
	err = sender.Flush(ctx)
	CheckErr(err)

	// Progress bar update
	pbar.Add(1)

	// close sender
	sender.Close()

	return nil
}

// QDBPushAllAggIntoDB Entire pipeline of querying all tickers and then pushing it to the db
func QDBPushAllAggIntoDB(ctx context.Context, urls []*url.URL, timespan string, multiplier int) {
	// Use a WaitGroup to make things simpler.
	// Create a buffer of the WaitGroup
	var wg sync.WaitGroup
	wg.Add(len(urls))

	// Done channel
	var doneCh chan bool

	// Max 300 requests per second
	prev := time.Now()
	rateLimiter := ratelimit.New(300)

	// Get the http client
	httpClient := config.GetHttpClient()

	// Init a progress pbar here
	progressbar.OptionSetWidth(500)
	//pbar := progressbar.Default(int64(len(urls)), "Downloading...")
	pbar := progressbar.NewOptions(len(urls),
		progressbar.OptionSetDescription("Downloading..."),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionSetWidth(10),
		progressbar.OptionThrottle(65*time.Millisecond),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionFullWidth(),
		progressbar.OptionSpinnerType(14),
		progressbar.OptionSetRenderBlankState(true),
		progressbar.OptionOnCompletion(func() {
			_, _ = fmt.Fprint(os.Stderr, "\n")
			doneCh <- true
		}),
	)

	// Iterate over all these urls, and insert them into the db!
	var err error
	for _, u := range urls {
		// Rate limit
		now := rateLimiter.Take()

		// Create a goroutine that will take care of the querying and insert
		go func() {
			err = Retry(10, 2, func() error {
				err = QDBQueryAndInsertAggILP(ctx, httpClient, pbar, u, timespan, multiplier)
				return err
			})
		}()

		// Rate limit, recalculate
		now.Sub(prev)
		prev = now
	}

	// Wait for all of them to finish.
	wg.Wait()

	// Just close the progressbar
	pbar.Close()

	// Done
	<-doneCh
}

// Retry A "decorator" function that wraps around every func that needs a retry.
func Retry(attempts int, sleep time.Duration, f func() error) (err error) {
	for i := 0; i < attempts; i++ {
		if i > 0 {
			log.Println("Retrying after error: ", err)
			time.Sleep(sleep)
		}
		err = f()
		if err == nil {
			return nil
		}
	}
	return fmt.Errorf("after %d attempts, last error: %s", attempts, err)
}
