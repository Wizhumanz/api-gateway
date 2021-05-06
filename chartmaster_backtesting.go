package main

import (
	"context"
	"fmt"
	"strconv"
	"time"
)

var candleDisplay []CandlestickChartData
var profitCurveDisplay []ProfitCurveData
var simTradeDisplay []SimulatedTradeData

//test strat to pass to backtest func, runs once per historical candlestick
func strat1(open, high, low, close []float64, relCandleIndex int, strategy *StrategySimulator, storage *interface{}) {
	if strategy.PosLongSize == 0 {
		//if two green candles in a row, buy
		if (close[0] > open[0]) && (close[1] > open[1]) {
			strategy.Buy(close[0], 69.69, true)
		}
	} else {
		//if two red candles in a row, sell
		if (open[0] > close[0]) && (open[1] > close[1]) {
			strategy.CloseLong(close[0], 69.69)
		}
	}
}

func resetDisplayVars() {
	candleDisplay = []CandlestickChartData{}
	profitCurveDisplay = []ProfitCurveData{}
	simTradeDisplay = []SimulatedTradeData{}
}

func saveDisplayData(c Candlestick, strat StrategySimulator) {
	//candlestick
	newCandleD := CandlestickChartData{
		DateTime: c.DateTime,
		Open:     c.Open,
		High:     c.High,
		Low:      c.Low,
		Close:    c.Close,
	}
	//strategy enter/exit/label
	if strat.Actions[0].Action == "ENTER" {
		newCandleD.StratEnterPrice = strat.Actions[0].Price
		newCandleD.Label = fmt.Sprintf("SL = %v", strat.Actions[0].SL)
	} else if strat.Actions[0].Action == "EXIT" {
		newCandleD.StratExitPrice = strat.Actions[0].Price
		newCandleD.Label = fmt.Sprintf("SL = %v", strat.Actions[0].SL)
	}
	candleDisplay = append(candleDisplay, newCandleD)

	//TODO: profit curve
	//TODO: sim trades
}

func runBacktest(
	userStrat func(float64, float64, float64, float64, int, *StrategySimulator, *interface{}),
) {
	//get all candlestick data for selected backtest period
	format := "2006-01-02T15:04:05"
	startDateTime, _ := time.Parse(format, "2021-05-01T00:00:00") //TODO: get this as func arg
	data := []Candlestick{}
	for i := 0; i < 100; i++ {
		var new Candlestick
		ctx := context.Background()
		key := "BTCUSDT:1MIN:" + startDateTime.Format(format) + ".0000000Z"
		res, _ := rdb.HGetAll(ctx, key).Result()

		new.DateTime = startDateTime.Format(format)
		new.Open, _ = strconv.ParseFloat(res["open"], 32)
		new.High, _ = strconv.ParseFloat(res["open"], 32)
		new.Low, _ = strconv.ParseFloat(res["open"], 32)
		new.Close, _ = strconv.ParseFloat(res["open"], 32)
		data = append(data, new)

		startDateTime = startDateTime.Add(1 * time.Minute)
	}

	strategySim := StrategySimulator{}
	strategySim.Init(1000)
	var storage interface{}

	for i, candle := range data {
		userStrat(candle.Open, candle.High, candle.Low, candle.Close, i, &strategySim, &storage)
		//build display data using strategySim
		resetDisplayVars()
		saveDisplayData(candle, strategySim)
	}
}
