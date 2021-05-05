package db

import (
	"fmt"
	"lightning/utils/structs"
	"strconv"
	"time"
)

// msToTime Function that takes in the string within the json, and returns the time.Unix element.
func msToTime(ms string) (time.Time, error) {
	msInt, err := strconv.ParseInt(ms, 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(0, msInt*int64(time.Millisecond)), nil
}

// AggBarFlattenPayloadBeforeInsert Function that flattens result from '/v2/aggs/ticker/{stocksTicker}/range/{multiplier}/{timespan}/{from}/{to}'
func AggBarFlattenPayloadBeforeInsert(target structs.AggregatesBarsResponse, timespan string, multiplier int) []structs.AggregatesBars {
	var output []structs.AggregatesBars
	results := target.Results

	for i := range results {
		// convert the string time into time.Unix time
		var r = results[i]
		t, err := msToTime(strconv.FormatInt(int64(r.T), 10))
		if err != nil {
			fmt.Println("Some Error parsing date: ", err)
		}

		// flatten here
		newArr := structs.AggregatesBars{
			InsertDate:   time.Now(),
			Ticker:       target.Ticker,
			Status:       target.Status,
			Querycount:   target.Querycount,
			Resultscount: target.Resultscount,
			Adjusted:     target.Adjusted,
			V:            r.V,
			Vw:           r.Vw,
			O:            r.O,
			C:            r.C,
			H:            r.H,
			L:            r.L,
			T:            t,
			N:            r.N,
			RequestID:    target.RequestID,
			Multiplier:   multiplier,
			Timespan:     timespan,
		}
		output = append(output, newArr)
	}
	return output
}

// TickerTypesFlattenPayloadBeforeInsert Function that flattens result from '/v2/reference/types'
func TickerTypesFlattenPayloadBeforeInsert(target *structs.TickerTypeResponse) structs.TickerType {
	var Tt structs.TickerType
	Tt.Cs = target.Results.Types.Cs
	Tt.Adr = target.Results.Types.Adr
	Tt.Nvdr = target.Results.Types.Nvdr
	Tt.Gdr = target.Results.Types.Gdr
	Tt.Sdr = target.Results.Types.Sdr
	Tt.Cef = target.Results.Types.Cef
	Tt.Etp = target.Results.Types.Etp
	Tt.Reit = target.Results.Types.Reit
	Tt.Mlp = target.Results.Types.Mlp
	Tt.Wrt = target.Results.Types.Wrt
	Tt.Pub = target.Results.Types.Pub
	Tt.Nyrs = target.Results.Types.Nyrs
	Tt.Unit = target.Results.Types.Unit
	Tt.Right = target.Results.Types.Right
	Tt.Track = target.Results.Types.Track
	Tt.Ltdp = target.Results.Types.Ltdp
	Tt.Rylt = target.Results.Types.Rylt
	Tt.Mf = target.Results.Types.Mf
	Tt.Pfd = target.Results.Types.Pfd
	Tt.Fdr = target.Results.Types.Fdr
	Tt.Ost = target.Results.Types.Ost
	Tt.Fund = target.Results.Types.Fund
	Tt.Sp = target.Results.Types.Sp
	Tt.Si = target.Results.Types.Si
	Tt.Index = target.Results.IndexTypes.Index
	Tt.Etf = target.Results.IndexTypes.Etf
	Tt.Etn = target.Results.IndexTypes.Etf
	Tt.Etmf = target.Results.IndexTypes.Etmf
	Tt.Settlement = target.Results.IndexTypes.Settlement
	Tt.Spot = target.Results.IndexTypes.Spot
	Tt.Subprod = target.Results.IndexTypes.Subprod
	Tt.Wc = target.Results.IndexTypes.Wc
	Tt.Alphaindex = target.Results.IndexTypes.Alphaindex
	return Tt
}

// TickersVxFlattenPayloadBeforeInsert Function that flattens result from '/v3/reference/tickers'
func TickersVxFlattenPayloadBeforeInsert(target structs.TickersVxResponse) []structs.TickerVx {
	var output []structs.TickerVx
	var results = target.Results

	for i := range results {
		var res = results[i]
		r := structs.TickerVx{
			InsertDatetime:  time.Now(),
			Ticker:          res.Ticker,
			Name:            res.Name,
			Market:          res.Market,
			Locale:          res.Locale,
			PrimaryExchange: res.PrimaryExchange,
			Type:            res.Type,
			Active:          res.Active,
			CurrencyName:    res.CurrencyName,
			Cik:             res.Cik,
			CompositeFigi:   res.CompositeFigi,
			ShareClassFigi:  res.ShareClassFigi,
			LastUpdatedUtc:  res.LastUpdatedUtc,
		}
		output = append(output, r)
	}

	return output
}

// TickersFlattenPayloadBeforeInsert Function that flattens result from '/v2/reference/tickers' (depreciated)
func TickersFlattenPayloadBeforeInsert(target structs.TickersResponse) []structs.Tickers {
	var output []structs.Tickers
	tickersInner := target.Tickers // creates TickersInner

	for i := range tickersInner {
		t := tickersInner[i]

		r := structs.Tickers{
			InsertDate:  time.Now(),
			Page:        target.Page,
			Perpage:     target.Perpage,
			Count:       target.Count,
			Status:      target.Status,
			Ticker:      t.Ticker,
			Name:        t.Name,
			Market:      t.Market,
			Locale:      t.Locale,
			Currency:    t.Currency,
			Active:      t.Active,
			Primaryexch: t.Primaryexch,
		}

		if t.Type != nil {
			r.Type = t.Type
		}

		if t.Codes != nil {
			r.Cik = t.Codes.Cik
			r.Figiuid = t.Codes.Figiuid
			r.Scfigi = t.Codes.Scfigi
			r.Cfigi = t.Codes.Cfigi
			r.Figi = t.Codes.Figi
		}

		r.Updated = t.Updated
		r.URL = t.URL

		if t.Attrs != nil {
			r.Currencyname = t.Attrs.Currencyname
			r.Currency = t.Attrs.Currency
			r.Basename = t.Attrs.Basename
			r.Base = t.Attrs.Base
		}
		output = append(output, r)
	}
	return output
}
