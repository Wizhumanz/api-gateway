package main

import (
	"fmt"
	"math"
	"time"
)

type PivotsStore struct {
	PivotHighs  []int
	PivotLows   []int
	LookForHigh bool
}

//return signature: (label, bars back to add label, storage obj to pass to next func call/iteration)
func strat1(
	risk, lev, accSz float64,
	open, high, low, close []float64,
	relCandleIndex int,
	strategy *StrategySimulator,
	storage interface{}) (string, int, interface{}) {
	// fmt.Printf("Risk = %v, Leverage = %v, AccCap = $%v \n", risk, lev, accSz)

	// stored bool var is "lookForHigh"
	stored, ok := storage.(PivotsStore)
	if !ok {
		if relCandleIndex == 0 {
			stored.PivotHighs = []int{}
			stored.PivotLows = []int{}
			stored.LookForHigh = true //default to looking for pivot high first
		} else {
			fmt.Errorf("storage obj assertion fail")
			return "", 0, storage
		}
	}
	lookForHigh := stored.LookForHigh

	//find pivot highs + lows
	pivotLabel := ""
	pivotBarsBack := 0
	var startIndex int
	if lookForHigh {
		//find next candle that crosses low of previous (PH)
		if len(stored.PivotLows) == 0 {
			startIndex = 1
		} else {
			startIndex = stored.PivotLows[len(stored.PivotLows)-1]
		}
		for j := startIndex; j < relCandleIndex-1; j++ {
			if low[j+1] < low[j] {
				// fmt.Printf("Found PH at index %v", j)
				stored.PivotHighs = append(stored.PivotHighs, j)
				pivotLabel = "PH"
				pivotBarsBack = relCandleIndex - j
				stored.LookForHigh = false
				break
			}
		}
	} else {
		//find next candle that crosses high of previous (PL)
		if len(stored.PivotHighs) == 0 {
			startIndex = 1
		} else {
			startIndex = stored.PivotHighs[len(stored.PivotHighs)-1]
		}
		for j := startIndex; j < relCandleIndex-1; j++ {
			if high[j+1] > high[j] {
				// fmt.Printf("Found PL at index %v", j)
				stored.PivotLows = append(stored.PivotLows, j)
				pivotLabel = "PL"
				pivotBarsBack = relCandleIndex - j
				stored.LookForHigh = true
			}
		}
	}

	// //look for entry
	// if strategy.PosLongSize == 0 && relCandleIndex > 0 {

	// }

	// if strategy.PosLongSize == 0 && relCandleIndex > 0 {
	// 	if (close[relCandleIndex] > open[relCandleIndex]) && (close[relCandleIndex-1] > open[relCandleIndex-1]) {
	// 		// fmt.Printf("Buying at %v\n", close[relCandleIndex])
	// 		entryPrice := close[relCandleIndex]
	// 		slPrice := low[relCandleIndex-1]
	// 		rawRiskPerc := (entryPrice - slPrice) / entryPrice
	// 		accRiskedCap := (risk / 100) * float64(accSz)
	// 		posCap := (accRiskedCap / rawRiskPerc) / float64(lev)
	// 		posSize := posCap / entryPrice
	// 		// fmt.Printf("Entering with %v\n", posSize)
	// 		strategy.Buy(close[relCandleIndex], slPrice, posSize, true, relCandleIndex)
	// 		// fmt.Printf("BUY IN %v\n", close[relCandleIndex])
	// 		return fmt.Sprintf("▼ %.2f / %.2f", slPrice, posSize), stored
	// 	}
	// } else if relCandleIndex > 0 {
	// 	sl := strategy.CheckPositions(open[relCandleIndex], high[relCandleIndex], low[relCandleIndex], close[relCandleIndex], relCandleIndex)

	// 	if (strategy.PosLongSize > 0) && (close[relCandleIndex] < open[relCandleIndex]) {
	// 		// fmt.Printf("Closing trade at %v\n", close[relCandleIndex])
	// 		strategy.CloseLong(close[relCandleIndex], 0, relCandleIndex)
	// 		// fmt.Printf("SELL EXIT %v\n", close[relCandleIndex])
	// 		return fmt.Sprintf("▼ %.2f", sl), stored
	// 	}
	// }

	return pivotLabel, pivotBarsBack, stored
}

func runBacktest(
	risk, lev, accSz float64,
	userStrat func(float64, float64, float64, []float64, []float64, []float64, []float64, int, *StrategySimulator, interface{}) (string, int, interface{}),
	userID, rid, ticker, period string,
	startTime, endTime time.Time,
	packetSize int, packetSender func(string, string, []CandlestickChartData, []ProfitCurveData, []SimulatedTradeData),
) ([]CandlestickChartData, []ProfitCurveData, []SimulatedTradeData) {

	//init
	var store interface{} //save state between strategy executions on each candle
	var retCandles []CandlestickChartData
	var retProfitCurve []ProfitCurveData
	var retSimTrades []SimulatedTradeData
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
	lastPacketEndIndexCandles := 0
	lastPacketEndIndexPC := 0
	lastPacketEndIndexSimT := 0
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
		testRes, _ := rdbChartmaster.HGetAll(ctx, testKey).Result()
		if (testRes["open"] == "") && (testRes["close"] == "") {
			//if no data in cache, do fresh GET and save to cache
			periodCandles = fetchCandleData(ticker, period, fetchCandlesStart, fetchCandlesEnd)
		} else {
			//otherwise, get data in cache
			periodCandles = getCachedCandleData(ticker, period, fetchCandlesStart, fetchCandlesEnd)
		}

		//run strat for all chunk's candles
		var label string
		var labelBB int
		for i, candle := range periodCandles {
			allOpens = append(allOpens, candle.Open)
			allHighs = append(allHighs, candle.High)
			allLows = append(allLows, candle.Low)
			allCloses = append(allCloses, candle.Close)
			//TODO: build results and run for different param sets
			label, labelBB, store = userStrat(risk, lev, accSz, allOpens, allHighs, allLows, allCloses, i, &strategySim, store)
			fmt.Println(store)

			//build display data using strategySim
			var newCData CandlestickChartData
			var pcData ProfitCurveDataPoint
			var simTradeData SimulatedTradeDataPoint
			newCData, pcData, simTradeData = saveDisplayData(candle, strategySim, i, label, labelBB, retProfitCurve[0].Data)
			retCandles = append(retCandles, newCData)
			if pcData.Equity > 0 {
				retProfitCurve[0].Data = append(retProfitCurve[0].Data, pcData)
			}
			if simTradeData.DateTime != "" {
				retSimTrades[0].Data = append(retSimTrades[0].Data, simTradeData)
			}
		}

		//stream data back to client in every chunk
		//rm duplicates
		var uniquePCPoints []ProfitCurveDataPoint
		for i, p := range retProfitCurve[0].Data {
			if len(uniquePCPoints) == 0 {
				if i != 0 {
					uniquePCPoints = append(uniquePCPoints, p)
				}
			} else {
				var found ProfitCurveDataPoint
				for _, search := range uniquePCPoints {
					if search.Equity == p.Equity {
						found = search
					}
				}

				if found.Equity == 0 && found.DateTime == "" {
					uniquePCPoints = append(uniquePCPoints, p)
				}
			}
		}
		retProfitCurve[0].Data = uniquePCPoints

		var uniqueStPoints []SimulatedTradeDataPoint
		for i, p := range retSimTrades[0].Data {
			if len(uniqueStPoints) == 0 {
				if i != 0 {
					uniqueStPoints = append(uniqueStPoints, p)
				}
			} else {
				var found SimulatedTradeDataPoint
				for _, search := range uniqueStPoints {
					if search.DateTime == p.DateTime {
						found = search
					}
				}

				if found.EntryPrice == 0 && found.DateTime == "" {
					uniqueStPoints = append(uniqueStPoints, p)
				}
			}
		}
		retSimTrades[0].Data = uniqueStPoints

		packetEndIndex := lastPacketEndIndexCandles + packetSize
		if packetEndIndex > len(retCandles) {
			packetEndIndex = len(retCandles)
		}
		// fmt.Printf("Sending candles %v to %v\n", lastPacketEndIndexCandles, packetEndIndex)
		pcFetchEndIndex := len(retProfitCurve[0].Data)
		packetPC := retProfitCurve[0].Data[lastPacketEndIndexPC:pcFetchEndIndex]
		stFetchEndIndex := len(retSimTrades[0].Data)
		packetSt := retSimTrades[0].Data[lastPacketEndIndexSimT:stFetchEndIndex]
		packetSender(userID, rid,
			retCandles[lastPacketEndIndexCandles:packetEndIndex],
			[]ProfitCurveData{
				{
					Label: "strat1", //TODO: prep for dynamic strategy param values
					Data:  packetPC,
				},
			},
			[]SimulatedTradeData{
				{
					Label: "strat1",
					Data:  packetSt,
				},
			})

		//save last index for streaming next chunk
		lastPacketEndIndexCandles = packetEndIndex
		lastPacketEndIndexPC = int(math.Max(float64(pcFetchEndIndex-1), float64(0)))
		lastPacketEndIndexSimT = int(math.Max(float64(stFetchEndIndex-1), float64(0)))
		//increment
		fetchCandlesStart = fetchCandlesEnd.Add(periodDurationMap[period])
	}

	fmt.Println(colorGreen + "Backtest complete!" + colorReset)
	return retCandles, retProfitCurve, retSimTrades
}
