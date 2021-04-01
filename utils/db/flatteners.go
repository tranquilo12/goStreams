package db

import (
	"fmt"
	"lightning/utils/structs"
	"strconv"
	"time"
)

func msToTime(ms string) (time.Time, error) {
	msInt, err := strconv.ParseInt(ms, 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(0, msInt*int64(time.Millisecond)), nil
}

func AggBarFlattenPayloadBeforeInsert1(target structs.AggregatesBarsResponse, timespan string, multiplier int) []structs.AggregatesBars {
	var output []structs.AggregatesBars

	results := target.Results

	for i := range results {
		var r = results[i]

		t, err := msToTime(strconv.FormatInt(int64(r.T), 10))
		if err != nil {
			fmt.Println("Some Error parsing date: ", err)
		}

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

func TickersFlattenPayloadBeforeInsert(target structs.TickersResponse) []structs.Tickers {
	var output []structs.Tickers
	tickersInner := target.Tickers // creates TickersInner

	for i := range tickersInner {
		t := tickersInner[i]

		r := structs.Tickers{}
		r.InsertDate = time.Now()
		r.Page = target.Page
		r.Perpage = target.Perpage
		r.Count = target.Count
		r.Status = target.Status
		r.Ticker = t.Ticker
		r.Name = t.Name
		r.Market = t.Market
		r.Locale = t.Locale
		r.Currency = t.Currency
		r.Active = t.Active
		r.Primaryexch = t.Primaryexch

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

//func AggBarFlattenPayloadBeforeInsert(inputChan <- chan structs.AggregatesBarsResponse, timespan string, multiplier int) <- chan structs.AggregatesBars {
//	// define a WaitGroup, that will be used to control the start/stop of the exec of goroutines
//	println("Started")
//	var wg sync.WaitGroup
//
//	// make an output channel, that will store the Expanded flattener and will be returned at the end of
//	outputChan := make(chan structs.AggregatesBars, len(inputChan))
//	println("Started 2")
//
//	// define a function here that will be "fanned-outputChan" by using goroutines
//	flattener := func(c structs.AggregatesBarsResponse) {
//
//		// for each element in inputChannel (containing un-expanded json structure), expand and store in channel
//		for n := range inputChan {
//			println("Started 3")
//			results := n.Results
//
//			for i := range results {
//				var r = results[i]
//
//				// convert the time string into time.Time
//				t, err := msToTime(strconv.FormatInt(int64(r.T), 10))
//				if err != nil {
//					fmt.Println("Some Error parsing date: ", err)
//				}
//
//				// convert to the final expanded structure
//				res := structs.AggregatesBars{
//					InsertDate:   time.Now(),
//					Ticker:       n.Ticker,
//					Status:       n.Status,
//					Querycount:   n.Querycount,
//					Resultscount: n.Resultscount,
//					Adjusted:     n.Adjusted,
//					Timespan:     timespan,
//					Multiplier:   multiplier,
//					V:            r.V,
//					Vw:           r.Vw,
//					O:            r.O,
//					C:            r.C,
//					H:            r.H,
//					L:            r.L,
//					T:            t,
//				}
//
//				// push to channel
//				fmt.Printf("Something: %v", res)
//				outputChan <- res
//			}
//		}
//
//		// for each channel input, after this operation is completed ... tell waitGroup that it's done
//		wg.Done()
//	}
//
//	// Add the length of input channels to the WaitGroup, to make it ready for all goroutines that it will executed
//	wg.Add(len(inputChan))
//
//	// for each input in the channel, execute goroutine
//	for toFlatten := range inputChan {
//		go flattener(toFlatten)
//	}
//
//	// A lambda function that ensures the WaitGroup waits for all channels, before the channel is closed
//	go func() {
//		wg.Wait()
//		close(outputChan)
//	}()
//
//	return outputChan
//}
