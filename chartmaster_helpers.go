package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/gorilla/websocket"
	"google.golang.org/api/iterator"
)

func copyObjs(base []Candlestick, copyer func(Candlestick) CandlestickChartData) []CandlestickChartData {
	var ret []CandlestickChartData
	for _, obj := range base {
		ret = append(ret, copyer(obj))
	}
	return ret
}

func cacheCandleData(candles []Candlestick, ticker, period string) {
	// fmt.Printf("Adding %v candles to cache %v %v\n", len(candles), ticker, period)
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
		rdbChartmaster.HMSet(ctx, key, "open", c.Open, "high", c.High, "low", c.Low, "close", c.Close, "volume", c.Volume, "tradesCount", c.TradesCount, "timeOpen", c.TimeOpen, "timeClose", c.TimeClose, "periodStart", c.PeriodStart, "periodEnd", c.PeriodEnd)

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
		cachedData, _ := rdbChartmaster.HGetAll(ctx, key).Result()

		//if candle not found in cache, fetch new
		if cachedData["open"] == "" {
			//find end time for fetch
			var fetchEndTime time.Time
			calcTime := cTime
			for {
				calcTime = calcTime.Add(periodDurationMap[period])
				key := ticker + ":" + period + ":" + calcTime.Format(httpTimeFormat) + ".0000000Z" //TODO: update for diff period
				cached, _ := rdbChartmaster.HGetAll(ctx, key).Result()
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

func saveDisplayData(cArr []CandlestickChartData, c Candlestick, strat StrategySimulator, relIndex int, label string, labelBB int, profitCurveSoFar []ProfitCurveDataPoint) ([]CandlestickChartData, ProfitCurveDataPoint, SimulatedTradeDataPoint) {
	//candlestick
	retCandlesArr := cArr
	newCandleD := CandlestickChartData{
		DateTime: c.DateTime,
		Open:     c.Open,
		High:     c.High,
		Low:      c.Low,
		Close:    c.Close,
	}
	//strategy enter/exit
	if strat.Actions[relIndex].Action == "ENTER" {
		newCandleD.StratEnterPrice = strat.Actions[relIndex].Price
	} else if strat.Actions[relIndex].Action == "SL" {
		newCandleD.StratExitPrice = strat.Actions[relIndex].Price
	}
	retCandlesArr = append(retCandlesArr, newCandleD)
	//candle label
	if label != "" {
		index := len(cArr) - labelBB
		cArr[index].Label = label
	}

	//profit curve
	var pd ProfitCurveDataPoint
	//only add data point if changed from last point OR 1st or 2nd datapoint
	if (relIndex == 0) || (strat.GetEquity() != profitCurveSoFar[len(profitCurveSoFar)-1].Equity) {
		pd = ProfitCurveDataPoint{
			DateTime: c.DateTime,
			Equity:   strat.GetEquity(),
		}
	}

	//sim trades
	sd := SimulatedTradeDataPoint{}
	if strat.Actions[relIndex].Action == "SL" || strat.Actions[relIndex].Action == "TP" {
		//find entry conditions
		var entryPrice float64
		var size float64
		for i := 1; i < len(strat.Actions)-1; i++ {
			current := strat.Actions[relIndex-i]
			if current.Action == "ENTER" {
				entryPrice = current.Price
				size = current.PosSize
			}
		}

		sd.DateTime = c.DateTime
		sd.Direction = "LONG" //TODO: fix later when strategy changes
		sd.EntryPrice = entryPrice
		sd.ExitPrice = strat.Actions[relIndex].Price
		sd.PosSize = size
		sd.RiskedEquity = size * entryPrice
		sd.RawProfitPerc = ((sd.ExitPrice - sd.EntryPrice) / sd.EntryPrice) * 100
	}

	return retCandlesArr, pd, sd
}

func streamPacket(ws *websocket.Conn, chartData []interface{}, resID string) {
	packet := WebsocketPacket{
		ResultID: resID,
		Data:     chartData,
	}
	data, _ := json.Marshal(packet)
	ws.WriteMessage(1, data)
}

func streamBacktestResData(userID, rid string, c []CandlestickChartData, pc []ProfitCurveData, st []SimulatedTradeData) {
	ws := wsConnectionsChartmaster[userID]
	if ws != nil {
		//profit curve
		if len(pc) > 0 {
			var pcStreamData []interface{}
			for _, pCurve := range pc {
				pcStreamData = append(pcStreamData, pCurve)
			}
			streamPacket(ws, pcStreamData, rid)
		}

		//sim trades
		if len(st) > 0 {
			var stStreamData []interface{}
			for _, trade := range st {
				stStreamData = append(stStreamData, trade)
			}
			streamPacket(ws, stStreamData, rid)
		}

		//candlesticks
		var pushCandles []CandlestickChartData
		for _, candle := range c {
			if candle.DateTime == "" {

			} else {
				pushCandles = append(pushCandles, candle)
			}
		}
		var cStreamData []interface{}
		for _, can := range pushCandles {
			cStreamData = append(cStreamData, can)
		}
		streamPacket(ws, cStreamData, rid)
	}
}

// makeBacktestResFile creates backtest result file with passed args and returns the name of the new file.
func makeBacktestResFile(c []CandlestickChartData, p []ProfitCurveData, s []SimulatedTradeData, ticker, period, start, end string) string {
	//only save candlesticks which are modified
	saveCandles := []CandlestickChartData{}
	for i, candle := range c {
		//only save first or last candles, and candles with entry/exit/label
		if ((candle.StratEnterPrice != 0) || (candle.StratExitPrice != 0) || (candle.Label != "")) || ((i == 0) || (i == len(c)-1)) {
			saveCandles = append(saveCandles, candle)
		}
	}

	data := BacktestResFile{
		Ticker:               ticker,
		Period:               period,
		Start:                start,
		End:                  end,
		ModifiedCandlesticks: saveCandles,
		ProfitCurve:          p, //optimize for when equity doesn't change
		SimulatedTrades:      s,
	}
	file, _ := json.MarshalIndent(data, "", " ")
	fileName := fmt.Sprintf("%v.json", time.Now().Unix())
	_ = ioutil.WriteFile(fileName, file, 0644)

	return fileName
}

func saveBacktestRes(
	c []CandlestickChartData,
	p []ProfitCurveData,
	s []SimulatedTradeData,
	rid, reqBucketname, ticker, period, start, end string) {
	resFileName := makeBacktestResFile(c, p, s, ticker, period, start, end)

	storageClient, _ := storage.NewClient(ctx)
	defer storageClient.Close()
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	//if bucket doesn't exist, create new
	buckets, _ := listBuckets()
	var bucketName string
	for _, bn := range buckets {
		if bn == reqBucketname {
			bucketName = bn
		}
	}
	if bucketName == "" {
		bucket := storageClient.Bucket(reqBucketname)
		if err := bucket.Create(ctx, googleProjectID, nil); err != nil {
			fmt.Printf("Failed to create bucket: %v", err)
		}
		bucketName = reqBucketname
	}

	//create obj
	object := rid + ".json"
	// Open local file
	f, err := os.Open(resFileName)
	if err != nil {
		fmt.Printf("os.Open: %v", err)
	}
	defer f.Close()
	ctx2, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()
	// upload object with storage.Writer
	wc := storageClient.Bucket(bucketName).Object(object).NewWriter(ctx2)
	if _, err = io.Copy(wc, f); err != nil {
		fmt.Printf("io.Copy: %v", err)
	}
	if err := wc.Close(); err != nil {
		fmt.Printf("Writer.Close: %v", err)
	}

	//remove local file
	_ = os.Remove(resFileName)
}

func completeBacktestResFile(
	rawData BacktestResFile,
	userID, rid string,
	packetSize int, packetSender func(string, string, []CandlestickChartData, []ProfitCurveData, []SimulatedTradeData),
) ([]CandlestickChartData, []ProfitCurveData, []SimulatedTradeData) {
	//init
	var completeCandles []CandlestickChartData
	start, _ := time.Parse(httpTimeFormat, rawData.Start)
	end, _ := time.Parse(httpTimeFormat, rawData.End)
	fetchCandlesStart := start

	//complete in chunks
	for {
		if fetchCandlesStart.After(end) {
			break
		}

		fetchCandlesEnd := fetchCandlesStart.Add(periodDurationMap[rawData.Period] * time.Duration(packetSize))
		if fetchCandlesEnd.After(end) {
			fetchCandlesEnd = end
		}

		//fetch all standard data
		var chunkCandles []CandlestickChartData
		blankCandles := copyObjs(getCachedCandleData(rawData.Ticker, rawData.Period, fetchCandlesStart, fetchCandlesEnd),
			func(obj Candlestick) CandlestickChartData {
				chartC := CandlestickChartData{
					DateTime: obj.DateTime,
					Open:     obj.Open,
					High:     obj.High,
					Low:      obj.Low,
					Close:    obj.Close,
				}
				return chartC
			})
		//update with added info if exists in res file
		for _, candle := range blankCandles {
			var candleToAdd CandlestickChartData
			for _, rCan := range rawData.ModifiedCandlesticks {
				if rCan.DateTime == candle.DateTime {
					candleToAdd = rCan
				}
			}
			if candleToAdd.DateTime == "" || candleToAdd.Open == 0 {
				candleToAdd = candle
			}

			chunkCandles = append(chunkCandles, candleToAdd)
		}
		completeCandles = append(completeCandles, chunkCandles...)

		//stream data back to client in every chunk
		// fmt.Printf("Sending candles %v to %v\n", fetchCandlesStart, fetchCandlesEnd)
		packetSender(userID, rid, chunkCandles, rawData.ProfitCurve, rawData.SimulatedTrades)

		//increment
		fetchCandlesStart = fetchCandlesEnd.Add(periodDurationMap[rawData.Period])
	}

	return completeCandles, rawData.ProfitCurve, rawData.SimulatedTrades
}

// listBuckets lists buckets in the project.
func listBuckets() ([]string, error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("storage.NewClient: %v", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	var buckets []string
	it := client.Buckets(ctx, googleProjectID)
	for {
		battrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		buckets = append(buckets, battrs.Name)
	}
	return buckets, nil
}

// listFiles lists objects within specified bucket.
func listFiles(bucket string) []string {
	// bucket := "bucket-name"
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		fmt.Println(err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	var buckets []string
	it := client.Bucket(bucket).Objects(ctx, nil)
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err == nil {
			buckets = append(buckets, attrs.Name)
		}
	}
	return buckets
}

func deleteFile(bucket, object string) error {
	// bucket := "bucket-name"
	// object := "object-name"
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("storage.NewClient: %v", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	o := client.Bucket(bucket).Object(object)
	if err := o.Delete(ctx); err != nil {
		return fmt.Errorf("Object(%q).Delete: %v", object, err)
	}
	// fmt.Fprintf(w, "Blob %v deleted.\n", object)
	return nil
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
	keys, _ := rdbChartmaster.Keys(ctx, "*").Result()
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
