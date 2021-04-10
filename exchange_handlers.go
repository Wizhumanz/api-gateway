package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/gorilla/mux"
	"google.golang.org/api/iterator"
)

func getAllExchangeConnectionsHandler(w http.ResponseWriter, r *http.Request) {
	setupCORS(&w, r)
	if (*r).Method == "OPTIONS" {
		return
	}

	if flag.Lookup("test.v") != nil {
		initDatastore()
	}

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
	query = datastore.NewQuery("ExchangeConnection").Filter("UserID =", userIDParam)

	//run query
	exResp := make([]ExchangeConnection, 0)
	t := client.Run(ctx, query)
	for {
		var x ExchangeConnection
		_, err := t.Next(&x)
		if err == iterator.Done {
			break
		}
		if x.K.ID != 0 {
			x.KEY = fmt.Sprint(x.K.ID)
		}
		// if err != nil {
		// 	// Handle error.
		// }

		//event sourcing (pick latest snapshot)
		if len(exResp) == 0 {
			exResp = append(exResp, x)
		} else {
			//find exchange in existing array
			var exEx ExchangeConnection
			for _, e := range exResp {
				if e.APIKey == x.APIKey {
					exEx = e
				}
			}

			//if exchange exists, append row/entry with the latest timestamp
			if exEx.APIKey != "" || exEx.Timestamp != "" {
				//compare timestamps
				layout := "2006-01-02_15:04:05_-0700"
				existingTime, _ := time.Parse(layout, exEx.Timestamp)
				newTime, _ := time.Parse(layout, x.Timestamp)
				//if existing is older, remove it and add newer current one; otherwise, do nothing
				if existingTime.Before(newTime) {
					//rm existing listing
					exResp = deleteExchangeConnection(exResp, exEx)
					//append current listing
					exResp = append(exResp, x)
				}
			} else {
				//otherwise, just append newly decoded (so far unique) bot
				exResp = append(exResp, x)
			}
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(exResp)
}

func createNewExchangeConnectionHandler(w http.ResponseWriter, r *http.Request) {
	setupCORS(&w, r)
	if (*r).Method == "OPTIONS" {
		return
	}

	if flag.Lookup("test.v") != nil {
		initDatastore()
	}

	var newEx ExchangeConnection
	// decode data
	err := json.NewDecoder(r.Body).Decode(&newEx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	//log creation timestamp
	newEx.Timestamp = time.Now().Format("2006-01-02_15:04:05_-0700")

	// create new ExchangeConnection in DB
	kind := "ExchangeConnection"
	newUserKey := datastore.IncompleteKey(kind, nil)
	if _, err := client.Put(ctx, newUserKey, &newEx); err != nil {
		log.Fatalf("Failed to save ExchangeConnection: %v", err)
	}

	// return
	data := jsonResponse{
		Msg:  "Added exchange connection.",
		Body: newEx.Name + " - " + newEx.APIKey,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(data)
}

func deleteExchangeConnectionHandler(w http.ResponseWriter, r *http.Request) {
	setupCORS(&w, r)
	if (*r).Method == "OPTIONS" {
		return
	}

	if flag.Lookup("test.v") != nil {
		initDatastore()
	}

	exs := make([]ExchangeConnection, 0)

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

	//check if ExchangeConnection already exists to delete
	exDelID, unescapeErr := url.QueryUnescape(mux.Vars(r)["id"]) //aggregate ID, not DB __key__
	if unescapeErr != nil {
		data := jsonResponse{Msg: "Exchange ID Parse Error", Body: unescapeErr.Error()}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(data)
		return
	}
	intID, _ := strconv.Atoi(exDelID)
	key := datastore.IDKey("ExchangeConnection", int64(intID), nil)
	query := datastore.NewQuery("ExchangeConnection").
		Filter("__key__ =", key)
	t := client.Run(ctx, query)
	for {
		var x ExchangeConnection
		key, err := t.Next(&x)
		if err == iterator.Done {
			break
		}
		// if err != nil {
		// 	// Handle error.
		// }
		if key != nil {
			x.KEY = fmt.Sprint(key.ID)
		}
		exs = append(exs, x)
	}

	//return if ExchangeConnection to update doesn't exist
	isDelIdValid := len(exs) > 0 && exs[0].APIKey != ""
	if !isDelIdValid {
		data := jsonResponse{Msg: "ExchangeConnection ID Invalid", Body: "ExchangeConnection with provided ID does not exist."}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(data)
		return
	}

	// add new row to DB
	exToDel := exs[len(exs)-1]
	exToDel.IsDeleted = true
	kind := "ExchangeConnection"
	newKey := datastore.IncompleteKey(kind, nil)
	if _, err := client.Put(ctx, newKey, &exToDel); err != nil {
		log.Fatalf("Failed to delete ExchangeConnection: %v", err)
	}

	// return
	data := jsonResponse{
		Msg:  "DELETED exchange connection.",
		Body: exToDel.Name + " - " + exToDel.APIKey,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
	json.NewEncoder(w).Encode(data)
}

func getAllWebhookConnectionHandler(w http.ResponseWriter, r *http.Request) {
	setupCORS(&w, r)
	if (*r).Method == "OPTIONS" {
		return
	}

	if flag.Lookup("test.v") != nil {
		initDatastore()
	}

	webhookResp := make([]WebhookConnection, 0)

	//configs before running query
	var query *datastore.Query
	query = datastore.NewQuery("WebhookConnection").Filter("IsPublic =", true)

	//run query
	t := client.Run(ctx, query)
	for {
		var x WebhookConnection
		key, err := t.Next(&x)
		if err == iterator.Done {
			break
		}

		if key != nil {
			x.KEY = fmt.Sprint(key.ID)
		}
		// if err != nil {
		// 	// Handle error.
		// }
		webhookResp = append(webhookResp, x)
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(webhookResp)
}
