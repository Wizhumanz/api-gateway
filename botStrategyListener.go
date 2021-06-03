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
			n := time.Now()
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
	// for n := range MinuteTicker().C {
	// 	fmt.Println("NOW: ", n)
	// 	fetchedCandles = fetchCandleData(ticker, period, n.Add(-time.Minute*1000), n.Add(-time.Minute*990))
	// 	fmt.Println(fetchedCandles)
	// }
	n := time.Now()
	fetchedCandles = fetchCandleData(ticker, period, n.Add(-time.Minute*2000), n.Add(-time.Minute*1990))
	fmt.Println(fetchedCandles)

}
