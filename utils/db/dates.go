package db

import (
	"lightning/utils/structs"
	"time"
)

// GetNewDatePair gets a new Date Pair, with the Interval already initialized.
func GetNewDatePair(start time.Time, end time.Time, target time.Time) structs.DatePairs {
	// If we're intitiating a new DatePair, we don't have an end, just a target
	// So just make sure the end is one more day ahead of the start date.
	if end.IsZero() {
		end = start.Add(24 * time.Hour)
	}
	dp := structs.DatePairs{Start: start, End: end, Target: target}
	_ = dp.SetInterval()
	return dp
}

// RecursivelyAddOneDay Just adds one day to the date pair, depending upon the end date
// if the interval between the two times is less than a day, just return end else add one more day
func RecursivelyAddOneDay(dp structs.DatePairs, end time.Time, results *[]structs.DatePairs) {
	// Check if the gap between the days is more than 1 day, and if so create a new DatePair with one day interval
	// And add it to results
	if dp.End.Before(dp.Target) || (dp.Interval >= 24) {
		// Establish that Start and End have to be updated, and Interval is set/updated.
		dp.Start = dp.Start.Add(24 * time.Hour)
		dp.End = dp.Start.Add(24 * time.Hour)
		dp.Target = end
		_ = dp.SetInterval()
		// Append to results
		*results = append(*results, dp)
		// Now recursively call self till it's done.
		RecursivelyAddOneDay(dp, end, results)
	} else {
		// Else, just finish up
		dp.End = end
		*results = append(*results, dp)
	}
}

func CreateDatePairs(from string, to string) *[]structs.DatePairs {
	// Parse the datetime provided
	Start, _ := time.Parse(TimeLayout, from)
	End, _ := time.Parse(TimeLayout, to)
	dp := GetNewDatePair(Start, time.Time{}, End)

	// Now, according to the dates provided, make a slice consisting of DatePairs each with a gap of one day.
	var results []structs.DatePairs
	RecursivelyAddOneDay(dp, End, &results)
	return &results
}
