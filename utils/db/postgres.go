package db

import (
	"github.com/spf13/cobra"
	"lightning/utils/config"
)

// ReadPostgresDBParamsFromCMD A function that reads in parameters related to the postgres DB.

// ReadAggregateParamsFromCMD A function that reads in parameters related to the aggregate.
func ReadAggregateParamsFromCMD(cmd *cobra.Command) config.AggCliParams {

	timespan, _ := cmd.Flags().GetString("timespan")
	if timespan == "" {
		panic("Cmon provide some context, which --timespan??")
	}

	from_, _ := cmd.Flags().GetString("from")
	if from_ == "" {
		from_ = "2021-01-01"
	}

	to_, _ := cmd.Flags().GetString("to")
	if to_ == "" {
		to_ = "2021-03-01"
	}

	adjusted, _ := cmd.Flags().GetInt("adjusted")

	res := config.AggCliParams{
		Timespan: timespan,
		From:     from_,
		To:       to_,
		Adjusted: adjusted,
	}

	return res
}

// ReadTickerNewsParamsFromCMD reads parameters like ticker, startDate, endDate

// GetPostgresDBConn Makes sure the connection object to the postgres instance is returned.

// ExecCreateAllTablesModels Makes sure CreateAllTablesModels() is called and all table models are made.
