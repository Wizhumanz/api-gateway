package main

import (
	"fmt"
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

	newLabels, foundPL := findPivots(open, high, low, close, relCandleIndex, &(stored.PivotHighs), &(stored.PivotLows))

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

type PivotTrendScanDataPoint struct {
	EntryTime                string      `json:"EntryTime"`
	EntryFirstPivotIndex     int         `json:"EntryFirstPivotIndex"`
	EntrySecondPivotIndex    int         `json:"EntrySecondPivotIndex"`
	EntryPivotsPriceDiffPerc float64     `json:"EntryPivotsPriceDiffPerc"`
	EntryTradeOpenCandle     Candlestick `json:"EntryTradeOpenCandle"`
	ExtentTime               string      `json:"ExtentTime"`
	Duration                 float64     `json:"Duration"`
	Growth                   float64     `json:"Growth"`
}

type PivotTrendScanStore struct {
	PivotHighs     []int
	PivotLows      []int
	MinSearchIndex int
	CurrentPoint   PivotTrendScanDataPoint
	WatchingTrend  bool
}

func breakTrend(candles []Candlestick, breakIndex, relCandleIndex int, high, close []float64, newLabels *(map[string]map[int]string), retData *PivotTrendScanDataPoint, stored *PivotTrendScanStore) {
	(*newLabels)["bottom"] = map[int]string{
		relCandleIndex - breakIndex: "X",
	}

	//find highest point between second entry pivot and trend break
	trendExtentIndex := retData.EntrySecondPivotIndex
	for i := retData.EntrySecondPivotIndex + 1; i <= relCandleIndex; i++ {
		if high[i] > high[trendExtentIndex] {
			trendExtentIndex = i
		}
	}
	(*newLabels)["middle"] = map[int]string{
		relCandleIndex - trendExtentIndex: "^",
	}
	retData.ExtentTime = candles[trendExtentIndex].DateTime

	(*retData).Growth = ((high[breakIndex] - retData.EntryTradeOpenCandle.Close) / retData.EntryTradeOpenCandle.Close) * 100

	entryTime, _ := time.Parse(httpTimeFormat, retData.EntryTime)
	trendEndTime, _ := time.Parse(httpTimeFormat, candles[breakIndex].DateTime)
	retData.Duration = trendEndTime.Sub(entryTime).Minutes()

	//reset
	(*stored).WatchingTrend = false
	(*stored).CurrentPoint = PivotTrendScanDataPoint{}
	(*stored).MinSearchIndex = breakIndex //don't enter with same PL as past trend, must be after break of past trend
}

func contains(sli []int, find int) bool {
	found := false
	for _, e := range sli {
		if e == find {
			found = true
			break
		}
	}
	return found
}

func scanPivotTrends(
	candles []Candlestick,
	open, high, low, close []float64,
	relCandleIndex int,
	storage *interface{}) (map[string]map[int]string, PivotTrendScanDataPoint) {
	exitWatchPivots := 3
	checkTrendFromStartingPivots := false

	stored, ok := (*storage).(PivotTrendScanStore)
	if !ok {
		if relCandleIndex == 0 {
			stored.PivotHighs = []int{}
			stored.PivotLows = []int{}
		} else {
			fmt.Errorf("storage obj assertion fail")
			return nil, PivotTrendScanDataPoint{}
		}
	}
	newLabels, _ := findPivots(open, high, low, close, relCandleIndex, &(stored.PivotHighs), &(stored.PivotLows))
	// newLabels["middle"] = map[int]string{
	// 	0: fmt.Sprintf("%v", relCandleIndex),
	// }

	retData := PivotTrendScanDataPoint{}
	if len(stored.PivotLows) >= 2 {
		if stored.WatchingTrend {
			//manage/watch ongoing trend
			// fmt.Printf(colorYellow+"checking existing trend %v %v\n"+colorReset, relCandleIndex, candles[len(candles)-1].DateTime)
			retData = stored.CurrentPoint

			//check sl
			if low[relCandleIndex] <= low[retData.EntryFirstPivotIndex] {
				breakTrend(candles, relCandleIndex, relCandleIndex, high, close, &newLabels, &retData, &stored)
				fmt.Println(stored.WatchingTrend)
				*storage = stored
				return newLabels, retData
			}

			//check for dynamic number of trend breaks
			type PivotCalc struct {
				Index int
				Type  string //"PL" or "PH"
			}
			var pivotIndexesToCheck []PivotCalc
			//find all pivots since trend start, append to slice in order
			for i := retData.EntryFirstPivotIndex; i < relCandleIndex; i++ {
				addPivot := PivotCalc{}
				if contains(stored.PivotLows, i) {
					addPivot.Index = i
					addPivot.Type = "PL"
				} else if contains(stored.PivotHighs, i) {
					addPivot.Index = i
					addPivot.Type = "PH"
				}

				if addPivot.Index != 0 {
					pivotIndexesToCheck = append(pivotIndexesToCheck, addPivot)
				}
			}

			//check each pivot for trend break
			var trendBreakPivots []PivotCalc
			for j, p := range pivotIndexesToCheck {
				if j > len(pivotIndexesToCheck)-1 {
					break
				}
				//don't check trend's starting pivots
				if j < 2 {
					continue
				}

				//determine pivot type, set vars
				currentPivotIndex := pivotIndexesToCheck[j].Index
				var prevPivotIndex int
				var checkVal []float64
				if contains(stored.PivotHighs, pivotIndexesToCheck[j].Index) {
					checkVal = high
					if checkTrendFromStartingPivots {
						prevPivotIndex = pivotIndexesToCheck[1].Index //use trend's starting high
					} else {
						prevPivotIndex = pivotIndexesToCheck[j-2].Index
					}
				} else {
					checkVal = low
					if checkTrendFromStartingPivots {
						prevPivotIndex = pivotIndexesToCheck[0].Index //use trend's starting high
					} else {
						prevPivotIndex = pivotIndexesToCheck[j-2].Index
					}
				}

				//check if break trend
				if checkVal[prevPivotIndex] > checkVal[currentPivotIndex] {
					//if lower high, record as trend break
					trendBreakPivots = append(trendBreakPivots, p)
					if len(trendBreakPivots) >= exitWatchPivots {
						break
					}
				} else {
					if len(trendBreakPivots) < exitWatchPivots {
						trendBreakPivots = []PivotCalc{} //reset exit watch if not consecutive breaks
					} else {
						break
					}
				}
			}

			//break trend scan if exitWatch sufficient
			if len(trendBreakPivots) >= exitWatchPivots {
				breakTrend(candles, trendBreakPivots[exitWatchPivots-1].Index, relCandleIndex, high, close, &newLabels, &retData, &stored)
			}

			// //search for all pivot highs since entry pivots
			// var checkPHIndexes []int
			// for _, phi := range stored.PivotHighs {
			// 	if phi > retData.EntryFirstPivotIndex {
			// 		checkPHIndexes = append(checkPHIndexes, phi)
			// 	}
			// }
			// var checkPLIndexes []int
			// for _, pli := range stored.PivotLows {
			// 	if pli >= retData.EntryFirstPivotIndex {
			// 		checkPLIndexes = append(checkPLIndexes, pli)
			// 	}
			// }

			// //for each pivot, check if break trend
			// for i := 0; i+1 < len(checkPHIndexes); i++ {
			// 	if high[checkPHIndexes[i]] >= high[checkPHIndexes[i+1]] {
			// 		breakTrend(candles, checkPHIndexes[i+1], relCandleIndex, high, close, &newLabels, &retData, &stored)
			// 		break
			// 	}
			// }
			// for i := 0; i+1 < len(checkPLIndexes); i++ {
			// 	if low[checkPLIndexes[i]] >= low[checkPLIndexes[i+1]] {
			// 		breakTrend(candles, checkPLIndexes[i+1], relCandleIndex, high, close, &newLabels, &retData, &stored)
			// 		break
			// 	}
			// }
		} else {
			// fmt.Printf("finding new trend %v %v\n", relCandleIndex, candles[len(candles)-1].DateTime)

			//find new trend to watch
			latestPLIndex := stored.PivotLows[len(stored.PivotLows)-1]
			latestPL := low[latestPLIndex]
			prevPLIndex := stored.PivotLows[len(stored.PivotLows)-2]
			prevPL := low[prevPLIndex]
			if latestPL > prevPL && latestPLIndex > stored.MinSearchIndex && prevPLIndex > stored.MinSearchIndex {
				retData.EntryTime = candles[latestPLIndex].DateTime
				retData.EntryFirstPivotIndex = prevPLIndex
				retData.EntrySecondPivotIndex = latestPLIndex
				retData.EntryPivotsPriceDiffPerc = ((low[latestPLIndex] - low[prevPLIndex]) / low[prevPLIndex]) * 100
				entryCandle := candles[retData.EntrySecondPivotIndex]
				for i := retData.EntrySecondPivotIndex + 1; i <= relCandleIndex; i++ {
					if candles[i].High > candles[retData.EntrySecondPivotIndex].High && candles[i].Low > candles[retData.EntrySecondPivotIndex].Low {
						entryCandle = candles[i]
						break
					}
				}
				retData.EntryTradeOpenCandle = entryCandle

				stored.CurrentPoint = retData
				stored.WatchingTrend = true

				newLabels["middle"] = map[int]string{
					relCandleIndex - latestPLIndex: "L2",
				}
			}
		}
	}

	*storage = stored
	return newLabels, retData
}
