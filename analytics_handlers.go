package main

import (
	"encoding/json"
	"flag"
	"math/rand"
	"net/http"
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

	var scatterRes []ScatterData
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

	duration := make(map[int]string)
	// var duration []int64
	for _, x := range tradeAction {
		// tickerData = append(tickerData, x.AggregateID)
		if duration[x.AggregateID] == "" {
			duration[x.AggregateID] = x.Timestamp
		} else {
			layout := "2006-01-02_15:04:05_-0700"
			str1 := duration[x.AggregateID]
			t1, _ := time.Parse(layout, str1)
			str2 := x.Timestamp
			t2, _ := time.Parse(layout, str2)

			scatterRes = append(scatterRes,
				ScatterData{
					Profit:   0 + rand.Float64()*(1-0),
					Duration: t1.Sub(t2).Minutes(),
					Size:     5,
					Leverage: 13,
					Time:     20,
				})
		}
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