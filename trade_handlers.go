package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"cloud.google.com/go/datastore"
	"google.golang.org/api/iterator"
)

func createNewTradeHandler(w http.ResponseWriter, r *http.Request) {
	setupCORS(&w, r)
	if (*r).Method == "OPTIONS" {
		return
	}

	if flag.Lookup("test.v") != nil {
		initDatastore()
	}

	var newTrade TradeAction
	// decode data
	err := json.NewDecoder(r.Body).Decode(&newTrade)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// create new listing in DB
	// kind := "TradeAction"
	// newKey := datastore.IncompleteKey(kind, nil)
	// addedKey, err := client.Put(ctx, newKey, &newTrade)
	// if err != nil {
	// 	log.Fatalf("Failed to save TradeAction: %v", err)
	// }

	//send on websocket stream
	ws := wsConnections[newTrade.UserID]
	if ws != nil {
		jsonTrade, _ := json.Marshal(newTrade)
		err := ws.WriteMessage(1, jsonTrade)
		if err != nil {
			log.Println(err)
		}
	}

	// return
	// data := jsonResponse{
	// 	Msg:  "Added " + newKey.String(),
	// 	Body: fmt.Sprint(addedKey.ID),
	// }
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	// json.NewEncoder(w).Encode(data)
}

func getAllTradesHandler(w http.ResponseWriter, r *http.Request) {
	setupCORS(&w, r)
	if (*r).Method == "OPTIONS" {
		return
	}

	if flag.Lookup("test.v") != nil {
		initDatastore()
	}

	tradesResp := make([]TradeAction, 0)
	auth, _ := url.QueryUnescape(r.Header.Get("Authorization"))
	authReq := loginReq{
		ID:       r.URL.Query()["user"][0],
		Password: auth,
	}
	authSuccess, _ := authenticateUser(authReq)
	if !authSuccess {
		data := jsonResponse{Msg: "Authorization Invalid", Body: "Go away."}
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(data)
		return
	}

	//configs before running query
	var query *datastore.Query
	userIDParam := r.URL.Query()["user"][0]
	query = datastore.NewQuery("TradeAction").Filter("UserID =", userIDParam)

	//run query
	t := client.Run(ctx, query)
	for {
		var x TradeAction
		key, err := t.Next(&x)
		if key != nil {
			x.KEY = fmt.Sprint(key.ID)
		}
		if err == iterator.Done {
			break
		}
		// if err != nil {
		// 	// Handle error.
		// }
		tradesResp = append(tradesResp, x)
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tradesResp)
}
