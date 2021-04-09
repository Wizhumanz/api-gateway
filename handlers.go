package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/gorilla/mux"
	"google.golang.org/api/iterator"
)

// route handlers

func indexHandler(w http.ResponseWriter, r *http.Request) {
	setupCORS(&w, r)
	if (*r).Method == "OPTIONS" {
		return
	}

	if flag.Lookup("test.v") != nil {
		initDatastore()
	}

	var data jsonResponse
	w.Header().Set("Content-Type", "application/json")
	data = jsonResponse{Msg: "Anastasia API Gateway", Body: "Ready"}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
	// w.Write([]byte(`{"msg": "привет сука"}`))
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	setupCORS(&w, r)
	if (*r).Method == "OPTIONS" {
		return
	}

	if flag.Lookup("test.v") != nil {
		initDatastore()
	}

	var newLoginReq loginReq
	// decode data
	err := json.NewDecoder(r.Body).Decode(&newLoginReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var data jsonResponse
	authSuccess, loggedInUser := authenticateUser(newLoginReq)
	if authSuccess {
		data = jsonResponse{
			Msg:  "Successfully logged in!",
			Body: fmt.Sprint(loggedInUser.K.ID),
		}
		w.WriteHeader(http.StatusOK)
	} else {
		data = jsonResponse{
			Msg:  "Authentication failed. Fuck off!",
			Body: newLoginReq.Email,
		}
		w.WriteHeader(http.StatusUnauthorized)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func createNewUserHandler(w http.ResponseWriter, r *http.Request) {
	setupCORS(&w, r)
	if (*r).Method == "OPTIONS" {
		return
	}

	if flag.Lookup("test.v") != nil {
		initDatastore()
	}

	var newUser User
	// decode data
	err := json.NewDecoder(r.Body).Decode(&newUser)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// create password hash
	newUser.Password, _ = HashPassword(newUser.Password)
	// create encrypt key of fixed length
	rand.Seed(time.Now().UnixNano())
	newUser.EncryptKey = generateEncryptKey(32)

	// create new listing in DB
	kind := "User"
	newUserKey := datastore.IncompleteKey(kind, nil)
	addedKey, err := client.Put(ctx, newUserKey, &newUser)
	if err != nil {
		log.Fatalf("Failed to save User: %v", err)
	}

	// return
	data := jsonResponse{
		Msg:  "Added " + newUserKey.String(),
		Body: fmt.Sprint(addedKey.ID),
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(data)
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

func getAllBotsHandler(w http.ResponseWriter, r *http.Request) {
	setupCORS(&w, r)
	if (*r).Method == "OPTIONS" {
		return
	}

	if flag.Lookup("test.v") != nil {
		initDatastore()
	}

	botsResp := make([]Bot, 0)

	//check for user query string
	if len(r.URL.Query().Get("user")) == 0 {
		data := jsonResponse{Msg: "User param missing.", Body: "User param must be passed in query string."}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(data)
		return
	}

	auth, _ := url.QueryUnescape(r.Header.Get("Authorization"))
	authReq := loginReq{
		ID:       r.URL.Query().Get("user"),
		Password: auth,
	}
	authSuccess, reqUser := authenticateUser(authReq)
	if len(r.URL.Query()["isActive"]) == 0 && !authSuccess {
		data := jsonResponse{Msg: "Authorization Invalid", Body: "Go away."}
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(data)
		return
	}
	//build query based on passed URL params
	var query *datastore.Query
	userIDParam := r.URL.Query()["user"][0]
	var isActiveParam = true //default
	if len(r.URL.Query()["isActive"]) > 0 {
		//extract correct isActive param
		isActiveQueryStr := r.URL.Query()["isActive"][0]
		if isActiveQueryStr == "true" {
			isActiveParam = true
		} else if isActiveQueryStr == "false" {
			isActiveParam = false
		}

		query = datastore.NewQuery("Bot").
			Filter("UserID =", userIDParam).
			Filter("IsActive =", isActiveParam)
	} else {
		query = datastore.NewQuery("Bot").
			Filter("UserID =", userIDParam)
	}

	//run query
	t := client.Run(ctx, query)
	botsResp = parseBotsQueryRes(t, reqUser)

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(botsResp)
}

// almost identical logic with create and update (event sourcing)
func addBot(w http.ResponseWriter, r *http.Request, isPutReq bool, reqBot Bot, reqUser User) {
	setupCORS(&w, r)
	if (*r).Method == "OPTIONS" {
		return
	}

	if flag.Lookup("test.v") != nil {
		initDatastore()
	}
	newBot := reqBot

	// if updating, name field not passed in JSON body, so must fill
	if isPutReq {
		newBot.AggregateID = reqBot.AggregateID
	} else {
		// else increment aggregate ID
		var x Bot
		//get highest aggregate ID
		query := datastore.NewQuery("Bot").
			Project("AggregateID").
			Order("-AggregateID")
		t := client.Run(ctx, query)
		_, error := t.Next(&x)
		if error != nil {
			// Handle error.
		}
		newBot.AggregateID = x.AggregateID + 1
	}

	//encrypt sensitive bot data
	newBot.AccountRiskPercPerTrade = encrypt(reqUser.EncryptKey, newBot.AccountRiskPercPerTrade)
	newBot.AccountSizePercToTrade = encrypt(reqUser.EncryptKey, newBot.AccountSizePercToTrade)
	newBot.Leverage = encrypt(reqUser.EncryptKey, newBot.Leverage)

	//set timestamp
	newBot.Timestamp = time.Now().Format("2006-01-02_15:04:05_-0700")

	// create new bot in DB
	ctx := context.Background()
	var newBotKey *datastore.Key
	clientAdd, err := datastore.NewClient(ctx, googleProjectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	kind := "Bot"
	newBotKey = datastore.IncompleteKey(kind, nil)

	if _, err := clientAdd.Put(ctx, newBotKey, &newBot); err != nil {
		log.Fatalf("Failed to save Bot: %v", err)
	}

	// return
	data := jsonResponse{
		Msg:  "Added " + newBotKey.String(),
		Body: newBot.String(),
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(data)
}

func updateBotHandler(w http.ResponseWriter, r *http.Request) {
	setupCORS(&w, r)
	if (*r).Method == "OPTIONS" {
		return
	}

	if flag.Lookup("test.v") != nil {
		initDatastore()
	}

	botsResp := make([]Bot, 0)

	auth, _ := url.QueryUnescape(r.Header.Get("Authorization"))
	authReq := loginReq{
		ID:       r.URL.Query()["user"][0],
		Password: auth,
	}
	authSuccess, reqUser := authenticateUser(authReq)
	if !authSuccess {
		data := jsonResponse{Msg: "Authorization Invalid", Body: "Go away."}
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(data)
		return
	}

	// if updating bot, don't allow AggregateID change
	var reqBotData Bot
	err := json.NewDecoder(r.Body).Decode(&reqBotData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if reqBotData.AggregateID != 0 {
		data := jsonResponse{Msg: "ID property of Bot is immutable.", Body: "Do not pass ID property in request body, instead pass in URL."}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(data)
		return
	}

	//check if bot already exists to update
	botToUpdateID, unescapeErr := url.QueryUnescape(mux.Vars(r)["id"]) //aggregate ID, not DB __key__
	if unescapeErr != nil {
		data := jsonResponse{Msg: "Bot ID Parse Error", Body: unescapeErr.Error()}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(data)
		return
	}
	int, _ := strconv.Atoi(botToUpdateID)
	query := datastore.NewQuery("Bot").
		Filter("AggregateID =", int)
	t := client.Run(ctx, query)
	for {
		var x Bot
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

		//decrypt props
		if isBase64(x.AccountRiskPercPerTrade) {
			x.AccountRiskPercPerTrade = decrypt(reqUser.EncryptKey, x.AccountRiskPercPerTrade)
		}
		if isBase64(x.AccountSizePercToTrade) {
			x.AccountSizePercToTrade = decrypt(reqUser.EncryptKey, x.AccountSizePercToTrade)
		}
		if isBase64(x.Leverage) {
			x.Leverage = decrypt(reqUser.EncryptKey, x.Leverage)
		}

		botsResp = append(botsResp, x)
	}

	//return if bot to update doesn't exist
	putIDValid := len(botsResp) > 0 && botsResp[0].UserID != ""
	if !putIDValid {
		data := jsonResponse{Msg: "Bot ID Invalid", Body: "Bot with provided ID does not exist."}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(data)
		return
	}

	//must assign aggregate ID from existing bot
	reqBotData.AggregateID = botsResp[len(botsResp)-1].AggregateID

	addBot(w, r, true, reqBotData, reqUser)
}

func createNewBotHandler(w http.ResponseWriter, r *http.Request) {
	setupCORS(&w, r)
	if (*r).Method == "OPTIONS" {
		return
	}

	if flag.Lookup("test.v") != nil {
		initDatastore()
	}

	// decode data
	var newBot Bot
	err := json.NewDecoder(r.Body).Decode(&newBot)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	//authenticate
	auth, _ := url.QueryUnescape(r.Header.Get("Authorization"))
	authReq := loginReq{
		ID:       newBot.UserID,
		Password: auth,
	}
	authSuccess, reqUser := authenticateUser(authReq)
	if !authSuccess {
		data := jsonResponse{Msg: "Authorization Invalid", Body: "Go away."}
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(data)
		return
	}

	addBot(w, r, false, newBot, reqUser) //empty Bot struct passed just for compiler
}

func deleteBotHandler(w http.ResponseWriter, r *http.Request) {
	setupCORS(&w, r)
	if (*r).Method == "OPTIONS" {
		return
	}

	if flag.Lookup("test.v") != nil {
		initDatastore()
	}

	bots := make([]Bot, 0)

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

	//check if bot already exists to delete
	botDelID, unescapeErr := url.QueryUnescape(mux.Vars(r)["id"]) //aggregate ID, not DB __key__
	if unescapeErr != nil {
		data := jsonResponse{Msg: "Bot ID Parse Error", Body: unescapeErr.Error()}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(data)
		return
	}
	intID, _ := strconv.Atoi(botDelID)
	key := datastore.IDKey("Bot", int64(intID), nil)
	query := datastore.NewQuery("Bot").
		Filter("__key__ =", key)
	t := client.Run(ctx, query)
	for {
		var x Bot
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
		bots = append(bots, x)
	}

	//return if ExchangeConnection to update doesn't exist
	isDelIdValid := len(bots) > 0 && bots[0].K.ID != 0
	if !isDelIdValid {
		data := jsonResponse{Msg: "Bot ID Invalid", Body: "Bot with provided ID does not exist."}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(data)
		return
	}

	// add new row to DB
	botToDel := bots[len(bots)-1]
	botToDel.IsArchived = true
	botToDel.Timestamp = time.Now().Format("2006-01-02_15:04:05_-0700")
	kind := "Bot"
	newKey := datastore.IncompleteKey(kind, nil)
	if _, err := client.Put(ctx, newKey, &botToDel); err != nil {
		log.Fatalf("Failed to delete Bot: %v", err)
	}

	// return
	data := jsonResponse{
		Msg:  "DELETED bot.",
		Body: fmt.Sprint(botToDel.K.ID),
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
	json.NewEncoder(w).Encode(data)
}

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

	// create new listing in DB
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

	var webHookReq webHookRequest
	err := json.NewDecoder(r.Body).Decode(&webHookReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	//req validity check
	if webHookReq.User == "" {
		data := jsonResponse{
			Msg:  "User field in webhook body nil!",
			Body: "Must pass User field in webhook body JSON.",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(data)
		return
	}

	// get user with ID
	var webhookUser User
	intUserID, _ := strconv.Atoi(webHookReq.User)
	key := datastore.IDKey("User", int64(intUserID), nil)
	query := datastore.NewQuery("User").
		Filter("__key__ =", key)
	t := client.Run(ctx, query)
	_, error := t.Next(&webhookUser)
	if error != nil {
		// Handle error.
	}

	// get the bot referred to by webhookID
	// webhookID := mux.Vars(r)["id"]
	var allBots []Bot
	var botToUse Bot
	botQuery := datastore.NewQuery("Bot").
		Filter("UserID =", webHookReq.User)
	tBot := client.Run(ctx, botQuery)
	allBots = parseBotsQueryRes(tBot, webhookUser)

	if len(allBots) == 0 {
		data := jsonResponse{
			Msg:  "Unable to get bots for this userID.",
			Body: webHookReq.User,
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(data)
		return
	}

	//find bot that owns webhookURL passed in request
	// URLToFind := "https://ana-api.myika.co/webhook/" + webhookID
	// for _, bot := range allBots {
	// 	if bot.WebhookURL == URLToFind {
	// 		botToUse = bot
	// 	}
	// }

	//TODO: call other services for given bot based on body props
	data := jsonResponse{
		Msg:  fmt.Sprintf("Bot to use: \n %s", botToUse.String()),
		Body: webHookReq.Msg + "/" + webHookReq.Size + "/" + webHookReq.User,
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
}
