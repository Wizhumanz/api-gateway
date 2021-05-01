package main

import (
	"encoding/json"
	"flag"
	"net/http"

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
	var pieData []TradeAction
	kind := "TradeAction"
	query := datastore.NewQuery(kind)

	t := client.Run(ctx, query)
	for {
		var x TradeAction
		_, err := t.Next(&x)

		if err == iterator.Done {
			break
		}
		pieData = append(pieData, x)
	}

	var tickerData []string
	for _, x := range pieData {
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

	var scatterRes []ScatterData

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
	var pieData []TradeAction
	kind := "TradeAction"
	query := datastore.NewQuery(kind)

	t := client.Run(ctx, query)
	for {
		var x TradeAction
		_, err := t.Next(&x)

		if err == iterator.Done {
			break
		}
		pieData = append(pieData, x)
	}

	var tickerData []string
	for _, x := range pieData {
		tickerData = append(tickerData, x.Ticker)
	}

	// if rawIDs == "Scatter" {
	scatterRes = append(scatterRes,
		ScatterData{
			Profit:   1,
			Duration: 2,
			Size:     5,
			Leverage: 13,
			Time:     20,
		},
		ScatterData{
			Profit:   2,
			Duration: 3,
			Size:     4,
			Leverage: 4,
			Time:     12,
		},
		ScatterData{
			Profit:   3,
			Duration: 1,
			Size:     4,
			Leverage: 16,
			Time:     3,
		},
		ScatterData{
			Profit:   4,
			Duration: 7,
			Size:     4,
			Leverage: 12,
			Time:     19,
		},
	)
	// }

	// //get the bot from DB
	// for _, id := range batchReqIDs {
	// 	intID, _ := strconv.Atoi(id)
	// 	key := datastore.IDKey("Bot", int64(intID), nil)
	// 	query := datastore.NewQuery("Bot").
	// 		Filter("__key__ =", key)
	// 	t := client.Run(ctx, query)
	// 	botRes := parseBotsQueryRes(t)
	// 	botsRes = append(botsRes, botRes[0])
	// }

	// return
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tickerData)
}

func stackedHandler(w http.ResponseWriter, r *http.Request) {
	setupCORS(&w, r)
	if (*r).Method == "OPTIONS" {
		return
	}

	if flag.Lookup("test.v") != nil {
		initDatastore()
	}

	var scatterRes []ScatterData

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
	var pieData []TradeAction
	kind := "TradeAction"
	query := datastore.NewQuery(kind)

	t := client.Run(ctx, query)
	for {
		var x TradeAction
		_, err := t.Next(&x)

		if err == iterator.Done {
			break
		}
		pieData = append(pieData, x)
	}

	var tickerData []string
	for _, x := range pieData {
		tickerData = append(tickerData, x.Ticker)
	}

	// if rawIDs == "Scatter" {
	scatterRes = append(scatterRes,
		ScatterData{
			Profit:   1,
			Duration: 2,
			Size:     5,
			Leverage: 13,
			Time:     20,
		},
		ScatterData{
			Profit:   2,
			Duration: 3,
			Size:     4,
			Leverage: 4,
			Time:     12,
		},
		ScatterData{
			Profit:   3,
			Duration: 1,
			Size:     4,
			Leverage: 16,
			Time:     3,
		},
		ScatterData{
			Profit:   4,
			Duration: 7,
			Size:     4,
			Leverage: 12,
			Time:     19,
		},
	)
	// }

	// //get the bot from DB
	// for _, id := range batchReqIDs {
	// 	intID, _ := strconv.Atoi(id)
	// 	key := datastore.IDKey("Bot", int64(intID), nil)
	// 	query := datastore.NewQuery("Bot").
	// 		Filter("__key__ =", key)
	// 	t := client.Run(ctx, query)
	// 	botRes := parseBotsQueryRes(t)
	// 	botsRes = append(botsRes, botRes[0])
	// }

	// return
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tickerData)
}
