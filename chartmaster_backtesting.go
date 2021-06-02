package main

import (
	"fmt"
	"math"
	"time"
)

type PivotsStore struct {
	PivotHighs     []int
	PivotLows      []int
	LongEntryPrice float64
	LongSLPrice    float64
	LongPosSize    float64
}

//return signature: (label, bars back to add label, storage obj to pass to next func call/iteration)
func strat1(
	candle Candlestick, risk, lev, accSz float64,
	open, high, low, close []float64,
	relCandleIndex int,
	strategy *StrategySimulator,
	storage *interface{}) map[string]map[int]string {
	tpPerc := 0.5

	stored, ok := (*storage).(PivotsStore)
	if !ok {
		if relCandleIndex == 0 {
			stored.PivotHighs = []int{}
			stored.PivotLows = []int{}
		} else {
			fmt.Errorf("storage obj assertion fail")
			return nil
		}
	}

	newLabels, foundPL := findPivots(open, high, low, close, relCandleIndex, &stored)

	//manage positions
	if (*strategy).PosLongSize == 0 && relCandleIndex > 0 { //no long pos
		//enter if current PL higher than previous
		if foundPL {
			if len(stored.PivotLows)-2 >= 0 {
				currentPL := low[stored.PivotLows[len(stored.PivotLows)-1]]
				prevPL := low[stored.PivotLows[len(stored.PivotLows)-2]]
				// fmt.Printf(colorCyan+"currentPL = %v (%v), prevPL = %v (%v)\n"+colorReset, currentPL, stored.PivotLows[len(stored.PivotLows)-1], prevPL, stored.PivotLows[len(stored.PivotLows)-2])
				if currentPL > prevPL {
					// fmt.Printf("Buying at %v\n", close[relCandleIndex-1])
					entryPrice := close[relCandleIndex-1]
					stored.LongEntryPrice = entryPrice
					slPrice := prevPL
					stored.LongSLPrice = slPrice
					rawRiskPerc := (entryPrice - slPrice) / entryPrice
					accRiskedCap := (risk / 100) * float64(accSz)
					posCap := (accRiskedCap / rawRiskPerc) / float64(lev)
					if posCap > strategy.availableEquity {
						posCap = strategy.availableEquity
					}
					posSize := posCap / entryPrice

					strategy.Buy(close[relCandleIndex], slPrice, posSize, true, relCandleIndex)
					// newLabels["middle"] = map[int]string{
					// 	0: fmt.Sprintf("%v|SL %v, TP %v", relCandleIndex, slPrice, ((1 + (tpPerc / 100)) * stored.LongEntryPrice)),
					// }
				}
			}
		}
	} else if strategy.PosLongSize > 0 && relCandleIndex > 0 { //long pos open
		tpPrice := ((1 + (tpPerc / 100)) * stored.LongEntryPrice)
		if high[relCandleIndex] >= tpPrice {
			strategy.CloseLong(tpPrice, 0, relCandleIndex, "TP")
			stored.LongEntryPrice = 0
			stored.LongSLPrice = 0
			// newLabels["middle"] = map[int]string{
			// 	// pivotBarsBack: fmt.Sprintf("L from %v", relCandleIndex),
			// 	0: "EXIT TRADE " + fmt.Sprint(relCandleIndex),
			// }
		} else {
			if low[relCandleIndex] <= stored.LongSLPrice {
				strategy.CloseLong(stored.LongSLPrice, 0, relCandleIndex, "SL")
				stored.LongEntryPrice = 0
				stored.LongSLPrice = 0
			}
		}
	}

	*storage = stored
	return newLabels
}

func scanPivotTrends(
	candle Candlestick, risk, lev, accSz float64,
	open, high, low, close []float64,
	relCandleIndex int,
	strategy *StrategySimulator,
	storage *interface{}) (map[string]map[int]string, PivotTrendScanDataPoint) {
	//TODO: make pivot scanner
	return nil, PivotTrendScanDataPoint{}
}

func runBacktest(
	risk, lev, accSz float64,
	userID, rid, ticker, period string,
	startTime, endTime time.Time,
	packetSize int,
	userStrat func(Candlestick, float64, float64, float64, []float64, []float64, []float64, []float64, int, *StrategySimulator, *interface{}) map[string]map[int]string,
	packetSender func(string, string, []CandlestickChartData, []ProfitCurveData, []SimulatedTradeData),
) ([]CandlestickChartData, []ProfitCurveData, []SimulatedTradeData) {
	var allCandleData []Candlestick
	var chunksArr []*[]Candlestick

	//fetch all candle data concurrently
	concFetchCandleData(startTime, endTime, period, ticker, packetSize, &chunksArr)

	//wait for all candle data fetch complete before running strategy
	for {
		allChunksFilled := true
		for _, e := range chunksArr {
			if len(*e) <= 0 {
				allChunksFilled = false
				break
			}
		}
		if allChunksFilled {
			break
		}
	}

	for _, e := range chunksArr {
		allCandleData = append(allCandleData, *e...)
		// progressBar(userID, rid, len(allCandleData), startTime, endTime)
	}

	//run strat on all candles in chunk, stream each chunk to client
	retCandles, retProfitCurve, retSimTrades := computeBacktest(allCandleData, risk, lev, accSz, packetSize, userID, rid, startTime, endTime, userStrat, packetSender)

	fmt.Println(colorGreen + "\n!!! Backtest complete!" + colorReset)
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
	relIndex := 0
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
			retCandles, pcData, simTradeData = saveDisplayData(retCandles, &(retProfitCurve[0].Data), candle, strategySim, relIndex, labels)
			if pcData.Equity > 0 {
				retProfitCurve[0].Data = append(retProfitCurve[0].Data, pcData)
			}
			if simTradeData.DateTime != "" {
				retSimTrades[0].Data = append(retSimTrades[0].Data, simTradeData)
			}

			//absolute index from absolute start of computation period
			relIndex++
		}

		// progressBar(userID, rid, retCandles, startTime, endTime)

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
