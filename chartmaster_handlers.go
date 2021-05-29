package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/datastore"
	"cloud.google.com/go/storage"
	"github.com/gorilla/mux"
	"google.golang.org/api/iterator"
)

func backtestHandler(w http.ResponseWriter, r *http.Request) {
	//create result ID for websocket packets + res storage
	rid := fmt.Sprintf("%v", time.Now().UnixNano())

	setupCORS(&w, r)
	if (*r).Method == "OPTIONS" {
		return
	}

	var backtest Backtest
	err := json.NewDecoder(r.Body).Decode(&backtest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	//get backtest res
	userID := backtest.User
	ticker := backtest.Ticker
	period := backtest.Period
	risk := backtest.Risk
	leverage := backtest.Leverage
	size := backtest.Size

	candlePacketSize, err := strconv.Atoi(backtest.CandlePacketSize)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	start, err := time.Parse(httpTimeFormat, backtest.TimeStart)
	if err != nil {
		fmt.Println(err)
	}
	end, err2 := time.Parse(httpTimeFormat, backtest.TimeEnd)
	if err2 != nil {
		fmt.Println(err)
	}

	//strat params
	rF, _ := strconv.ParseFloat(risk, 32)
	lF, _ := strconv.ParseFloat(leverage, 32)
	szF, _ := strconv.ParseFloat(size, 32)

	var candles []CandlestickChartData
	var profitCurve []ProfitCurveData
	var simTrades []SimulatedTradeData
	candles, profitCurve, simTrades = runBacktest(rF, lF, szF, scan1, userID, rid, ticker, period, start, end, candlePacketSize, streamBacktestResData)

	// Get all of user's shared history json data
	shareResult := getAllShareResult(userID)

	// Delete an element in a bucket if len greater than 10
	bucketName := "res-" + userID
	bucketData := listFiles(bucketName)
	if len(bucketData) >= 10+len(shareResult) {
		for i, file := range bucketData {
			// fmt.Println(file)
			// fmt.Println(shareResult)
			// fmt.Println(contains(shareResult, strings.Split(file, ".")[0]))
			if !containsString(shareResult, strings.Split(file, ".")[0]) {
				deleteFile(bucketName, bucketData[i])
				break
			}
		}
	}

	//save result to bucket
	go saveBacktestRes(candles, profitCurve, simTrades, rid, bucketName, ticker, period, backtest.TimeStart, backtest.TimeEnd)

	// return
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	// json.NewEncoder(w).Encode(finalRet)
}

func getAllShareResult(userID string) []string {
	// Get all of user's shared history json data
	var shareResult []string
	query := datastore.NewQuery("ShareResult").Filter("UserID =", userID)
	t := client.Run(ctx, query)
	for {
		var x ShareResult
		_, err := t.Next(&x)
		if err == iterator.Done {
			break
		}
		shareResult = append(shareResult, x.ResultFileName)
	}
	return shareResult
}

func getAllShareResultHandler(w http.ResponseWriter, r *http.Request) {
	setupCORS(&w, r)
	if (*r).Method == "OPTIONS" {
		return
	}

	if flag.Lookup("test.v") != nil {
		initDatastore()
	}
	userID := r.URL.Query()["user"][0]
	shareResult := getAllShareResult(userID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(shareResult)
}

func shareResultHandler(w http.ResponseWriter, r *http.Request) {
	setupCORS(&w, r)
	if (*r).Method == "OPTIONS" {
		return
	}

	if flag.Lookup("test.v") != nil {
		initDatastore()
	}

	uniqueURL := fmt.Sprintf("%v", time.Now().UnixNano()) + generateRandomID(20)
	fmt.Println(uniqueURL)

	var share ShareResult
	err := json.NewDecoder(r.Body).Decode(&share)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// add new row to DB
	share.ShareID = uniqueURL
	kind := "ShareResult"
	newKey := datastore.IncompleteKey(kind, nil)
	if _, err := client.Put(ctx, newKey, &share); err != nil {
		log.Fatalf("Failed to delete Bot: %v", err)
	}

	fmt.Println(share)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(share)
}

func getShareResultHandler(w http.ResponseWriter, r *http.Request) {
	setupCORS(&w, r)
	if (*r).Method == "OPTIONS" {
		return
	}

	var shareResult ShareResult

	shareID := r.URL.Query()["share"][0]
	fmt.Println(shareID)
	query := datastore.NewQuery("ShareResult").Filter("ShareID =", shareID)
	t := client.Run(ctx, query)
	_, error := t.Next(&shareResult)
	if error != nil {
		fmt.Println(error.Error())
	}

	// candlePacketSize, err := strconv.Atoi(r.URL.Query()["candlePacketSize"][0])
	// if err != nil {
	// 	fmt.Println(err)
	// 	w.WriteHeader(http.StatusBadRequest)
	// 	return
	// }

	candlePacketSize := 100

	//create result ID for websocket packets + res storage
	rid := fmt.Sprintf("%v", time.Now().UnixNano())

	//get backtest hist file
	storageClient, _ := storage.NewClient(ctx)
	defer storageClient.Close()
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	userID := shareResult.UserID
	bucketName := "res-" + userID
	backtestResID := shareResult.ResultFileName
	objName := backtestResID + ".json"
	rc, _ := storageClient.Bucket(bucketName).Object(objName).NewReader(ctx)
	defer rc.Close()

	backtestResByteArr, _ := ioutil.ReadAll(rc)
	var rawRes BacktestResFile
	json.Unmarshal(backtestResByteArr, &rawRes)

	//rehydrate backtest results
	candles, profitCurve, simTrades := completeBacktestResFile(rawRes, userID, rid, candlePacketSize, streamBacktestResData)
	ret := BacktestResFile{
		Ticker:               rawRes.Ticker,
		Period:               rawRes.Period,
		Start:                rawRes.Start,
		End:                  rawRes.End,
		ModifiedCandlesticks: candles,
		ProfitCurve:          profitCurve,
		SimulatedTrades:      simTrades,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ret)
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
	var rawRes BacktestResFile
	json.Unmarshal(backtestResByteArr, &rawRes)

	//rehydrate backtest results
	candles, profitCurve, simTrades := completeBacktestResFile(rawRes, userID, rid, candlePacketSize, streamBacktestResData)
	ret := BacktestResFile{
		Ticker:               rawRes.Ticker,
		Period:               rawRes.Period,
		Start:                rawRes.Start,
		End:                  rawRes.End,
		ModifiedCandlesticks: candles,
		ProfitCurve:          profitCurve,
		SimulatedTrades:      simTrades,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ret)
}
