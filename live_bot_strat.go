package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gitlab.com/myikaco/msngr"
)

func activateBot(bot Bot) {
	// add new trade info into stream (triggers other services)
	msgs := []string{}
	msgs = append(msgs, "Timestamp")
	msgs = append(msgs, time.Now().Format("2006-01-02_15:04:05_-0700"))
	msgs = append(msgs, "BotID")
	msgs = append(msgs, fmt.Sprint(bot.K.ID))

	botStreamMsgs := []string{}
	botStreamMsgs = append(msgs, "Timestamp")
	botStreamMsgs = append(msgs, time.Now().Format("2006-01-02_15:04:05_-0700"))
	botStreamMsgs = append(msgs, "CMD")
	botStreamMsgs = append(msgs, "INIT")

	msngr.AddToStream(fmt.Sprint(bot.K.ID), botStreamMsgs)
	msngr.AddToStream("activeBots", msgs)
}

// logLiveStrategyExecution saves state of strategy execution loop to bot's dedicated stream in redis
func logLiveStrategyExecution(execTimestamp, storageObj, botStreamName string) {
	// add new trade info into stream (triggers other services)
	msgs := []string{}
	msgs = append(msgs, "Timestamp")
	msgs = append(msgs, execTimestamp)
	msgs = append(msgs, "StorageObj")
	msgs = append(msgs, storageObj)

	msngr.AddToStream(botStreamName, msgs)
}

func minuteTicker(period string) *time.Ticker {
	c := make(chan time.Time, 1)
	t := &time.Ticker{C: c}
	var count float64
	go func() {
		for {
			n := time.Now().UTC()
			if n.Second() == 0 {
				count += 1
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

//1. store state of running strategy loops (across multiple instances)
// a. api-gateway XADD trade stream ID to newTrades stream
// b. strat-svc listen on trade streams, add msg on every iteration of live loop
// c. strat-svc instances check unacknowledged entries in newTrades stream, then XAUTOCLAIM old msgs in trade stream

//X 2. store state of storage obj + relIndex + OHLC for each running live strategy loop
// key:JSON in redis

//3. how to stop running live strategy loop when bot status changed to inactive
// before exec strat on each iteration, check for ending command in trade stream

func executeLiveStrategy(
	bot Bot, ticker, period string,
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
				//TODO: fetch saved storage obj for strategy from redis
				var stratStore interface{}

				fetchedCandles = fetchCandleData(ticker, period, n.Add(-periodDurationMap[period]*1), n.Add(-periodDurationMap[period]*1))
				//TODO: get bot's real settings to pass to strategy
				userStrat(fetchedCandles[0], 0.0, 0.0, 0.0,
					[]float64{fetchedCandles[0].Open},
					[]float64{fetchedCandles[0].High},
					[]float64{fetchedCandles[0].Low},
					[]float64{fetchedCandles[0].Close},
					-1, &StrategyExecutor{}, &stratStore)

				//save state to retrieve for next iteration
				obj, err := json.Marshal(stratStore)
				if err != nil {
					fmt.Printf(colorRed+"%v\n"+colorReset, err)
				}
				logLiveStrategyExecution(n.Format(httpTimeFormat), string(obj), fmt.Sprint(bot.K.ID))
			}
		}
	}
}
