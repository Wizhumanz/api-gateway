package main

import (
	"fmt"
	"strings"
	"time"
)

func minuteTicker(period string) *time.Ticker {
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

func liveStrategyExecute(
	ticker, period string,
	userStrat func(Candlestick, float64, float64, float64, []float64, []float64, []float64, []float64, int, *StrategyExecutor, *interface{}) map[string]map[int]string) {
	var fetchedCandles []Candlestick

	timeNow := time.Now().UTC()

	//find time interval to trigger fetches
	checkCandle := fetchCandleData(ticker, period, timeNow.Add(-1*periodDurationMap[period]), timeNow.Add(-1*periodDurationMap[period]))
	layout := "2006-01-02T15:04:05.000Z"
	str := strings.Replace(checkCandle[len(checkCandle)-1].PeriodEnd, "0000", "", 1)
	t, _ := time.Parse(layout, str) //CoinAPI's standardized time interval

	for {
		//wait for current time to equal closest standardized interval time, t (only once)
		if t == time.Now().UTC() {
			//fetch closed latest candle (same as the one checked before)
			fetchedCandles = fetchCandleData(ticker, period, t.Add(-periodDurationMap[period]*1), t.Add(-periodDurationMap[period]*1))
			fmt.Println(fetchedCandles)

			//fetch candle and run live strat on every interval tick
			for n := range minuteTicker(period).C {
				fetchedCandles = fetchCandleData(ticker, period, n.Add(-periodDurationMap[period]*1), n.Add(-periodDurationMap[period]*1))
				//TODO: get bot's real settings to pass to strategy
				userStrat(fetchedCandles[0], 0.0, 0.0, 0.0,
					[]float64{fetchedCandles[0].Open},
					[]float64{fetchedCandles[0].High},
					[]float64{fetchedCandles[0].Low},
					[]float64{fetchedCandles[0].Close},
					-1, &StrategyExecutor{}, nil)
			}
		}
	}
}
