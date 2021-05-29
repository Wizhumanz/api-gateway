package main

import (
	"fmt"
	"math"
	"reflect"
	"sort"
	"time"
)

type PivotsStore struct {
	PivotHighs []int
	PivotLows  []int
}

//return signature: (label, bars back to add label, storage obj to pass to next func call/iteration)

func strat1(
	candles []Candlestick, risk, lev, accSz float64,
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
	// if stored.PivotLows

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

func getChunkCandleData(allCandles *[]Candlestick, packetSize int, ticker, period string,
	startTime, endTime, fetchCandlesStart, fetchCandlesEnd time.Time, buildCandleDataSync chan string) {
	var chunkCandles []Candlestick
	//check if candles exist in cache
	redisKeyPrefix := ticker + ":" + period + ":"
	testKey := redisKeyPrefix + fetchCandlesStart.Format(httpTimeFormat) + ".0000000Z"
	testRes, _ := rdbChartmaster.HGetAll(ctx, testKey).Result()
	if (testRes["open"] == "") && (testRes["close"] == "") {
		//if no data in cache, do fresh GET and save to cache
		chunkCandles = fetchCandleData(ticker, period, fetchCandlesStart, fetchCandlesEnd)
	} else {
		//otherwise, get data in cache
		chunkCandles = getCachedCandleData(ticker, period, fetchCandlesStart, fetchCandlesEnd)
		// fmt.Printf("Each Chunk: %v", chunkCandles)
	}
	//append chunk's candles to global slice
	*allCandles = append(*allCandles, chunkCandles...)
	// fmt.Printf("Get Chunks: %v", *allCandles)

	chunkLastCandleTime, err2 := time.Parse(httpTimeFormat, chunkCandles[len(chunkCandles)-1].DateTime)
	if err2 != nil {
		fmt.Printf("parsing lastCandleTime err = %v", err2)
	}
	msg := chunkLastCandleTime.Format(httpTimeFormat)
	select {
	case buildCandleDataSync <- msg:
		fmt.Printf(colorGreen+"appended for chunk %v\n"+colorReset, chunkLastCandleTime.Format(httpTimeFormat))
	default:
		fmt.Printf("no message sent for chunk %v", chunkLastCandleTime.Format(httpTimeFormat))
	}
}

func MaxFloatSlice(v []float64) float64 {
	sort.Float64s(v)
	return v[len(v)-1]
}

var comparePivotHighs []int
var comparePivotLows []int

func scan1(
	candles []Candlestick, risk, lev, accSz float64,
	open, high, low, close []float64,
	relCandleIndex int,
	strategy *StrategySimulator,
	storage *interface{}) map[string]map[int]string {

	// foundPL := false
	// foundPH := false
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

	newLabels := make(map[string]map[int]string) //map of labelPos:map of labelBarsBack:labelText

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

				//find highest high since last PL
				newPHIndex := i

				if len(stored.PivotLows) > 0 && len(stored.PivotHighs) > 0 && newPHIndex > 0 {
					latestPLIndex := stored.PivotLows[len(stored.PivotLows)-1]
					latestPHIndex := stored.PivotHighs[len(stored.PivotHighs)-1]
					for f := newPHIndex - 1; f >= latestPLIndex && f > latestPHIndex; f-- {
						if high[f] > high[newPHIndex] && !found {
							newPHIndex = f
						}
					}

					//check if current candle actually clears new selected candle as pivot high

					if !((low[i+1] < low[newPHIndex]) && (high[i+1] < high[newPHIndex])) {
						continue
					}
				}

				if newPHIndex > 0 {
					stored.PivotHighs = append(stored.PivotHighs, newPHIndex)
					pivotBarsBack = relCandleIndex - newPHIndex - 1

					newLabels["top"] = map[int]string{
						pivotBarsBack: "H",
					}
					// foundPH = true
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

				//find lowest low since last PL
				newPLIndex := i

				if len(stored.PivotHighs) > 0 && len(stored.PivotLows) > 0 && newPLIndex > 0 {
					latestPHIndex := stored.PivotHighs[len(stored.PivotHighs)-1]
					latestPLIndex := stored.PivotLows[len(stored.PivotLows)-1]

					for f := newPLIndex - 1; f >= latestPHIndex && f > latestPLIndex; f-- {
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
						pivotBarsBack: "L",
					}
					// foundPL = true
				}
			}
		}
	}

	/// WORK

	// comparePivotHighs = stored.PivotHighs
	// comparePivotLows = stored.PivotLows
	fmt.Println("relCandleIndex")
	fmt.Println(relCandleIndex)

	// fmt.Println(open)
	// fmt.Println(high)
	// fmt.Println(low)
	// fmt.Println(close)
	// !reflect.DeepEqual(comparePivotLows, stored.PivotLows)
	if !reflect.DeepEqual(comparePivotLows, stored.PivotLows) {
		var tempPivotLows []int

		for _, e := range stored.PivotLows {
			if !containsInt(comparePivotLows, e) {
				tempPivotLows = append(tempPivotLows, e)
			}
		}
		comparePivotLows = stored.PivotLows

		fmt.Println(stored.PivotHighs)
		fmt.Println(stored.PivotLows)
		startTrend := false
		var endTrend bool
		var startCandleIndex int
		var endCandleIndex int
		var trend upwardTrend
		var trendArray []upwardTrend
		var pivotHighsIndexInterval []int
		var allPivotHighsStart []float64
		var pivotHighsStart float64
		// var pivotHighsExtent float64
		var allPivotHighsExtent []float64
		// var growth []float64
		for i, _ := range stored.PivotLows {
			if i > 0 && !startTrend && low[stored.PivotLows[i-1]] < low[stored.PivotLows[i]] {
				startTrend = true
				startCandleIndex = stored.PivotLows[i-1]
				allPivotHighsStart = append(allPivotHighsStart, close[startCandleIndex])
				pivotHighsStart = close[startCandleIndex]
				fmt.Println("startCandleIndex")
				fmt.Println(startCandleIndex)
			} else if i > 0 && startTrend && !endTrend && low[stored.PivotLows[i-1]] > low[stored.PivotLows[i]] {
				endTrend = true
				endCandleIndex = stored.PivotLows[i-1]
				fmt.Println("endCandleIndex")
				fmt.Println(endCandleIndex)
			}
			if startTrend && endTrend {
				startTrend = false
				endTrend = false
				var pivotHighsInterval []float64

				//endTime
				for _, t := range stored.PivotHighs {
					if t > startCandleIndex && t < endCandleIndex {
						pivotHighsIndexInterval = append(pivotHighsIndexInterval, t)
						pivotHighsInterval = append(pivotHighsInterval, high[t])
					}
				}
				allPivotHighsExtent = append(allPivotHighsExtent, MaxFloatSlice(pivotHighsInterval))
				// fmt.Println("pivotHighsStart")
				// fmt.Println(pivotHighsStart)
				trend.Growth = MaxFloatSlice(pivotHighsInterval) - pivotHighsStart
				trend.Duration = endCandleIndex - startCandleIndex
				trendArray = append(trendArray, trend)
			}
		}

		// var arrayData []interface{}

		// ws := wsConnectionsChartmaster[userID]

		// arrayData = append(arrayData, trend)
		// streamPacket(ws, arrayData, rid)

		// fmt.Println(pivotHighsIndexInterval)
		fmt.Println(allPivotHighsStart)
		fmt.Println(allPivotHighsExtent)
		fmt.Println(trendArray)
	}

	return nil
}

func runBacktest(
	risk, lev, accSz float64,
	userStrat func([]Candlestick, float64, float64, float64, []float64, []float64, []float64, []float64, int, *StrategySimulator, *interface{}) map[string]map[int]string,
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

		go getChunkCandleData(&allCandleData, packetSize, ticker, period,
			startTime, endTime, fetchCandlesStart, fetchCandlesEnd, buildCandleDataSync)

		//increment
		fetchCandlesStart = fetchCandlesEnd.Add(periodDurationMap[period])
	}
	// fmt.Printf("\n total: %v\n", endTime.Sub(startTime).Minutes())
	//wait for all candle data fetch complete before running strategy
	for {
		fmt.Printf("len(allCandles) = %v, waiting for chan msg\n", len(allCandleData))
		msg := <-buildCandleDataSync
		// fmt.Println("LOOOOOOOK HEERERERE: " + msg)

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
			// fmt.Println("Ending")
			break
		}
	}
	// fmt.Printf("LOOOOOOOK HEERERERE: %v", len(allCandleData))
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
		var candles []Candlestick
		var chunkAddedCandles []CandlestickChartData //separate chunk added vars to stream new data in packet only
		var chunkAddedPCData []ProfitCurveDataPoint
		var chunkAddedSTData []SimulatedTradeDataPoint
		var labels map[string]map[int]string
		for i, candle := range periodCandles {
			candles = append(candles, candle)
			allOpens = append(allOpens, candle.Open)
			allHighs = append(allHighs, candle.High)
			allLows = append(allLows, candle.Low)
			allCloses = append(allCloses, candle.Close)
			//TODO: build results and run for different param sets
			labels = userStrat(candles, risk, lev, accSz, allOpens, allHighs, allLows, allCloses, relIndex, &strategySim, &store)

			//build display data using strategySim
			var pcData ProfitCurveDataPoint
			var simTradeData SimulatedTradeDataPoint
			chunkAddedCandles, pcData, simTradeData = saveDisplayData(chunkAddedCandles, candle, strategySim, i, labels, chunkAddedPCData)
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
