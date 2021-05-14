package main

import (
	"fmt"
	"time"
)

//test strat to pass to backtest func, runs once per historical candlestick
func strat1(
	open, high, low, close []float64,
	relCandleIndex int,
	strategy *StrategySimulator,
	storage *interface{}) string {
	accRiskPerTrade := 0.5
	accSz := 1000
	leverage := 25 //limits raw price SL %

	if strategy.PosLongSize == 0 && relCandleIndex > 0 {
		if (close[relCandleIndex] > open[relCandleIndex]) && (close[relCandleIndex-1] > open[relCandleIndex-1]) {
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
	} else if relCandleIndex > 0 {
		sl := strategy.CheckPositions(open[relCandleIndex], high[relCandleIndex], low[relCandleIndex], close[relCandleIndex], relCandleIndex)

		if (strategy.PosLongSize > 0) && (close[relCandleIndex] < open[relCandleIndex]) {
			// fmt.Printf("Closing trade at %v\n", close[relCandleIndex])
			strategy.CloseLong(close[relCandleIndex], 0, relCandleIndex)
			return fmt.Sprintf("▼ %.2f", sl)
		}
	}

	return ""
}

func runBacktest(
	userStrat func([]float64, []float64, []float64, []float64, int, *StrategySimulator, *interface{}) string,
	userID, rid, ticker, period string,
	startTime, endTime time.Time,
	packetSize int, packetSender func(string, string, []CandlestickChartData, []ProfitCurveData, []SimulatedTradeData),
) ([]CandlestickChartData, []ProfitCurveData, []SimulatedTradeData) {

	fmt.Println(userID)
	fmt.Println(ticker)
	fmt.Println(period)
	fmt.Println(packetSize)
	fmt.Println(startTime)
	fmt.Println(endTime)

	//init
	var retCandles []CandlestickChartData
	var retProfitCurve []ProfitCurveData
	var retSimTrades []SimulatedTradeData
	var storage interface{}
	retProfitCurve = []ProfitCurveData{
		{
			Label: "strat1", //TODO: prep for dynamic strategy param values
		},
	}
	retSimTrades = []SimulatedTradeData{
		{
			Label: "strat1",
		},
	}
	strategySim := StrategySimulator{}
	strategySim.Init(500) //TODO: take func arg

	//run backtest in chunks for client stream responsiveness
	allOpens := []float64{}
	allHighs := []float64{}
	allLows := []float64{}
	allCloses := []float64{}
	lastPacketEndIndex := 0
	fetchCandlesStart := startTime
	for {
		if fetchCandlesStart.After(endTime) {
			break
		}

		//get all candles of chunk
		var periodCandles []Candlestick

		fetchCandlesEnd := fetchCandlesStart.Add(periodDurationMap[period] * time.Duration(packetSize))
		if fetchCandlesEnd.After(endTime) {
			fetchCandlesEnd = endTime
		}

		//check if candles exist in cache
		redisKeyPrefix := ticker + ":" + period + ":"
		testKey := redisKeyPrefix + fetchCandlesStart.Format(httpTimeFormat) + ".0000000Z"
		testRes, _ := rdb.HGetAll(ctx, testKey).Result()
		if (testRes["open"] == "") && (testRes["close"] == "") {
			//if no data in cache, do fresh GET and save to cache
			periodCandles = fetchCandleData(ticker, period, fetchCandlesStart, fetchCandlesEnd)
		} else {
			//otherwise, get data in cache
			periodCandles = getCachedCandleData(ticker, period, fetchCandlesStart, fetchCandlesEnd)
		}

		//run strat for all chunk's candles
		for i, candle := range periodCandles {
			allOpens = append(allOpens, candle.Open)
			allHighs = append(allHighs, candle.High)
			allLows = append(allLows, candle.Low)
			allCloses = append(allCloses, candle.Close)
			//TODO: build results and run for different param sets
			lb := userStrat(allOpens, allHighs, allLows, allCloses, i, &strategySim, &storage)

			//build display data using strategySim
			var newCData CandlestickChartData
			var pcData ProfitCurveDataPoint
			var simTradeData SimulatedTradeDataPoint
			newCData, pcData, simTradeData = saveDisplayData(candle, strategySim, i, lb, retProfitCurve[0].Data)
			retCandles = append(retCandles, newCData)
			if pcData.Equity > 0 {
				retProfitCurve[0].Data = append(retProfitCurve[0].Data, pcData)
			}
			if simTradeData.DateTime != "" {
				retSimTrades[0].Data = append(retSimTrades[0].Data, simTradeData)
			}
		}

		//stream data back to client in every chunk
		packetEndIndex := lastPacketEndIndex + packetSize
		if packetEndIndex > len(retCandles) {
			packetEndIndex = len(retCandles) - 1
		}
		fmt.Printf("Sending candles %v to %v\n", lastPacketEndIndex, packetEndIndex)
		packetSender(userID, rid, retCandles[lastPacketEndIndex:packetEndIndex], retProfitCurve, retSimTrades)
		lastPacketEndIndex = packetEndIndex

		//increment
		fetchCandlesStart = fetchCandlesEnd.Add(periodDurationMap[period])
	}

	fmt.Println(colorGreen + "Backtest complete!" + colorReset)
	return retCandles, retProfitCurve, retSimTrades
}
