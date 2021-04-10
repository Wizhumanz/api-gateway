package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/gorilla/mux"
	"gitlab.com/myikaco/msngr"
	"google.golang.org/api/iterator"
)

func getWebhookConnectionHandler(w http.ResponseWriter, r *http.Request) {
	setupCORS(&w, r)
	if (*r).Method == "OPTIONS" {
		return
	}

	if flag.Lookup("test.v") != nil {
		initDatastore()
	}

	//get query string ids
	rawIDs := r.URL.Query()["ids"][0]
	batchReqIDs := strings.Split(rawIDs, " ")

	if !(len(batchReqIDs) > 0) {
		data := jsonResponse{Msg: "IDs array param empty.", Body: "Pass ids property in json as array of strings."}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(data)
		return
	}

	var retWebhookConns []WebhookConnection
	for _, id := range batchReqIDs {
		//query object with id
		var query *datastore.Query
		intID, _ := strconv.Atoi(id)
		k := datastore.IDKey("WebhookConnection", int64(intID), nil)
		query = datastore.NewQuery("WebhookConnection").Filter("__key__ =", k)

		//parse into struct
		var res WebhookConnection
		t := client.Run(ctx, query)
		for {
			key, err := t.Next(&res)
			if err == iterator.Done {
				break
			}

			if key != nil {
				res.KEY = fmt.Sprint(key.ID)
			} else {
				break
			}
			// if err != nil {
			// 	// Handle error.
			// }
		}
		retWebhookConns = append(retWebhookConns, res)
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(retWebhookConns)
}

func createNewWebhookConnectionHandler(w http.ResponseWriter, r *http.Request) {
	setupCORS(&w, r)
	if (*r).Method == "OPTIONS" {
		return
	}

	if flag.Lookup("test.v") != nil {
		initDatastore()
	}

	//set webhook URL
	plainWebhookID := generateWebhookID(100)
	webhookURL := "https://ana-api.myika.co/webhook/" + plainWebhookID

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

	var webhook WebhookConnection

	// error := json.NewDecoder(r.Body).Decode(&webhook)
	// if error != nil {
	// 	http.Error(w, error.Error(), http.StatusBadRequest)
	// 	return
	// }

	webhook.IsPublic = false
	webhook.URL = webhookURL

	// create new listing in DB
	kind := "WebhookConnection"
	newWebhookKey := datastore.IncompleteKey(kind, nil)
	newKey, err := client.Put(ctx, newWebhookKey, &webhook)
	if err != nil {
		log.Fatalf("Failed to save User: %v", err)
	}

	// return
	data := jsonResponse{
		Msg:  "Added " + newWebhookKey.String(),
		Body: fmt.Sprint(newKey.ID),
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(data)
}

func tvWebhookHandler(w http.ResponseWriter, r *http.Request) {
	setupCORS(&w, r)
	if (*r).Method == "OPTIONS" {
		return
	}

	if flag.Lookup("test.v") != nil {
		initDatastore()
	}

	var webhookReq webHookRequest
	err := json.NewDecoder(r.Body).Decode(&webhookReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	//check required body params
	if webhookReq.Ticker == "" {
		//TODO: alert user of error, not caller
		data := jsonResponse{
			Msg:  "Webhook body invalid.",
			Body: "Ticker field in request body is required, found \"\".",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(data)
		return
	}

	// check webhookConnID valid
	webhookID := mux.Vars(r)["id"]
	// fmt.Printf("webhookID: %s \n", webhookID)
	if webhookID == "" {
		//TODO: alert user of error, not caller
		data := jsonResponse{
			Msg:  "Webhook URL invalid.",
			Body: "No webhook id passed in /webhook/{id}.",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(data)
		return
	}

	//get webhookConn ID
	var theWebConn WebhookConnection
	fullURL := "https://ana-api.myika.co" + r.URL.String()
	webConnQuery := datastore.NewQuery("WebhookConnection").
		Filter("URL =", fullURL)
	tWebConn := client.Run(ctx, webConnQuery)
	_, webConnErr := tWebConn.Next(&theWebConn)
	if webConnErr != nil {
		//handle error
	}
	theWebConn.KEY = fmt.Sprint(theWebConn.K.ID)

	//get bot(s) to execute strategy on
	var allBots []Bot
	var botQuery *datastore.Query
	if webhookReq.User == "" {
		fmt.Println("PUBLIC strat")
		//public strategy
		botQuery = datastore.NewQuery("Bot").
			Filter("WebhookConnectionID =", theWebConn.KEY)
	} else {
		fmt.Println("PRIVATE strat")
		//private strategy (custom webhookURL)
		botQuery = datastore.NewQuery("Bot").
			Filter("UserID =", webhookReq.User).
			Filter("WebhookConnectionID =", theWebConn.KEY)
	}
	tBot := client.Run(ctx, botQuery)
	allBots = parseBotsQueryRes(tBot, User{})

	if len(allBots) == 0 {
		//TODO: alert user of error, not caller
		data := jsonResponse{
			Msg:  "Unable to get bots for this userID.",
			Body: webhookReq.User,
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(data)
		return
	}

	//exec trade for each bot
	for _, botToUse := range allBots {
		fmt.Println(botToUse.String())
		//check bot validity
		if botToUse.AccountRiskPercPerTrade == "" ||
			botToUse.AccountSizePercToTrade == "" ||
			botToUse.Leverage == "" ||
			botToUse.Ticker == "" {
			//TODO: alert user of error, not caller
			data := jsonResponse{
				Msg:  "Bot with ID invalid: " + botToUse.KEY,
				Body: botToUse.String(),
			}
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(data)
			return
		}

		// save new TradeAction to DB //
		x := TradeAction{
			UserID:    botToUse.UserID,
			BotID:     fmt.Sprint(botToUse.K.ID),
			Timestamp: time.Now().Format("2006-01-02_15:04:05_-0700"),
			Ticker:    webhookReq.Ticker,
			Exchange:  botToUse.ExchangeConnection,
		}

		//set aggregate ID + trade desc
		switch webhookReq.TradeActionType {
		case "ENTER":
			fmt.Println("NEW aggr ID")
			// NEW aggr ID (get highest, then increment)
			var calcTA TradeAction
			query := datastore.NewQuery("TradeAction").
				Project("AggregateID").
				Order("-AggregateID")
			t := client.Run(ctx, query)
			_, error := t.Next(&calcTA)
			if error != nil {
				// Handle error.
			}

			x.AggregateID = calcTA.AggregateID + 1
			x.Action = "ENTERIntentSubmitted"
		case "EXIT", "SL", "TP":
			fmt.Println("REUSING aggr ID")
			// use previous aggr ID from existing trade
			var calcTA TradeAction
			//NOTE: one bot can only be in one trade at any one time
			query := datastore.NewQuery("TradeAction").
				Filter("UserID =", botToUse.UserID).
				Filter("BotID =", fmt.Sprint(botToUse.K.ID))
				// Project("AggregateID").
				// Order("-AggregateID")
			t := client.Run(ctx, query)
			_, error := t.Next(&calcTA)
			if error != nil {
				// Handle error.
			}

			x.AggregateID = calcTA.AggregateID
			x.Action = webhookReq.TradeActionType + "IntentSubmitted"
		}

		//add row to DB
		kind := "TradeAction"
		newUserKey := datastore.IncompleteKey(kind, nil)
		if _, err := client.Put(ctx, newUserKey, &x); err != nil {
			log.Fatalf("Failed to save ExchangeConnection: %v", err)
		}

		//create redis stream key <aggregateID>:<userID>:<botID>
		tradeStreamName := fmt.Sprint(x.AggregateID) + ":" + botToUse.UserID + ":" + botToUse.KEY

		// add new trade info into stream (triggers other services)
		msgs := []string{}
		msgs = append(msgs, "TradeStreamName")
		msgs = append(msgs, tradeStreamName)
		msgs = append(msgs, "AccountRiskPercPerTrade")
		msgs = append(msgs, fmt.Sprint(botToUse.AccountRiskPercPerTrade))
		msgs = append(msgs, "AccountSizePercToTrade")
		msgs = append(msgs, fmt.Sprint(botToUse.AccountSizePercToTrade))
		msgs = append(msgs, "Leverage")
		msgs = append(msgs, fmt.Sprint(botToUse.Leverage))
		msgs = append(msgs, "Ticker")
		msgs = append(msgs, fmt.Sprint(botToUse.Ticker))
		msgs = append(msgs, "CMD")
		msgs = append(msgs, webhookReq.TradeActionType)

		msngr.AddToStream("webhookTrades", msgs)
		fmt.Println("-")

		w.WriteHeader(http.StatusOK)
		return
	}
}
