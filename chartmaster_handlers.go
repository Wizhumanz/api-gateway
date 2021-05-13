package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"cloud.google.com/go/storage"
	"github.com/gorilla/mux"
)

// func indexChartmasterHandler(w http.ResponseWriter, r *http.Request) {
// 	// var retData []CandlestickChartData

// 	setupCORS(&w, r)
// 	if (*r).Method == "OPTIONS" {
// 		return
// 	}

// 	if flag.Lookup("test.v") != nil {
// 		initDatastore()
// 	}

// 	//filter by time
// 	var finalRet []CandlestickChartData
// 	start, err := time.Parse(httpTimeFormat, r.URL.Query()["time_start"][0])
// 	if err != nil {
// 		fmt.Println(err)
// 	}
// 	end, _ := time.Parse(httpTimeFormat, r.URL.Query()["time_end"][0])
// 	for _, c := range candleDisplay {
// 		cTime, err2 := time.Parse(httpTimeFormat, c.DateTime)
// 		if err2 != nil {
// 			fmt.Println(err)
// 		}
// 		if (cTime.After(start) || cTime == start) && (cTime.Before(end) || cTime == start) {
// 			finalRet = append(finalRet, c)
// 		} else if cTime.After(end) {
// 			break
// 		}
// 	}

// 	// return
// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(http.StatusOK)
// 	json.NewEncoder(w).Encode(finalRet)
// }

// func profitCurveHandler(w http.ResponseWriter, r *http.Request) {
// 	setupCORS(&w, r)
// 	if (*r).Method == "OPTIONS" {
// 		return
// 	}

// 	if flag.Lookup("test.v") != nil {
// 		initDatastore()
// 	}

// 	// return
// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(http.StatusOK)
// 	json.NewEncoder(w).Encode(profitCurveDisplay)
// }

// func simulatedTradesHandler(w http.ResponseWriter, r *http.Request) {
// 	var retData []SimulatedTradeData

// 	setupCORS(&w, r)
// 	if (*r).Method == "OPTIONS" {
// 		return
// 	}

// 	if flag.Lookup("test.v") != nil {
// 		initDatastore()
// 	}

// 	//generate random trade entry data
// 	minPrice := 500000  //divided by 100
// 	maxPrice := 900000  //divided by 100
// 	minChange := -5000  //divided by 100
// 	maxChange := 5000   //divided by 100
// 	minSize := 5        //divided by 1000
// 	maxSize := 400      //divided by 1000
// 	minRawProfit := -25 //divided by 10
// 	maxRawProfit := 40  //divided by 10
// 	minPeriodChange := 0
// 	maxPeriodChange := 4
// 	for j := 0; j < 3; j++ {
// 		startDate := time.Date(2021, time.January, 1, 0, 0, 0, 0, time.Now().UTC().Location())
// 		retData = append(retData, SimulatedTradeData{
// 			Label: fmt.Sprintf("Param %v", j+1),
// 			Data:  []SimulatedTradeDataPoint{},
// 		})

// 		for i := 0; i < 40; i++ {
// 			var new SimulatedTradeDataPoint

// 			//randomize trade data
// 			new.Direction = "LONG"
// 			new.EntryPrice = (float64(rand.Intn(maxPrice-minPrice+1)+minPrice) / 100)
// 			new.ExitPrice = new.EntryPrice + (float64(rand.Intn(maxChange-minChange+1)+minChange) / 100)
// 			new.PosSize = (float64(rand.Intn(maxSize-minSize+1)+minSize) / 1000)
// 			new.RiskedEquity = (float64(rand.Intn(maxPrice-minPrice+1)+minPrice) / 100) / 5
// 			new.RawProfitPerc = (float64(rand.Intn(maxRawProfit-minRawProfit+1)+minRawProfit) / 10)

// 			new.DateTime = startDate.Format("2006-01-02")

// 			//randomize period skip
// 			randSkip := (rand.Intn(maxPeriodChange-minPeriodChange+1) + minPeriodChange)
// 			i = i + randSkip

// 			startDate = startDate.AddDate(0, 0, randSkip+1)
// 			retData[j].Data = append(retData[j].Data, new)
// 		}
// 	}

// 	// return
// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(http.StatusOK)
// 	json.NewEncoder(w).Encode(simTradeDisplay)
// }

func backtestHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.URL)

	//create result ID for websocket packets + res storage
	rid := fmt.Sprintf("%v", time.Now().UnixNano())

	setupCORS(&w, r)
	if (*r).Method == "OPTIONS" {
		return
	}

	//get backtest res
	userID := r.URL.Query()["user"][0]
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
	candles, profitCurve, simTrades := runBacktest(strat1, ticker, period, start, end)

	bucketName := "res-" + userID
	go saveBacktestRes(candles, profitCurve, simTrades, rid, bucketName)

	//send display data on ws stream
	ws := wsConnectionsChartmaster[userID]
	if ws != nil {
		pc, _ := json.Marshal(profitCurve)
		ws.WriteMessage(1, pc)
		st, _ := json.Marshal(simTrades)
		ws.WriteMessage(1, st)

		half := len(candles) / 2
		h1 := WebsocketCandlestickPacket{
			ResultID: rid,
			Data:     candles[half:],
		}
		c1, _ := json.Marshal(h1)
		ws.WriteMessage(1, c1)
		h2 := WebsocketCandlestickPacket{
			ResultID: rid,
			Data:     candles[:half],
		}
		c2, _ := json.Marshal(h2)
		ws.WriteMessage(1, c2)
	}

	// return
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	// json.NewEncoder(w).Encode(finalRet)
}

func getTickersHandler(w http.ResponseWriter, r *http.Request) {
	setupCORS(&w, r)
	if (*r).Method == "OPTIONS" {
		return
	}

	data, err := ioutil.ReadFile("./json-data/symbols-binance-fut-perp.json")
	if err != nil {
		fmt.Print(err)
	}

	var t []CoinAPITicker
	json.Unmarshal(data, &t)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(t)
}

func getBacktestHistoryHandler(w http.ResponseWriter, r *http.Request) {
	setupCORS(&w, r)
	if (*r).Method == "OPTIONS" {
		return
	}

	userID := r.URL.Query()["user"][0]
	bucketData := listFiles("res-" + userID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(bucketData)
}

func getBacktestResHandler(w http.ResponseWriter, r *http.Request) {
	setupCORS(&w, r)
	if (*r).Method == "OPTIONS" {
		return
	}

	//get backtest hist file
	storageClient, _ := storage.NewClient(ctx)
	defer storageClient.Close()
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	userID := r.URL.Query()["user"][0]
	bucketName := "res-" + userID
	backtestResID, _ := url.QueryUnescape(mux.Vars(r)["id"])
	objName := backtestResID + ".json"
	rc, _ := storageClient.Bucket(bucketName).Object(objName).NewReader(ctx)
	defer rc.Close()

	backtestResByteArr, _ := ioutil.ReadAll(rc)
	// fmt.Println(string(backtestResByteArr))
	var rawRes BacktestResFile
	json.Unmarshal(backtestResByteArr, &rawRes)

	//rehydrate backtest results
	// candles, profitCurve, simTrades := completeBacktestResFile(rawRes)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(rawRes)
}
