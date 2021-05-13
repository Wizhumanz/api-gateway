package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"cloud.google.com/go/storage"
	"github.com/gorilla/mux"
)

func backtestHandler(w http.ResponseWriter, r *http.Request) {
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
	candlePacketSize, err := strconv.Atoi(r.URL.Query()["candlePacketSize"][0])
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	start, err := time.Parse(httpTimeFormat, r.URL.Query()["time_start"][0])
	if err != nil {
		fmt.Println(err)
	}
	end, err2 := time.Parse(httpTimeFormat, r.URL.Query()["time_end"][0])
	if err2 != nil {
		fmt.Println(err)
	}
	candles, profitCurve, simTrades := runBacktest(strat1, ticker, period, start, end, candlePacketSize, func(c []CandlestickChartData) {
		ws := wsConnectionsChartmaster[userID]
		if ws != nil {
			var pushCandles []CandlestickChartData
			for _, candle := range c {
				if candle.DateTime == "" {
					fmt.Println(candle)
				} else {
					pushCandles = append(pushCandles, candle)
				}
			}
			streamCandlesData(ws, pushCandles, rid)
		}
	})

	//save result to bucket
	bucketName := "res-" + userID
	go saveBacktestRes(candles, profitCurve, simTrades, rid, bucketName, ticker, period, r.URL.Query()["time_start"][0], r.URL.Query()["time_end"][0])

	// go streamBacktestData(userID, rid, candlePacketSize, candles, profitCurve, simTrades)

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

	candlePacketSize, err := strconv.Atoi(r.URL.Query()["candlePacketSize"][0])
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//create result ID for websocket packets + res storage
	rid := fmt.Sprintf("%v", time.Now().UnixNano())

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
	candles, profitCurve, simTrades := completeBacktestResFile(rawRes)
	ret := BacktestResFile{
		ModifiedCandlesticks: candles,
		ProfitCurve:          profitCurve,
		SimulatedTrades:      simTrades,
	}

	go streamBacktestData(userID, rid, candlePacketSize, candles, profitCurve, simTrades)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ret)
}
