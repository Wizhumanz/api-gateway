package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

func cacheCandleData(candles []Candlestick, ticker, period string) {
	fmt.Printf("Adding %v candles to cache %v %v\n", len(candles), ticker, period)

	//progress indicator
	indicatorParts := 30
	totalLen := len(candles)
	if totalLen < indicatorParts {
		indicatorParts = 1
	}
	lenPart := totalLen / indicatorParts
	for i, c := range candles {
		// fmt.Println(c)
		ctx := context.Background()
		key := ticker + ":" + period + ":" + c.PeriodStart
		rdb.HMSet(ctx, key, "open", c.Open, "high", c.High, "low", c.Low, "close", c.Close, "volume", c.Volume, "tradesCount", c.TradesCount, "timeOpen", c.TimeOpen, "timeClose", c.TimeClose, "periodStart", c.PeriodStart, "periodEnd", c.PeriodEnd)

		if (i > 1) && ((i % lenPart) == 0) {
			fmt.Printf("Section %v of %v complete\n", (i / lenPart), indicatorParts)
		}
	}
	fmt.Println(colorGreen + "Save json to redis complete!" + colorReset)
}

func fetchCandleData(ticker, period string, start, end time.Time) []Candlestick {
	fmt.Printf("FETCHING from %v to %v\n", start.Format(httpTimeFormat), end.Format(httpTimeFormat))

	//send request
	base := "https://rest.coinapi.io/v1/ohlcv/BINANCEFTS_PERP_BTC_USDT/history" //TODO: build dynamically based on ticker
	full := fmt.Sprintf("%s?period_id=%s&time_start=%s&time_end=%s",
		base,
		period,
		start.Format(httpTimeFormat),
		end.Format(httpTimeFormat))

	req, _ := http.NewRequest("GET", full, nil)
	req.Header.Add("X-CoinAPI-Key", "A2642A7A-A8C8-48C1-83CE-8D258BD7BBF5")
	client := &http.Client{}
	response, err := client.Do(req)

	if err != nil {
		fmt.Printf("GET candle data err %v\n", err)
		return nil
	}

	//parse data
	body, _ := ioutil.ReadAll(response.Body)
	fmt.Println(string(body))
	var jStruct []Candlestick
	json.Unmarshal(body, &jStruct)
	//save data to cache so don't have to fetch again
	if len(jStruct) > 0 {
		go cacheCandleData(jStruct, ticker, period)
	}

	fmt.Println("Fresh fetch complete")
	return jStruct
}

func getCachedCandleData(ticker, period string, start, end time.Time) []Candlestick {
	fmt.Printf("CACHE getting from %v to %v\n", start.Format(httpTimeFormat), end.Format(httpTimeFormat))

	var retCandles []Candlestick
	checkEnd := end.Add(periodDurationMap[period])
	for cTime := start; cTime.Before(checkEnd); cTime = cTime.Add(periodDurationMap[period]) {
		key := ticker + ":" + period + ":" + cTime.Format(httpTimeFormat) + ".0000000Z"
		cachedData, _ := rdb.HGetAll(ctx, key).Result()

		//if candle not found in cache, fetch new
		if cachedData["open"] == "" {
			//find end time for fetch
			var fetchEndTime time.Time
			calcTime := cTime
			for {
				calcTime = calcTime.Add(periodDurationMap[period])
				key := ticker + ":" + period + ":" + calcTime.Format(httpTimeFormat) + ".0000000Z" //TODO: update for diff period
				cached, _ := rdb.HGetAll(ctx, key).Result()
				//find index where next cache starts again, or break if passed end time of backtest
				if (cached["open"] != "") || (calcTime.After(end)) {
					fetchEndTime = calcTime
					break
				}
			}
			//fetch missing candles
			fetchedCandles := fetchCandleData(ticker, period, cTime, fetchEndTime)
			retCandles = append(retCandles, fetchedCandles...)
			//start getting cache again from last fetch time
			cTime = fetchEndTime.Add(-periodDurationMap[period])
		} else {
			newCandle := Candlestick{}
			newCandle.Create(cachedData)
			retCandles = append(retCandles, newCandle)
		}
	}

	fmt.Println("Cache fetch complete")
	return retCandles
}

// makeBacktestResFile creates backtest result file with passed args and returns the name of the new file.
func makeBacktestResFile(c []CandlestickChartData, p []ProfitCurveData, s []SimulatedTradeData) string {
	data := BacktestResFile{
		ModifiedCandlesticks: c, //TODO: only save modified candlesticks for space saving
		ProfitCurve:          p,
		SimulatedTrades:      s,
	}
	file, _ := json.MarshalIndent(data, "", " ")
	fileName := fmt.Sprintf("%v.json", time.Now().Unix())
	_ = ioutil.WriteFile(fileName, file, 0644)

	return fileName
}

func saveJsonToRedis() {
	data, err := ioutil.ReadFile("./mar-apr2021.json")
	if err != nil {
		fmt.Print(err)
	}

	var jStruct []Candlestick
	json.Unmarshal(data, &jStruct)
	// go cacheCandleData(jStruct, ticker, period)
}

func renameKeys() {
	keys, _ := rdb.Keys(ctx, "*").Result()
	var splitKeys = map[string]string{}
	for _, k := range keys {
		splitKeys[k] = "BINANCEFTS_PERP_BTC_USDT:" + strings.SplitN(k, ":", 2)[1]
	}

	// for k, v := range splitKeys {
	// 	rdb.Rename(ctx, k, v)
	// }
}

func generateRandomCandles() {
	retData := []CandlestickChartData{}
	min := 500000
	max := 900000
	minChange := -40000
	maxChange := 45000
	minWick := 1000
	maxWick := 30000
	startDate := time.Date(2021, time.January, 1, 0, 0, 0, 0, time.Now().UTC().Location())
	for i := 0; i < 250; i++ {
		var new CandlestickChartData

		//body
		if i != 0 {
			startDate = startDate.AddDate(0, 0, 1)
			new = CandlestickChartData{
				DateTime: startDate.Format(httpTimeFormat),
				Open:     retData[len(retData)-1].Close,
			}
		} else {
			new = CandlestickChartData{
				DateTime: startDate.Format(httpTimeFormat),
				Open:     float64(rand.Intn(max-min+1)+min) / 100,
			}
		}
		new.Close = new.Open + (float64(rand.Intn(maxChange-minChange+1)+minChange) / 100)

		//wick
		if new.Close > new.Open {
			new.High = new.Close + (float64(rand.Intn(maxWick-minWick+1)+minWick) / 100)
			new.Low = new.Open - (float64(rand.Intn(maxWick-minWick+1)+minWick) / 100)
		} else {
			new.High = new.Open + (float64(rand.Intn(maxWick-minWick+1)+minWick) / 100)
			new.Low = new.Close - (float64(rand.Intn(maxWick-minWick+1)+minWick) / 100)
		}

		retData = append(retData, new)
	}
}

func generateRandomProfitCurve() {
	retData := []ProfitCurveData{}
	minChange := -110
	maxChange := 150
	minPeriodChange := 0
	maxPeriodChange := 4
	for j := 0; j < 10; j++ {
		startEquity := 1000
		startDate := time.Date(2021, time.January, 1, 0, 0, 0, 0, time.Now().UTC().Location())
		retData = append(retData, ProfitCurveData{
			Label: fmt.Sprintf("Param %v", j+1),
			Data:  []ProfitCurveDataPoint{},
		})

		for i := 0; i < 40; i++ {
			rand.Seed(time.Now().UTC().UnixNano())
			var new ProfitCurveDataPoint

			//randomize equity change
			if i == 0 {
				new.Equity = float64(startEquity)
			} else {
				change := float64(rand.Intn(maxChange-minChange+1) + minChange)
				latestIndex := len(retData[j].Data) - 1
				new.Equity = math.Abs(retData[j].Data[latestIndex].Equity + change)
			}

			new.DateTime = startDate.Format("2006-01-02")

			//randomize period skip
			randSkip := (rand.Intn(maxPeriodChange-minPeriodChange+1) + minPeriodChange)
			i = i + randSkip

			startDate = startDate.AddDate(0, 0, randSkip+1)
			retData[j].Data = append(retData[j].Data, new)
		}
	}
}
