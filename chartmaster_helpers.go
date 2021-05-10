package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

func getCandleData(ticker, period string, start, end time.Time) []Candlestick {
	//send request
	format := "2006-01-02T15:04:05"
	base := "https://rest.coinapi.io/v1/ohlcv/BINANCEFTS_PERP_BTC_USDT/history" //TODO: build dynamically based on ticker
	full := fmt.Sprintf("%s?period_id=%s&time_start=%s&time_end=%s&limit=100000",
		base,
		period,
		start.Format(format),
		end.Format(format))
	req, _ := http.NewRequest("GET", full, nil)
	req.Header.Add("X-CoinAPI-Key", "4D684039-406E-451F-BB2B-6BDC123808E1")
	client := &http.Client{}
	response, err := client.Do(req)

	if err != nil {
		log.Fatalf("An Error Occured %v", err)
	} else {
		body, _ := ioutil.ReadAll(response.Body)
		log.Println(string(body))
	}

	//parse data

	return []Candlestick{}
}

func saveJsonToRedis() {
	data, err := ioutil.ReadFile("./mar-apr2021.json")
	if err != nil {
		fmt.Print(err)
	}

	var jStruct []RawOHLCGetResp
	json.Unmarshal(data, &jStruct)
	//progress indicator
	indicatorParts := 5
	totalLen := len(jStruct)
	lenPart := totalLen / indicatorParts
	for i, c := range jStruct {
		// fmt.Println(c)
		ctx := context.Background()
		key := "BTCUSDT:1MIN:" + c.PeriodStart
		rdb.HMSet(ctx, key, "open", c.Open, "high", c.High, "low", c.Low, "close", c.Close, "volume", c.Volume, "tradesCount", c.TradesCount, "timeOpen", c.TimeOpen, "timeClose", c.TimeClose, "periodStart", c.PeriodStart, "periodEnd", c.PeriodEnd)

		if (i > 1) && ((i % lenPart) == 0) {
			fmt.Printf("Section %v of %v complete\n", (i / lenPart), indicatorParts)
		}
	}
	fmt.Println(colorGreen + "Save json to redis complete!" + colorReset)
}
