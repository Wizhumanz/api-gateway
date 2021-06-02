package main

import "fmt"

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

type PivotTrendScanDataPoint struct {
	EntryTime  string  `json:"EntryTime"`
	ExtentTime string  `json:"ExtentTime"`
	Duration   int     `json:"Duration"`
	Growth     float64 `json:"Growth"`
}

func scanPivotTrends(
	candles []Candlestick,
	open, high, low, close []float64,
	relCandleIndex int,
	storage *interface{}) (map[string]map[int]string, PivotTrendScanDataPoint) {
	stored, ok := (*storage).(PivotsStore)
	if !ok {
		if relCandleIndex == 0 {
			stored.PivotHighs = []int{}
			stored.PivotLows = []int{}
		} else {
			fmt.Errorf("storage obj assertion fail")
			return nil, PivotTrendScanDataPoint{}
		}
	}
	newLabels, _ := findPivots(open, high, low, close, relCandleIndex, &stored)
	// newLabels["middle"] = map[int]string{
	// 	0: fmt.Sprintf("%v", relCandleIndex),
	// }

	//TODO: make pivot scanner
	retData := PivotTrendScanDataPoint{}
	if relCandleIndex%10 == 0 {
		retData.Growth = 69
	}

	*storage = stored
	return newLabels, retData
}
