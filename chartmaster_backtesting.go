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
	candle Candlestick, risk, lev, accSz float64,
	open, high, low, close []float64,
	relCandleIndex int,
	strategy *StrategySimulator,
	storage *interface{}) map[string]map[int]string {
	// if len(close) > 0 {
	// 	fmt.Printf("len(close) = %v, last = %v", len(close), close[len(close)-1])
	// }
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
		for i := lastPivotIndex; (i+1) < len(low) && (i+1) < len(high); i++ { //TODO: should be relCandleIndex-1 but causes index outta range err
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
		for i := lastPivotIndex; (i+1) < len(high) && (i+1) < len(low); i++ {
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
					for f := newPLIndex - 1; f >= latestPHIndex && f > latestPLIndex && f < len(low) && f < len(high); f-- {
						if low[f] < low[newPLIndex] && !found {
							newPLIndex = f
						}
					}

					//check if current candle actually clears new selected candle as pivot high
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
			if len(stored.PivotLows)-2 >= 0 {
				currentPL := low[stored.PivotLows[len(stored.PivotLows)-1]]
				prevPL := low[stored.PivotLows[len(stored.PivotLows)-2]]
				if currentPL > prevPL {
					// fmt.Printf("Buying at %v\n", close[relCandleIndex-1])
					entryPrice := close[relCandleIndex-1]
					slPrice := prevPL
					rawRiskPerc := (entryPrice - slPrice) / entryPrice
					accRiskedCap := (risk / 100) * float64(accSz)
					posCap := (accRiskedCap / rawRiskPerc) / float64(lev)
					posSize := posCap / entryPrice
					// fmt.Printf("Entering with %v\n", posSize)
					strategy.Buy(close[relCandleIndex-1], slPrice, posSize, true, relCandleIndex)
					// fmt.Printf("BUY IN %v\n", close[relCandleIndex])
				}
			}
		}
	} else if strategy.PosLongSize > 0 && relCandleIndex > 0 { //long pos open
		if foundPH {
			strategy.CloseLong(close[relCandleIndex-1], 0, relCandleIndex, "TP")
			// newLabels["middle"] = map[int]string{
			// 	// pivotBarsBack: fmt.Sprintf("L from %v", relCandleIndex),
			// 	0: "EXIT TRADE " + fmt.Sprint(relCandleIndex),
			// }
		}
	}

	*storage = stored
	return newLabels
}

func getChunkCandleData(allCandles *[]Candlestick, chunkSlice *[]Candlestick, packetSize int, ticker, period string,
	startTime, endTime, fetchCandlesStart, fetchCandlesEnd time.Time, buildCandleDataSync chan string) {
	var chunkCandles []Candlestick
	//check if candles exist in cache
	redisKeyPrefix := ticker + ":" + period + ":"
	testKey := redisKeyPrefix + fetchCandlesStart.Format(httpTimeFormat) + ".0000000Z"
	testRes, _ := rdbChartmaster.HGetAll(ctx, testKey).Result()
	if (testRes["open"] == "") && (testRes["close"] == "") {
		fmt.Println("HELLO")
		//if no data in cache, do fresh GET and save to cache
		chunkCandles = fetchCandleData(ticker, period, fetchCandlesStart, fetchCandlesEnd)
	} else {
		fmt.Println("BYE")

		//otherwise, get data in cache
		chunkCandles = getCachedCandleData(ticker, period, fetchCandlesStart, fetchCandlesEnd)
		// fmt.Printf("Each Chunk: %v", chunkCandles)
	}
	fmt.Println("DONE")
	//append chunk's candles to global slice

	fmt.Printf("Get Chunks: %v", startTime.Add(time.Minute*80))
	// tempArr := *allCandles
	// if len(tempArr) != 0 {
	// 	fmt.Printf("Get Chunks: %v", len(tempArr)-1)
	// 	fmt.Printf("Get Chunks: %v", tempArr[len(tempArr)-1].DateTime)
	// }

	chunkLastCandleTime, err2 := time.Parse(httpTimeFormat, chunkCandles[len(chunkCandles)-1].DateTime)
	if err2 != nil {
		fmt.Printf("parsing lastCandleTime err = %v", err2)
	}
	msg := chunkLastCandleTime.Format(httpTimeFormat)

	buildCandleDataSync <- msg

	// for {
	// 	var tempArr []Candlestick
	// 	tempArr = *allCandles
	// 	var t time.Time
	// 	// fmt.Println(" ")
	// 	// fmt.Printf("LOLO: %v", msg)
	// 	if len(tempArr) != 0 {
	// 		t, _ = time.Parse("2006-01-02T15:04:05.000Z", tempArr[len(chunkCandles)-1].DateTime+".000Z")
	// 		fmt.Println("TIME:")
	// 		fmt.Println(t.Add(time.Minute * 81).Format("2006-01-02T15:04:05"))
	// 		break
	// 	}

	// 	if chunkLastCandleTime == startTime.Add(time.Minute*80) {
	*allCandles = append(*allCandles, chunkCandles...)
	*chunkSlice = chunkCandles
	// 		fmt.Printf("Help: %v", msg)

	// 		tempArr = *allCandles
	// 		fmt.Printf("\nTest: %v", t.Add(time.Minute*81).Format("2006-01-02T15:04:05"))
	// 		// tempArr[len(chunkCandles)-1].DateTime.Add(time.Minute*81) == msg
	// 		break
	// 	} else if len(tempArr) != 0 && msg == "2021-05-01T02:41:00" {
	// 		*allCandles = append(*allCandles, chunkCandles...)
	// 		t, _ := time.Parse("2006-01-02T15:04:05.000Z", tempArr[len(chunkCandles)-1].DateTime+".000Z")
	// 		fmt.Printf("Hi: %v", t.Add(time.Minute*80).Format("2006-01-02T15:04:05"))
	// 		break
	// 	} else if len(tempArr) != 0 && msg == "2021-05-01T04:00:00" {
	// 		*allCandles = append(*allCandles, chunkCandles...)
	// 		fmt.Printf("Hi: %v", msg)
	// 		break
	// 	}
	// 	// if len(tempArr) != 0 {
	// 	// 	fmt.Printf("\nTest: %v", len(tempArr) != 0 && tempArr[len(chunkCandles)-1].DateTime == msg)
	// 	// 	break
	// 	// }
	// }

	select {
	case buildCandleDataSync <- msg:
		fmt.Printf(colorGreen+"appended for chunk %v\n"+colorReset, chunkLastCandleTime.Format(httpTimeFormat))
	default:
		fmt.Printf("no message sent for chunk %v", chunkLastCandleTime.Format(httpTimeFormat))
	}
}

func runBacktest(
	risk, lev, accSz float64,
	userStrat func(Candlestick, float64, float64, float64, []float64, []float64, []float64, []float64, int, *StrategySimulator, *interface{}) map[string]map[int]string,
	userID, rid, ticker, period string,
	startTime, endTime time.Time,
	packetSize int, packetSender func(string, string, []CandlestickChartData, []ProfitCurveData, []SimulatedTradeData),
) ([]CandlestickChartData, []ProfitCurveData, []SimulatedTradeData) {
	//init
	buildCandleDataSync := make(chan string)
	var store interface{} //save state between strategy executions on each candle
	var retCandles []CandlestickChartData
	var retProfitCurve []ProfitCurveData
	var retSimTrades []SimulatedTradeData
	var chunksArr []*[]Candlestick
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
	var allCandleData []Candlestick

	allOpens := []float64{}
	allHighs := []float64{}
	allLows := []float64{}
	allCloses := []float64{}
	relIndex := 1
	fetchCandlesStart := startTime

	//fetch all candle data concurrently
	for {
		if fetchCandlesStart.After(endTime) {
			break
		}

		fetchCandlesEnd := fetchCandlesStart.Add(periodDurationMap[period] * time.Duration(packetSize))
		if fetchCandlesEnd.After(endTime) {
			fetchCandlesEnd = endTime
		}
		var chunkSlice []Candlestick

		chunksArr = append(chunksArr, &chunkSlice)
		fmt.Printf("\nCount: %v\n", chunksArr)
		go getChunkCandleData(&allCandleData, &chunkSlice, packetSize, ticker, period,
			startTime, endTime, fetchCandlesStart, fetchCandlesEnd, buildCandleDataSync)

		//increment
		fetchCandlesStart = fetchCandlesEnd.Add(periodDurationMap[period])
	}

	// fmt.Printf("\n total: %v\n", endTime.Sub(startTime).Minutes())
	//wait for all candle data fetch complete before running strategy
	for {

		fmt.Printf("len(allCandles) = %v, waiting for chan msg\n", len(allCandleData))
		msg := <-buildCandleDataSync
		fmt.Println("LOOOOOOOK HEERERERE: " + msg)
		for _, e := range chunksArr {
			fmt.Printf("\nHERE: %v\n", *e == nil)

		}

		// var msgTime time.Time
		if msg != "" {
			_, err := time.Parse(httpTimeFormat, msg)
			// msgTime = t
			if err != nil {
				fmt.Printf("wait all append complete time parse err = %v", err)
				continue
			}
		} else {
			continue
		}
		// (msgTime.After(endTime) || msgTime == endTime) &&
		if len(allCandleData) >= int(endTime.Sub(startTime).Minutes()) {
			fmt.Println("Ending")
			break
		}
	}
	fmt.Printf("LOOOOOOOK HEERERERE: %v", len(allCandleData))
	//run strat on all candles in chunk, stream each chunk to client
	stratComputeStartIndex := 0
	for {
		if stratComputeStartIndex > len(allCandleData) {
			break
		}

		stratComputeEndIndex := stratComputeStartIndex + packetSize
		if stratComputeEndIndex > len(allCandleData) {
			stratComputeEndIndex = len(allCandleData)
		}
		fmt.Printf("computing index %v to %v", stratComputeStartIndex, stratComputeEndIndex)
		periodCandles := allCandleData[stratComputeStartIndex:stratComputeEndIndex]

		//run strat for all chunk's candles
		var chunkAddedCandles []CandlestickChartData //separate chunk added vars to stream new data in packet only
		var chunkAddedPCData []ProfitCurveDataPoint
		var chunkAddedSTData []SimulatedTradeDataPoint
		var labels map[string]map[int]string
		for _, candle := range periodCandles {
			allOpens = append(allOpens, candle.Open)
			allHighs = append(allHighs, candle.High)
			allLows = append(allLows, candle.Low)
			allCloses = append(allCloses, candle.Close)
			//TODO: build results and run for different param sets
			labels = userStrat(candle, risk, lev, accSz, allOpens, allHighs, allLows, allCloses, relIndex, &strategySim, &store)

			//build display data using strategySim
			var pcData ProfitCurveDataPoint
			var simTradeData SimulatedTradeDataPoint
			chunkAddedCandles, pcData, simTradeData = saveDisplayData(chunkAddedCandles, candle, strategySim, relIndex, labels, chunkAddedPCData)
			if pcData.Equity > 0 {
				chunkAddedPCData = append(chunkAddedPCData, pcData)
			}
			if simTradeData.DateTime != "" {
				chunkAddedSTData = append(chunkAddedSTData, simTradeData)
			}

			//absolute index from absolute start of computation period
			relIndex++
		}

		//update more global vars
		retCandles = append(retCandles, chunkAddedCandles...)
		(retProfitCurve)[0].Data = append((retProfitCurve)[0].Data, chunkAddedPCData...)
		(retSimTrades)[0].Data = append((retSimTrades)[0].Data, chunkAddedSTData...)

		progressBar(userID, rid, retCandles, startTime, endTime)

		//stream data back to client in every chunk

		//stream data in order, wait for previous chunk to stream first
		// streamPacket := func() {
		// 	packetSender(userID, rid,
		// 		chunkAddedCandles,
		// 		[]ProfitCurveData{
		// 			{
		// 				Label: "strat1", //TODO: prep for dynamic strategy param values
		// 				Data:  chunkAddedPCData,
		// 			},
		// 		},
		// 		[]SimulatedTradeData{
		// 			{
		// 				Label: "strat1",
		// 				Data:  chunkAddedSTData,
		// 			},
		// 		})
		// }
		// fmt.Printf("\nLLOOOOOOOOK: %v", chunkAddedCandles)
		// go func() {
		// time.Sleep(2 * time.Second)
		// }()
		if chunkAddedCandles != nil {
			firstCandleTime, _ := time.Parse(httpTimeFormat, chunkAddedCandles[0].DateTime)

			if firstCandleTime == startTime {
				fmt.Printf(colorRed+"streaming index %v to %v"+colorReset, stratComputeStartIndex, stratComputeEndIndex)

				// streamPacket()

				// //allows next chunk to stream results
				// chunkLastCandleTime, _ := time.Parse(httpTimeFormat, chunkAddedCandles[len(chunkAddedCandles)-1].DateTime)
				// buildCandleDataSync <- chunkLastCandleTime.Format(httpTimeFormat)
			} else {
				fmt.Printf(colorRed+"streaming index %v to %v"+colorReset, stratComputeStartIndex, stratComputeEndIndex)

				// for {
				// 	previousChunkLastCandleTime := firstCandleTime.Add(-1 * periodDurationMap[period])
				// 	msg := <-buildCandleDataSync
				// 	// fmt.Printf("msg read by %v chunk = %v, awaiting %v\n", firstCandleTime, msg, previousChunkLastCandleTime)
				// 	if msg == previousChunkLastCandleTime.Format(httpTimeFormat) {
				// 		// fmt.Printf(colorCyan+"chunk %v sending WS packet\n"+colorReset, firstCandleTime)

				// 		defer func() {
				// 			if r := recover(); r != nil {
				// 				streamPacket()
				// 			}
				// 		}()

				// 		streamPacket()

				// 		//allows next chunk to stream results
				// 		chunkLastCandleTime, _ := time.Parse(httpTimeFormat, chunkAddedCandles[len(chunkAddedCandles)-1].DateTime)
				// 		buildCandleDataSync <- chunkLastCandleTime.Format(httpTimeFormat)
				// 		break
				// 	} else {
				// 		//pass on message if not intended receiver
				// 		buildCandleDataSync <- msg
				// 	}
				// }
			}

			stratComputeStartIndex = stratComputeEndIndex
		} else {
			break
		}
	}

	fmt.Println(colorGreen + "Backtest complete!" + colorReset)
	return retCandles, retProfitCurve, retSimTrades
}

func runBacktestSequential(
	risk, lev, accSz float64,
	userStrat func(Candlestick, float64, float64, float64, []float64, []float64, []float64, []float64, int, *StrategySimulator, *interface{}) map[string]map[int]string,
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
		var labels map[string]map[int]string
		for _, candle := range periodCandles {
			allOpens = append(allOpens, candle.Open)
			allHighs = append(allHighs, candle.High)
			allLows = append(allLows, candle.Low)
			allCloses = append(allCloses, candle.Close)
			//TODO: build results and run for different param sets
			labels = userStrat(candle, risk, lev, accSz, allOpens, allHighs, allLows, allCloses, relIndex, &strategySim, &store)

			//build display data using strategySim
			var pcData ProfitCurveDataPoint
			var simTradeData SimulatedTradeDataPoint
			retCandles, pcData, simTradeData = saveDisplayData(retCandles, candle, strategySim, relIndex, labels, retProfitCurve[0].Data)
			if pcData.Equity > 0 {
				retProfitCurve[0].Data = append(retProfitCurve[0].Data, pcData)
			}
			if simTradeData.DateTime != "" {
				retSimTrades[0].Data = append(retSimTrades[0].Data, simTradeData)
			}

			//absolute index from absolute start of computation period
			relIndex++
		}

		progressBar(userID, rid, retCandles, startTime, endTime)

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
