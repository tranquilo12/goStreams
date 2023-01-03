package db

import (
	"fmt"
	"time"
)

func NYSEHolidays(startYear, endYear int) map[string]bool {
	holidays := map[string]bool{}
	for year := startYear; year <= endYear; year++ {
		// New Year's Day
		holidays[fmt.Sprintf("%d-01-01", year)] = true
		// Martin Luther King Jr. Day
		holidays[fmt.Sprintf("%d-01-17", year)] = true
		// Presidents' Day
		holidays[fmt.Sprintf("%d-02-21", year)] = true
		// Memorial Day
		holidays[fmt.Sprintf("%d-05-30", year)] = true
		// Independence Day
		holidays[fmt.Sprintf("%d-07-04", year)] = true
		// Labor Day
		holidays[fmt.Sprintf("%d-09-05", year)] = true
		// Thanksgiving Day
		holidays[fmt.Sprintf("%d-11-25", year)] = true
		// Day after Thanksgiving
		holidays[fmt.Sprintf("%d-11-26", year)] = true
		// Christmas Eve (early close)
		holidays[fmt.Sprintf("%d-12-24", year)] = true
		// Christmas Day
		holidays[fmt.Sprintf("%d-12-25", year)] = true
	}
	return holidays
}

func BusinessDayPairs(start, end string) [][2]string {
	s, err := time.Parse(TimeLayout, start)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	e, err := time.Parse(TimeLayout, end)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	holidays := NYSEHolidays(s.Year(), e.Year())

	var pairs [][2]string
	for d := s; d.Before(e) || d.Equal(e); d = d.AddDate(0, 0, 1) {
		if d.Weekday() == time.Saturday || d.Weekday() == time.Sunday || holidays[d.Format(TimeLayout)] {
			continue
		}
		pairs = append(pairs, [2]string{d.Format(TimeLayout), d.AddDate(0, 0, 1).Format(TimeLayout)})
	}
	return pairs
}
