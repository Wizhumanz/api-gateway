package main

import (
	"fmt"
	"strings"
	"time"
)

func MinuteTicker(period string) *time.Ticker {
	c := make(chan time.Time, 1)
	t := &time.Ticker{C: c}
	var count float64
	go func() {
		for {
			n := time.Now().UTC()
			if n.Second() == 0 {
				count += 1
				fmt.Printf("\nCount: %v\n", count)
				fmt.Printf("\nTIME: %v\n", n)
			}
			if count > periodDurationMap[period].Minutes() {
				c <- n
				count = 0
			}
			time.Sleep(time.Second)
		}
	}()
	return t
}

func botStrategy(ticker, period string, risk, lev, accSz float64) {
	var fetchedCandles []Candlestick
	var store interface{} //save state between strategy executions on each candle

	strategySim := StrategySimulator{}
	allOpens := []float64{}
	allHighs := []float64{}
	allLows := []float64{}
	allCloses := []float64{}
	relIndex := 0
	// packetSize := 80

	timeNow := time.Now().UTC()
	fmt.Println("NOW: ", timeNow)

	fetchedCandles = fetchCandleData(ticker, period, timeNow.Add(-1*periodDurationMap[period]), timeNow.Add(-1*periodDurationMap[period]))
	fmt.Println(fetchedCandles)

	layout := "2006-01-02T15:04:05.000Z"
	str := strings.Replace(fetchedCandles[len(fetchedCandles)-1].PeriodEnd, "0000", "", 1)

	t, _ := time.Parse(layout, str)
	fmt.Println(t)

	for {
		if t == time.Now().UTC() {
			fmt.Println("WORKING")

			fetchedCandles = fetchCandleData(ticker, period, t.Add(-periodDurationMap[period]*1), t.Add(-periodDurationMap[period]*1))
			fmt.Println(fetchedCandles)

			for n := range MinuteTicker(period).C {
				fmt.Println("NOW: ", n)
				fetchedCandles = fetchCandleData(ticker, period, n.Add(-periodDurationMap[period]*1), n.Add(-periodDurationMap[period]*1))
				fmt.Println(fetchedCandles)

				//run strat for all chunk's candles

				allOpens = append(allOpens, fetchedCandles[0].Open)
				allHighs = append(allHighs, fetchedCandles[0].High)
				allLows = append(allLows, fetchedCandles[0].Low)
				allCloses = append(allCloses, fetchedCandles[0].Close)
				//TODO: build results and run for different param sets
				strat1(fetchedCandles[0], risk, lev, accSz, allOpens, allHighs, allLows, allCloses, relIndex, &strategySim, &store)

				//absolute index from absolute start of computation period
				relIndex++
			}
		}
	}

	// for n := range MinuteTicker(period).C {
	// 	fmt.Println("NOW: ", n)
	// 	fetchedCandles = fetchCandleData(ticker, period, n.Add(-periodDurationMap[period]*1), n.Add(-periodDurationMap[period]*1))
	// 	fmt.Println(fetchedCandles)

	// 	layout := "2006-01-02T15:04:05.000Z"
	// 	str := strings.Replace(fetchedCandles[len(fetchedCandles)-1].PeriodEnd, "0000", "", 1)

	// 	t, _ := time.Parse(layout, str)
	// 	fmt.Println(t)
	// }

	// n := time.Now().UTC()
	// fetchedCandles = fetchCandleData(ticker, period, n.Add(-1*periodDurationMap[period]), n.Add(-1*periodDurationMap[period]))
	// fmt.Println(fetchedCandles)

}
