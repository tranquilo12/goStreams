package db

import (
	"context"
	qdb "github.com/questdb/go-questdb-client"
	"lightning/utils/structs"
)

// QDBConnect to QuestDB and return a sender.
func QDBConnect(ctx context.Context) (*qdb.LineSender, error) {
	sender, err := qdb.NewLineSender(ctx)
	CheckErr(err)
	return sender, err
}

// QDBInsertTickers to QuestDB.
// Ensure the sender is deferred closed before this function is called.
func QDBInsertTickers(ctx context.Context, sender *qdb.LineSender, ticker structs.TickersStruct) {
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

// QDBInsert
