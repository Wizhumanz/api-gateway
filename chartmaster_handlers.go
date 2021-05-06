package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"time"
)

func indexChartmasterHandler(w http.ResponseWriter, r *http.Request) {
	var retData []CandlestickChartData

	setupCORS(&w, r)
	if (*r).Method == "OPTIONS" {
		return
	}

	if flag.Lookup("test.v") != nil {
		initDatastore()
	}

	//generate random OHLC data
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
				DateTime: startDate.Format("2006-01-02T15:04:05"),
				Open:     retData[len(retData)-1].Close,
			}
		} else {
			new = CandlestickChartData{
				DateTime: startDate.Format("2006-01-02T15:04:05"),
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

	//filter by time
	var finalRet []CandlestickChartData
	format := "2006-01-02T15:04:05"
	start, err := time.Parse(format, r.URL.Query()["time_start"][0])
	if err != nil {
		fmt.Println(err)
	}
	end, _ := time.Parse(format, r.URL.Query()["time_end"][0])
	for _, c := range retData {
		cTime, err2 := time.Parse(format, c.DateTime)
		if err2 != nil {
			fmt.Println(err)
		}
		if (cTime.After(start) || cTime == start) && (cTime.Before(end) || cTime == start) {
			finalRet = append(finalRet, c)
		} else if cTime.After(end) {
			break
		}
	}

	// return
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(finalRet)
}

func profitCurveHandler(w http.ResponseWriter, r *http.Request) {
	var retData []ProfitCurveData

	setupCORS(&w, r)
	if (*r).Method == "OPTIONS" {
		return
	}

	if flag.Lookup("test.v") != nil {
		initDatastore()
	}

	//generate random profit data
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

			new.Date = startDate.Format("2006-01-02")

			//randomize period skip
			randSkip := (rand.Intn(maxPeriodChange-minPeriodChange+1) + minPeriodChange)
			i = i + randSkip

			startDate = startDate.AddDate(0, 0, randSkip+1)
			retData[j].Data = append(retData[j].Data, new)
		}
	}

	// return
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(retData)
}

func simulatedTradesHandler(w http.ResponseWriter, r *http.Request) {
	var retData []SimulatedTradeData

	setupCORS(&w, r)
	if (*r).Method == "OPTIONS" {
		return
	}

	if flag.Lookup("test.v") != nil {
		initDatastore()
	}

	//generate random trade entry data
	minPrice := 500000  //divided by 100
	maxPrice := 900000  //divided by 100
	minChange := -5000  //divided by 100
	maxChange := 5000   //divided by 100
	minSize := 5        //divided by 1000
	maxSize := 400      //divided by 1000
	minRawProfit := -25 //divided by 10
	maxRawProfit := 40  //divided by 10
	minPeriodChange := 0
	maxPeriodChange := 4
	for j := 0; j < 3; j++ {
		startDate := time.Date(2021, time.January, 1, 0, 0, 0, 0, time.Now().UTC().Location())
		retData = append(retData, SimulatedTradeData{
			Label: fmt.Sprintf("Param %v", j+1),
			Data:  []SimulatedTradeDataPoint{},
		})

		for i := 0; i < 40; i++ {
			var new SimulatedTradeDataPoint

			//randomize trade data
			new.Direction = "LONG"
			new.EntryPrice = (float64(rand.Intn(maxPrice-minPrice+1)+minPrice) / 100)
			new.ExitPrice = new.EntryPrice + (float64(rand.Intn(maxChange-minChange+1)+minChange) / 100)
			new.PosSize = (float64(rand.Intn(maxSize-minSize+1)+minSize) / 1000)
			new.RiskedEquity = (float64(rand.Intn(maxPrice-minPrice+1)+minPrice) / 100) / 5
			new.RawProfitPerc = (float64(rand.Intn(maxRawProfit-minRawProfit+1)+minRawProfit) / 10)

			new.Date = startDate.Format("2006-01-02")

			//randomize period skip
			randSkip := (rand.Intn(maxPeriodChange-minPeriodChange+1) + minPeriodChange)
			i = i + randSkip

			startDate = startDate.AddDate(0, 0, randSkip+1)
			retData[j].Data = append(retData[j].Data, new)
		}
	}

	// return
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(retData)
}
