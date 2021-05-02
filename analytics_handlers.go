package main

import (
	"encoding/json"
	"flag"
	"math/rand"
	"net/http"
	"sort"
	"time"

	"cloud.google.com/go/datastore"
	"google.golang.org/api/iterator"
)

func pieHandler(w http.ResponseWriter, r *http.Request) {
	setupCORS(&w, r)
	if (*r).Method == "OPTIONS" {
		return
	}

	if flag.Lookup("test.v") != nil {
		initDatastore()
	}

	// auth, _ := url.QueryUnescape(r.Header.Get("Authorization"))
	// authReq := loginReq{
	// 	ID:       r.URL.Query()["user"][0],
	// 	Password: auth,
	// }
	// authSuccess, _ := authenticateUser(authReq)
	// if !authSuccess {
	// 	data := jsonResponse{Msg: "Authorization Invalid", Body: "Go away."}
	// 	w.WriteHeader(http.StatusUnauthorized)
	// 	json.NewEncoder(w).Encode(data)
	// 	return
	// }

	// //get query string graph
	// rawIDs := r.URL.Query()["graph"][0]
	// batchReqIDs := strings.Split(rawIDs, " ")

	// if !(len(batchReqIDs) > 0) {
	// 	data := jsonResponse{Msg: "IDs array param empty.", Body: "Pass ids property in json as array of strings."}
	// 	w.WriteHeader(http.StatusBadRequest)
	// 	json.NewEncoder(w).Encode(data)
	// 	return
	// }
	var tradeAction []TradeAction
	kind := "TradeAction"
	query := datastore.NewQuery(kind)

	t := client.Run(ctx, query)
	for {
		var x TradeAction
		_, err := t.Next(&x)

		if err == iterator.Done {
			break
		}
		tradeAction = append(tradeAction, x)
	}

	var tickerData []string
	for _, x := range tradeAction {
		tickerData = append(tickerData, x.Ticker)
	}

	// return
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tickerData)
}

func scatterHandler(w http.ResponseWriter, r *http.Request) {
	var scatterRes []ScatterData

	setupCORS(&w, r)
	if (*r).Method == "OPTIONS" {
		return
	}

	if flag.Lookup("test.v") != nil {
		initDatastore()
	}

	//get all user's TradeActions
	var tas []TradeAction
	kind := "TradeAction"
	//TODO: dynamic user ID filter for TradeActions
	query := datastore.NewQuery(kind).Filter("UserID =", "5632499082330112")
	t := client.Run(ctx, query)
	for {
		var x TradeAction
		_, err := t.Next(&x)
		if err == iterator.Done {
			break
		}
		tas = append(tas, x)
	}

	//make map of [aggrID]TradeActions
	taAggrIDMap := make(map[int][]TradeAction)
	for _, x := range tas {
		if len(taAggrIDMap[x.AggregateID]) == 0 {
			var taArr []TradeAction
			taArr = append(taArr, x)
			taAggrIDMap[x.AggregateID] = taArr
		} else {
			taAggrIDMap[x.AggregateID] = append(taAggrIDMap[x.AggregateID], x)
		}
	}

	//find duration for each trade
	for _, sliTA := range taAggrIDMap {
		//sort each []TradeAction in ascending order of timestamp
		sort.SliceStable(sliTA, func(i, j int) bool {
			t1, _ := time.Parse("2006-01-02_15:04:05_-0700", sliTA[i].Timestamp)
			t2, _ := time.Parse("2006-01-02_15:04:05_-0700", sliTA[j].Timestamp)
			return t2.Sub(t1).Minutes() > 0
		})
		//last timestamp - first timestamp = duration
		end, _ := time.Parse("2006-01-02_15:04:05_-0700", sliTA[len(sliTA)-1].Timestamp)
		start, _ := time.Parse("2006-01-02_15:04:05_-0700", sliTA[0].Timestamp)
		duration := end.Sub(start).Minutes()

		scatterRes = append(scatterRes,
			ScatterData{
				Profit:   0 + rand.Float64()*(1-0),
				Duration: duration,
				Size:     5,
				Leverage: 13,
				Time:     20,
			})
	}

	// return
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(scatterRes)
}

func stackedHandler(w http.ResponseWriter, r *http.Request) {
	setupCORS(&w, r)
	if (*r).Method == "OPTIONS" {
		return
	}

	if flag.Lookup("test.v") != nil {
		initDatastore()
	}

	// auth, _ := url.QueryUnescape(r.Header.Get("Authorization"))
	// authReq := loginReq{
	// 	ID:       r.URL.Query()["user"][0],
	// 	Password: auth,
	// }
	// authSuccess, _ := authenticateUser(authReq)
	// if !authSuccess {
	// 	data := jsonResponse{Msg: "Authorization Invalid", Body: "Go away."}
	// 	w.WriteHeader(http.StatusUnauthorized)
	// 	json.NewEncoder(w).Encode(data)
	// 	return
	// }

	// //get query string graph
	// rawIDs := r.URL.Query()["graph"][0]
	// batchReqIDs := strings.Split(rawIDs, " ")

	// if !(len(batchReqIDs) > 0) {
	// 	data := jsonResponse{Msg: "IDs array param empty.", Body: "Pass ids property in json as array of strings."}
	// 	w.WriteHeader(http.StatusBadRequest)
	// 	json.NewEncoder(w).Encode(data)
	// 	return
	// }

	// return
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(nil)
}
