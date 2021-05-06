package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
)

func getData() {
	ctx := context.Background()
	res, _ := rdb.HGetAll(ctx, "BTCUSDT:1MIN:2021-03-03T00:47:00.0000000Z").Result()
	fmt.Println(res["open"])
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
