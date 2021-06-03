package main

import (
	"fmt"
	"time"
)

func MinuteTicker() *time.Ticker {
	c := make(chan time.Time, 1)
	t := &time.Ticker{C: c}
	go func() {
		for {
			n := time.Now().UTC()
			if n.Second() == 0 {
				c <- n
			}
			time.Sleep(time.Second)
		}
	}()
	return t
}

func botStrategy(ticker, period string) {
	var fetchedCandles []Candlestick
	for n := range MinuteTicker().C {
		fmt.Println("NOW: ", n)
		fetchedCandles = fetchCandleData(ticker, period, n.Add(-periodDurationMap[period]*1), n)
		fmt.Println(fetchedCandles)
	}
	// n := time.Now().UTC()
	// fetchedCandles = fetchCandleData(ticker, period, n.Add(-1*periodDurationMap[period]), n.Add(-1*periodDurationMap[period]))
	// fmt.Println(fetchedCandles)

}
