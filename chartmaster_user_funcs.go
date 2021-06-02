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
	EntryTime             string  `json:"EntryTime"`
	EntryFirstPivotIndex  int     `json:"EntryFirstPivotIndex"`
	EntrySecondPivotIndex int     `json:"EntrySecondPivotIndex"`
	ExtentTime            string  `json:"ExtentTime"`
	Duration              float64 `json:"Duration"`
	Growth                float64 `json:"Growth"`
}

type PivotTrendScanStore struct {
	PivotHighs    []int
	PivotLows     []int
	CurrentPoint  PivotTrendScanDataPoint
	WatchingTrend bool
}

func scanPivotTrends(
	candles []Candlestick,
	open, high, low, close []float64,
	relCandleIndex int,
	storage *interface{}) (map[string]map[int]string, PivotTrendScanDataPoint) {
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
			retData = stored.CurrentPoint

			//search for all pivot highs since entry pivots
			var checkPHIndexes []int
			for _, phi := range stored.PivotHighs {
				if phi > retData.EntrySecondPivotIndex {
					checkPHIndexes = append(checkPHIndexes, phi)
				}
			}

			//for each high, check if break trend
			for i := 0; i+1 < len(checkPHIndexes); i++ {
				if high[checkPHIndexes[i]] >= high[checkPHIndexes[i+1]] {
					trendBreakIndex := checkPHIndexes[i+1]

					newLabels["middle"] = map[int]string{
						relCandleIndex - trendBreakIndex: "X",
					}

					//find highest point between second entry pivot and trend break
					trendExtentIndex := retData.EntrySecondPivotIndex
					for i := retData.EntrySecondPivotIndex + 1; i <= relCandleIndex; i++ {
						if high[i] > high[trendExtentIndex] {
							trendExtentIndex = i
						}
					}
					newLabels["middle"] = map[int]string{
						relCandleIndex - trendExtentIndex: "^",
					}
					retData.ExtentTime = candles[trendExtentIndex].DateTime

					retData.Growth = (high[trendBreakIndex] - low[retData.EntrySecondPivotIndex]) / low[retData.EntrySecondPivotIndex]

					entryTime, _ := time.Parse(httpTimeFormat, retData.EntryTime)
					trendEndTime, _ := time.Parse(httpTimeFormat, candles[trendBreakIndex].DateTime)
					retData.Duration = trendEndTime.Sub(entryTime).Minutes()

					//reset
					stored.WatchingTrend = false
					stored.CurrentPoint = PivotTrendScanDataPoint{}
					break
				}
			}
		} else {
			//find new trend to watch
			latestPLIndex := stored.PivotLows[len(stored.PivotLows)-1]
			latestPL := low[latestPLIndex]
			prevPLIndex := stored.PivotLows[len(stored.PivotLows)-2]
			prevPL := low[prevPLIndex]
			if latestPL > prevPL {
				retData.EntryTime = candles[latestPLIndex].DateTime
				retData.EntryFirstPivotIndex = prevPLIndex
				retData.EntrySecondPivotIndex = latestPLIndex
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
