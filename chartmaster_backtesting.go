package main

import (
	"fmt"
	"math"
	"time"
)

type PivotsStore struct {
	PivotHighs []int
	PivotLows  []int
}

//return signature: (label, bars back to add label, storage obj to pass to next func call/iteration)
func strat1(
	risk, lev, accSz float64,
	open, high, low, close []float64,
	relCandleIndex int,
	strategy *StrategySimulator,
	storage *interface{}) map[string]map[int]string {
	if len(close) > 0 {
		fmt.Printf("len(close) = %v, last = %v", len(close), close[len(close)-1])
	}
	// fmt.Printf("Risk = %v, Leverage = %v, AccCap = $%v \n", risk, lev, accSz)

	// if relCandleIndex > 3 && relCandleIndex < 10 {
	// 	fmt.Printf("INDEX %v\n", relCandleIndex)
	// 	fmt.Printf("lows = %v\n", low)
	// 	fmt.Printf("candle low %v\n", low[len(low)-1])
	// }

	foundPL := false
	foundPH := false
	stored, ok := (*storage).(PivotsStore)
	if !ok {
		if relCandleIndex == 1 {
			stored.PivotHighs = []int{}
			stored.PivotLows = []int{}
		} else {
			fmt.Errorf("storage obj assertion fail")
			return nil
		}
	}

	//find pivot highs + lows
	lookForHigh := !(len(stored.PivotHighs) == len(stored.PivotLows)) //default to looking for low first
	// if relCandleIndex > 40 && relCandleIndex < 70 {
	// 	fmt.Println(colorRed + fmt.Sprintf("%v - %v | len(PH) = %v, len(PL) = %v | latestPH = %v, latestPL = %v", relCandleIndex, lookForLow, len(stored.PivotHighs), len(stored.PivotLows), stored.PivotHighs[len(stored.PivotHighs)-1], stored.PivotLows[len(stored.PivotLows)-1]) + colorReset)
	// }
	newLabels := make(map[string]map[int]string) //map of labelPos:map of labelBarsBack:labelText
	// newLabels["middle"] = map[int]string{
	// 	0: fmt.Sprintf("%v", relCandleIndex),
	// }
	pivotBarsBack := 0
	var lastPivotIndex int
	if len(stored.PivotHighs) == 0 || len(stored.PivotLows) == 0 {
		lastPivotIndex = 1
	} else {
		lastPivotIndex = int(math.Max(float64(stored.PivotHighs[len(stored.PivotHighs)-1]), float64(stored.PivotLows[len(stored.PivotLows)-1])))
		lastPivotIndex = int(math.Max(float64(1), float64(lastPivotIndex))) //make sure index is at least 1 to subtract 1 later
		lastPivotIndex++                                                    //don't allow both pivot high and low on same candle
	}
	if lookForHigh && relCandleIndex > 1 {
		//check if new candle took out the low of previous candles since last pivot
		for i := lastPivotIndex; i < relCandleIndex-1; i++ {
			if (low[i+1] < low[i]) && (high[i+1] < high[i]) {
				//check if pivot already exists
				found := false
				for _, ph := range stored.PivotHighs {
					if ph == i {
						found = true
						break
					}
				}
				if found {
					continue
				}
				// pivotLabel = pivotLabel + " LOW" + fmt.Sprint(low[i+1]) + " LOW" + fmt.Sprint(low[i]) + " "
				// fmt.Printf("Found PH at index %v", j)

				//find highest high since last PL
				newPHIndex := i
				// if relCandleIndex > 150 && relCandleIndex < 170 {
				// 	fmt.Printf(colorCyan+"<%v> ph index init search %v\n"+colorReset, relCandleIndex, newPHIndex)
				// }
				if len(stored.PivotLows) > 0 && len(stored.PivotHighs) > 0 && newPHIndex > 0 {
					latestPLIndex := stored.PivotLows[len(stored.PivotLows)-1]
					latestPHIndex := stored.PivotHighs[len(stored.PivotHighs)-1]
					for f := newPHIndex - 1; f >= latestPLIndex && f > latestPHIndex; f-- {
						if high[f] > high[newPHIndex] && !found {
							newPHIndex = f
						}
					}

					//check if current candle actually clears new selected candle as pivot high
					// if relCandleIndex > 150 && relCandleIndex < 170 {
					// 	fmt.Printf("Checking new PH index %v L = %v, H = %v + candle index %v L = %v, H = %v", newPHIndex, low[newPHIndex], high[newPHIndex], i+1, low[i+1], high[i+1])
					// }
					if !((low[i+1] < low[newPHIndex]) && (high[i+1] < high[newPHIndex])) {
						continue
					}
				}

				if newPHIndex > 0 {
					// fmt.Printf("Adding PH index %v\n", newPHIndex)
					stored.PivotHighs = append(stored.PivotHighs, newPHIndex)
					pivotBarsBack = relCandleIndex - newPHIndex - 1

					newLabels["top"] = map[int]string{
						// pivotBarsBack: fmt.Sprintf("H from %v", relCandleIndex),
						pivotBarsBack: "H",
					}
					foundPH = true
				}
			}
		}
	} else if relCandleIndex > 1 {
		for i := lastPivotIndex; i < relCandleIndex-1; i++ {
			if (high[i+1] > high[i]) && (low[i+1] > low[i]) {
				//check if pivot already exists
				found := false
				for _, pl := range stored.PivotLows {
					if pl == i {
						found = true
						break
					}
				}
				if found {
					continue
				}
				// fmt.Printf("Found PL at index %v", j)

				//find lowest low since last PL
				newPLIndex := i
				// if relCandleIndex > 150 && relCandleIndex < 170 {
				// 	fmt.Printf(colorYellow+"<%v> new PL init index = %v\n"+colorReset, relCandleIndex, newPLIndex)
				// }

				if len(stored.PivotHighs) > 0 && len(stored.PivotLows) > 0 && newPLIndex > 0 {
					latestPHIndex := stored.PivotHighs[len(stored.PivotHighs)-1]
					latestPLIndex := stored.PivotLows[len(stored.PivotLows)-1]
					// if relCandleIndex > 150 && relCandleIndex < 170 {
					// 	fmt.Printf("SEARCH lowest low latestPHIndex = %v, latestPLIndex = %v\n", latestPHIndex, latestPLIndex)
					// }
					for f := newPLIndex - 1; f >= latestPHIndex && f > latestPLIndex; f-- {
						if low[f] < low[newPLIndex] && !found {
							newPLIndex = f
						}
					}

					//check if current candle actually clears new selected candle as pivot high
					// if relCandleIndex > 150 && relCandleIndex < 170 {
					// 	fmt.Printf("Checking new PL index %v L = %v, H = %v + candle index %v L = %v, H = %v", newPLIndex, low[newPLIndex], high[newPLIndex], i+1, low[i+1], high[i+1])
					// }
					if !((high[i+1] > high[newPLIndex]) && (low[i+1] > low[newPLIndex])) {
						continue
					}
				}

				if newPLIndex > 0 {
					stored.PivotLows = append(stored.PivotLows, newPLIndex)
					pivotBarsBack = relCandleIndex - newPLIndex - 1
					newLabels["bottom"] = map[int]string{
						// pivotBarsBack: fmt.Sprintf("L from %v", relCandleIndex),
						pivotBarsBack: "L",
					}
					foundPL = true
					// fmt.Printf("Adding PL index %v\n", newPLIndex)
				}
			}
		}
	}

	//manage positions
	if (*strategy).PosLongSize == 0 && relCandleIndex > 0 { //no long pos
		//enter if current PL higher than previous
		if foundPL {
			currentPL := low[relCandleIndex-1]
			prevPL := low[stored.PivotLows[len(stored.PivotLows)-1]]
			if currentPL > prevPL {
				// fmt.Printf("Buying at %v\n", close[relCandleIndex-1])
				entryPrice := close[relCandleIndex-1]
				slPrice := prevPL
				rawRiskPerc := (entryPrice - slPrice) / entryPrice
				accRiskedCap := (risk / 100) * float64(accSz)
				posCap := (accRiskedCap / rawRiskPerc) / float64(lev)
				posSize := posCap / entryPrice
				// fmt.Printf("Entering with %v\n", posSize)
				(*strategy).Buy(close[relCandleIndex-1], slPrice, posSize, true, relCandleIndex)
				// fmt.Printf(colorGreen+"BUY IN %v\n"+colorReset, close[relCandleIndex-1])
			}
		}
	} else if strategy.PosLongSize > 0 && relCandleIndex > 0 { //long pos open
		if foundPH {
			// fmt.Printf("Closing trade at %v\n", close[relCandleIndex-1])
			(*strategy).CloseLong(close[relCandleIndex-1], 0, relCandleIndex)
			// fmt.Printf(colorRed+"SELL EXIT %v\n"+colorReset, close[relCandleIndex-1])
		}
	}

	*storage = stored
	return newLabels
}

func computeChunk(packetEndIndex, pcFetchEndIndex, stFetchEndIndex *int, store *interface{}, allOpens, allHighs, allLows, allCloses *[]float64,
	risk, lev, accSz float64, relIndex *int, packetSize int,
	userID, rid, ticker, period string,
	strategySim *StrategySimulator,
	retCandles *[]CandlestickChartData, retProfitCurve *[]ProfitCurveData, retSimTrades *[]SimulatedTradeData,
	startTime, endTime, fetchCandlesStart, fetchCandlesEnd time.Time,
	lastPacketEndIndexCandles, lastPacketEndIndexPC, lastPacketEndIndexSimT int,
	userStrat func(float64, float64, float64, []float64, []float64, []float64, []float64, int, *StrategySimulator, *interface{}) map[string]map[int]string,
	packetSender func(string, string, []CandlestickChartData, []ProfitCurveData, []SimulatedTradeData)) {
	//get all candles of chunk
	var periodCandles []Candlestick

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
	var labels map[string]map[int]string
	for i, candle := range periodCandles {
		*allOpens = append(*allOpens, candle.Open)
		*allHighs = append(*allHighs, candle.High)
		*allLows = append(*allLows, candle.Low)
		*allCloses = append(*allCloses, candle.Close)
		//TODO: build results and run for different param sets
		labels = userStrat(risk, lev, accSz, *allOpens, *allHighs, *allLows, *allCloses, *relIndex, strategySim, store)
		fmt.Println(strategySim.GetEquity())

		//build display data using strategySim
		var pcData ProfitCurveDataPoint
		var simTradeData SimulatedTradeDataPoint
		*retCandles, pcData, simTradeData = saveDisplayData(*retCandles, candle, *strategySim, i, labels, (*retProfitCurve)[0].Data)
		if pcData.Equity > 0 {
			(*retProfitCurve)[0].Data = append((*retProfitCurve)[0].Data, pcData)
		}
		if simTradeData.DateTime != "" {
			(*retSimTrades)[0].Data = append((*retSimTrades)[0].Data, simTradeData)
		}

		//absolute index from absolute start of computation period
		*relIndex++
	}

	progressBar(userID, rid, *retCandles, startTime, endTime)

	//stream data back to client in every chunk
	//rm duplicates
	var uniquePCPoints []ProfitCurveDataPoint
	for i, p := range (*retProfitCurve)[0].Data {
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
	(*retProfitCurve)[0].Data = uniquePCPoints

	var uniqueStPoints []SimulatedTradeDataPoint
	for i, p := range (*retSimTrades)[0].Data {
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
	(*retSimTrades)[0].Data = uniqueStPoints

	*packetEndIndex = lastPacketEndIndexCandles + packetSize
	if *packetEndIndex > len(*retCandles) {
		*packetEndIndex = len(*retCandles)
	}
	// fmt.Printf("Sending candles %v to %v\n", lastPacketEndIndexCandles, packetEndIndex)
	*pcFetchEndIndex = len((*retProfitCurve)[0].Data)
	packetPC := (*retProfitCurve)[0].Data[lastPacketEndIndexPC:*pcFetchEndIndex]
	*stFetchEndIndex = len((*retSimTrades)[0].Data)
	packetSt := (*retSimTrades)[0].Data[lastPacketEndIndexSimT:*stFetchEndIndex]
	packetSender(userID, rid,
		(*retCandles)[lastPacketEndIndexCandles:*packetEndIndex],
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
}

func runBacktest(
	risk, lev, accSz float64,
	userStrat func(float64, float64, float64, []float64, []float64, []float64, []float64, int, *StrategySimulator, *interface{}) map[string]map[int]string,
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
	strategySim.Init(accSz)

	//run backtest in chunks for client stream responsiveness
	allOpens := []float64{}
	allHighs := []float64{}
	allLows := []float64{}
	allCloses := []float64{}
	relIndex := 1
	lastPacketEndIndexCandles := 0
	lastPacketEndIndexPC := 0
	lastPacketEndIndexSimT := 0
	fetchCandlesStart := startTime
	var packetEndIndex, pcFetchEndIndex, stFetchEndIndex int
	for {
		if fetchCandlesStart.After(endTime) {
			break
		}

		fetchCandlesEnd := fetchCandlesStart.Add(periodDurationMap[period] * time.Duration(packetSize))
		if fetchCandlesEnd.After(endTime) {
			fetchCandlesEnd = endTime
		}

		// fmt.Printf("runnin with start = %v, end = %v\n", fetchCandlesStart.Format(httpTimeFormat), fetchCandlesEnd.Format(httpTimeFormat))
		// fmt.Printf("BEFORE len = %v, retCandles = %v\n", len(retCandles), retCandles)

		computeChunk(&packetEndIndex, &pcFetchEndIndex, &stFetchEndIndex, &store, &allOpens, &allHighs, &allLows, &allCloses,
			risk, lev, accSz, &relIndex, packetSize, userID, rid, ticker, period,
			&strategySim, &retCandles, &retProfitCurve, &retSimTrades, startTime, endTime, fetchCandlesStart, fetchCandlesEnd,
			lastPacketEndIndexCandles, lastPacketEndIndexPC, lastPacketEndIndexSimT,
			userStrat, packetSender)

		// fmt.Printf("AFTER len = %v, retCandles = %v\n", len(retCandles), retCandles)

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
