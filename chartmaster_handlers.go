package main

import (
	"encoding/json"
	"flag"
	"math/rand"
	"net/http"
	"time"
)

func indexChartmasterHandler(w http.ResponseWriter, r *http.Request) {
	var retData []ChartmasterData

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
	minChange := 4000
	maxChange := 50000
	minWick := 500
	maxWick := 5000
	startDate := time.Date(2021, time.January, 1, 0, 0, 0, 0, time.Now().UTC().Location())
	for i := 0; i < 80; i++ {
		var new ChartmasterData

		//body
		if i != 0 {
			startDate = startDate.AddDate(0, 0, 1)
			new = ChartmasterData{
				Date: startDate.Format("2006-01-02"),
				Open: retData[len(retData)-1].Close,
			}
		} else {
			new = ChartmasterData{
				Date: startDate.Format("2006-01-02"),
				Open: float64(rand.Intn(max-min+1)+min) / 100,
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

	// return
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(retData)
}
