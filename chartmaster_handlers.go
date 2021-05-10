package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

func indexChartmasterHandler(w http.ResponseWriter, r *http.Request) {
	// var retData []CandlestickChartData

	setupCORS(&w, r)
	if (*r).Method == "OPTIONS" {
		return
	}

	if flag.Lookup("test.v") != nil {
		initDatastore()
	}

	//filter by time
	var finalRet []CandlestickChartData
	start, err := time.Parse(httpTimeFormat, r.URL.Query()["time_start"][0])
	if err != nil {
		fmt.Println(err)
	}
	end, _ := time.Parse(httpTimeFormat, r.URL.Query()["time_end"][0])
	for _, c := range candleDisplay {
		cTime, err2 := time.Parse(httpTimeFormat, c.DateTime)
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
	setupCORS(&w, r)
	if (*r).Method == "OPTIONS" {
		return
	}

	if flag.Lookup("test.v") != nil {
		initDatastore()
	}

	// return
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(profitCurveDisplay)
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

			new.DateTime = startDate.Format("2006-01-02")

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
	json.NewEncoder(w).Encode(simTradeDisplay)
}

func backtestHandler(w http.ResponseWriter, r *http.Request) {
	setupCORS(&w, r)
	if (*r).Method == "OPTIONS" {
		return
	}

	ticker := r.URL.Query()["ticker"][0]
	period := r.URL.Query()["period"][0]
	start, err := time.Parse(httpTimeFormat, r.URL.Query()["time_start"][0])
	if err != nil {
		fmt.Println(err)
	}
	end, err2 := time.Parse(httpTimeFormat, r.URL.Query()["time_end"][0])
	if err2 != nil {
		fmt.Println(err)
	}
	runBacktest(strat1, ticker, period, start, end)

	// return
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	// json.NewEncoder(w).Encode(finalRet)
}
