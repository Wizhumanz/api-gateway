package main

import (
	"fmt"
	"time"
)

//test strat to pass to backtest func, runs once per historical candlestick
func strat1(open, high, low, close []float64, relCandleIndex int, strategy *StrategySimulator, storage *interface{}) string {
	accRiskPerTrade := 0.5
	accSz := 1000
	leverage := 25 //limits raw price SL %

	if strategy.PosLongSize == 0 {
		//if two green candles in a row, buy
		if close[relCandleIndex] > (1.005 * open[relCandleIndex]) {
			// fmt.Printf("Buying at %v\n", close[relCandleIndex])
			entryPrice := close[relCandleIndex]
			slPrice := low[relCandleIndex-1]
			rawRiskPerc := (entryPrice - slPrice) / entryPrice
			accRiskedCap := accRiskPerTrade * float64(accSz)
			posCap := (accRiskedCap / rawRiskPerc) / float64(leverage)
			posSize := posCap / entryPrice
			// fmt.Printf("Entering with %v\n", posSize)
			strategy.Buy(close[relCandleIndex], slPrice, posSize, true, relCandleIndex)
			return fmt.Sprintf("▼ %.2f / %.2f", slPrice, posSize)
		}
	} else {
		sl := strategy.CheckPositions(open[relCandleIndex], high[relCandleIndex], low[relCandleIndex], close[relCandleIndex], relCandleIndex)

		//if two red candles in a row, sell
		if (strategy.PosLongSize > 0) && (open[relCandleIndex] > close[relCandleIndex]) {
			// fmt.Printf("Closing trade at %v\n", close[relCandleIndex])
			strategy.CloseLong(close[relCandleIndex], 0, relCandleIndex)
			return fmt.Sprintf("▼ %.2f", sl)
		}
	}

	return ""
}

func resetDisplayVars() {
	candleDisplay = []CandlestickChartData{}
	profitCurveDisplay = []ProfitCurveData{}
	simTradeDisplay = []SimulatedTradeData{}
}

func saveDisplayData(c Candlestick, strat StrategySimulator, relIndex int, label string) (CandlestickChartData, ProfitCurveDataPoint, SimulatedTradeDataPoint) {
	//candlestick
	newCandleD := CandlestickChartData{
		DateTime: c.DateTime,
		Open:     c.Open,
		High:     c.High,
		Low:      c.Low,
		Close:    c.Close,
	}
	//strategy enter/exit
	if strat.Actions[relIndex].Action == "ENTER" {
		newCandleD.StratEnterPrice = strat.Actions[relIndex].Price
	} else if strat.Actions[relIndex].Action == "SL" {
		newCandleD.StratExitPrice = strat.Actions[relIndex].Price
	}
	//label
	if label != "" {
		newCandleD.Label = label
	} else {
		// if strat.Actions[relIndex].Action == "ENTER" {
		// 	newCandleD.Label = fmt.Sprintf("<SL=\n%v", strat.Actions[relIndex].SL)
		// } else if strat.Actions[relIndex].Action == "SL" {
		// 	newCandleD.Label = fmt.Sprintf("<SL=%.2f / low=%.2f", strat.Actions[relIndex].SL, c.Low)
		// }
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
	userStrat func([]float64, []float64, []float64, []float64, int, *StrategySimulator, *interface{}) string,
	ticker, period string,
	startTime, endTime time.Time,
) {
	//get candles to test strat
	var periodCandles []Candlestick
	//check if data exists in cache
	redisKeyPrefix := ticker + ":" + period + ":"
	testKey := redisKeyPrefix + startTime.Format(httpTimeFormat) + ".0000000Z"
	testRes, _ := rdb.HGetAll(ctx, testKey).Result()
	if (testRes["open"] == "") && (testRes["close"] == "") {
		//if no data in cache, do fresh GET and save to cache
		periodCandles = fetchCandleData(ticker, period, startTime, endTime)
	} else {
		//otherwise, get data in cache
		periodCandles = getCachedCandleData(ticker, period, startTime, endTime)
	}

	//init strat testing
	strategySim := StrategySimulator{}
	strategySim.Init(500) //TODO: take func arg
	var storage interface{}

	resetDisplayVars()
	profitCurveDisplay = []ProfitCurveData{
		{
			Label: "strat1", //TODO: prep for dynamic strategy param values
		},
	}
	simTradeDisplay = []SimulatedTradeData{
		{
			Label: "strat1",
		},
	}

	//run strat on each candle
	allOpens := []float64{}
	allHighs := []float64{}
	allLows := []float64{}
	allCloses := []float64{}
	for i, candle := range periodCandles {
		allOpens = append(allOpens, candle.Open)
		allHighs = append(allHighs, candle.High)
		allLows = append(allLows, candle.Low)
		allCloses = append(allCloses, candle.Close)
		lb := userStrat(allOpens, allHighs, allLows, allCloses, i, &strategySim, &storage)
		//build display data using strategySim
		newCData, pcData, simTradeData := saveDisplayData(candle, strategySim, i, lb)
		candleDisplay = append(candleDisplay, newCData)
		profitCurveDisplay[0].Data = append(profitCurveDisplay[0].Data, pcData)
		if simTradeData.DateTime != "" {
			simTradeDisplay[0].Data = append(simTradeDisplay[0].Data, simTradeData)
		}
	}

	fmt.Println(colorGreen + "Backtest complete!" + colorReset)
}
