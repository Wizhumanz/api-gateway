package main

import (
	"context"
	"fmt"
	"strconv"
	"time"
)

//test strat to pass to backtest func, runs once per historical candlestick
func strat1(open, high, low, close []float64, relCandleIndex int, strategy *StrategySimulator, storage *interface{}) {
	accRiskPerTrade := 0.5
	accSz := 1000
	leverage := 25 //limits raw price SL %

	strategy.CheckPositions(open[relCandleIndex], high[relCandleIndex], low[relCandleIndex], close[relCandleIndex])

	if strategy.PosLongSize == 0 {
		//if two green candles in a row, buy
		if (close[relCandleIndex] > open[relCandleIndex]) && (close[relCandleIndex-1] > open[relCandleIndex-1]) {
			// fmt.Printf("Buying at %v\n", close[relCandleIndex])
			entryPrice := close[relCandleIndex]
			slPrice := close[relCandleIndex-2]
			rawRiskPerc := (entryPrice - slPrice) / entryPrice
			accRiskedCap := accRiskPerTrade * float64(accSz)
			posCap := (accRiskedCap / rawRiskPerc) / float64(leverage)
			posSize := posCap / entryPrice
			fmt.Printf("Entering with %v\n", posSize)
			strategy.Buy(close[relCandleIndex], posSize, true)

			strategy.Actions = append(strategy.Actions, StrategySimulatorAction{
				Action: "ENTER",
				Price:  close[relCandleIndex],
				SL:     slPrice,
			})
			return
		}
	} else {
		//if two red candles in a row, sell
		if (open[relCandleIndex] > close[relCandleIndex]) && (open[relCandleIndex-1] > close[relCandleIndex-1]) {
			// fmt.Printf("Closing trade at %v\n", close[relCandleIndex])
			strategy.CloseLong(close[relCandleIndex], 0)
			strategy.Actions = append(strategy.Actions, StrategySimulatorAction{
				Action: "SL",
				Price:  close[relCandleIndex],
			})
			return
		}
	}

	//if no action taken, add blank action to maintain index
	strategy.Actions = append(strategy.Actions, StrategySimulatorAction{})
}

func resetDisplayVars() {
	candleDisplay = []CandlestickChartData{}
	profitCurveDisplay = []ProfitCurveData{}
	simTradeDisplay = []SimulatedTradeData{}
}

func saveDisplayData(c Candlestick, strat StrategySimulator, relIndex int) (CandlestickChartData, ProfitCurveDataPoint, SimulatedTradeDataPoint) {
	//candlestick
	newCandleD := CandlestickChartData{
		DateTime: c.DateTime,
		Open:     c.Open,
		High:     c.High,
		Low:      c.Low,
		Close:    c.Close,
	}
	//strategy enter/exit/label
	if strat.Actions[relIndex].Action == "ENTER" {
		newCandleD.StratEnterPrice = strat.Actions[relIndex].Price
		newCandleD.Label = fmt.Sprintf("SL = %v", strat.Actions[relIndex].SL)
	} else if strat.Actions[relIndex].Action == "SL" {
		newCandleD.StratExitPrice = strat.Actions[relIndex].Price
	}

	//profit curve
	pd := ProfitCurveDataPoint{
		DateTime: c.DateTime,
		Equity:   strat.GetEquity(),
	}

	//sim trades
	sd := SimulatedTradeDataPoint{}
	if strat.Actions[relIndex].Action == "SL" || strat.Actions[relIndex].Action == "TP" {
		sd.DateTime = c.DateTime
		sd.Direction = "LONG"                               //TODO: fix later when strategy changes
		sd.EntryPrice = strat.Actions[relIndex].Price - 1.0 //TODO: calculate actual entry price
		sd.ExitPrice = strat.Actions[relIndex].Price
		//TODO: add more props to strategy Actions
		sd.PosSize = 69.69
		sd.RiskedEquity = 699.69
		sd.RawProfitPerc = 0.69
	}

	return newCandleD, pd, sd
}

func runBacktest(
	userStrat func([]float64, []float64, []float64, []float64, int, *StrategySimulator, *interface{}),
) {
	//get all candlestick data for selected backtest period
	format := "2006-01-02T15:04:05"
	startDateTime, _ := time.Parse(format, "2021-05-01T00:00:00") //TODO: get this from func arg
	data := []Candlestick{}
	for i := 0; i < 300; i++ {
		var new Candlestick
		ctx := context.Background()
		key := "BTCUSDT:1MIN:" + startDateTime.Format(format) + ".0000000Z"
		res, _ := rdb.HGetAll(ctx, key).Result()

		new.DateTime = startDateTime.Format(format)
		new.Open, _ = strconv.ParseFloat(res["open"], 32)
		new.High, _ = strconv.ParseFloat(res["high"], 32)
		new.Low, _ = strconv.ParseFloat(res["low"], 32)
		new.Close, _ = strconv.ParseFloat(res["close"], 32)
		data = append(data, new)

		startDateTime = startDateTime.Add(1 * time.Minute)
	}

	strategySim := StrategySimulator{}
	strategySim.Init(500)
	var storage interface{}

	resetDisplayVars()
	profitCurveDisplay = []ProfitCurveData{
		{
			Label: "strat1",
		},
	}
	simTradeDisplay = []SimulatedTradeData{
		{
			Label: "strat1",
		},
	}

	allOpens := []float64{}
	allHighs := []float64{}
	allLows := []float64{}
	allCloses := []float64{}
	for i, candle := range data {
		allOpens = append(allOpens, candle.Open)
		allHighs = append(allHighs, candle.High)
		allLows = append(allLows, candle.Low)
		allCloses = append(allCloses, candle.Close)
		userStrat(allOpens, allHighs, allLows, allCloses, i, &strategySim, &storage)
		//build display data using strategySim
		newCData, pcData, simTradeData := saveDisplayData(candle, strategySim, i)
		candleDisplay = append(candleDisplay, newCData)
		profitCurveDisplay[0].Data = append(profitCurveDisplay[0].Data, pcData)
		if simTradeData.DateTime != "" {
			simTradeDisplay[0].Data = append(simTradeDisplay[0].Data, simTradeData)
		}
	}
}
