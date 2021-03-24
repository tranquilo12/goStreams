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
			Timespan:     timespan,
			Multiplier:   multiplier,
			V:            r.V,
			Vw:           r.Vw,
			O:            r.O,
			C:            r.C,
			H:            r.H,
			L:            r.L,
			T:            t,
		}
		output = append(output, newArr)
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
