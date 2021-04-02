package main

import (
	"context"
	"encoding/json"
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
	if _, err := client.Put(ctx, newUserKey, &newUser); err != nil {
		log.Fatalf("Failed to save User: %v", err)
	}

	// return
	data := jsonResponse{
		Msg:  "Added " + newUserKey.String(),
		Body: newUser.String(),
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
	botsResp := make([]Bot, 0)

	auth, _ := url.QueryUnescape(r.Header.Get("Authorization"))
	authReq := loginReq{
		ID:       r.URL.Query()["user"][0],
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

	//set webhook URL
	plainWebhookID := generateWebhookID(100)
	encryptedWebhookID := encrypt(reqUser.EncryptKey, plainWebhookID)
	newBot.WebhookURL = "https://ana-api.myika.co/webhook/" + encryptedWebhookID

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

	//don't allow webhookURL passed in body, must be generated randomly
	if reqBotData.WebhookURL != "" {
		data := jsonResponse{Msg: "WebhookURL property of Bot cannot be set explicitly.", Body: "Do not pass WebhookURL property in request body."}
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
		webhookID := strings.TrimPrefix(x.WebhookURL, "https://ana-api.myika.co/webhook/")
		if isBase64(webhookID) {
			x.WebhookURL = "https://ana-api.myika.co/webhook/" + decrypt(reqUser.EncryptKey, webhookID)
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

func getAllExchangeConnectionsHandler(w http.ResponseWriter, r *http.Request) {
	setupCORS(&w, r)
	if (*r).Method == "OPTIONS" {
		return
	}

	exResp := make([]ExchangeConnection, 0)
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
		exResp = append(exResp, x)
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

	var newEx ExchangeConnection
	// decode data
	err := json.NewDecoder(r.Body).Decode(&newEx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

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

}

func tvWebhookHandler(w http.ResponseWriter, r *http.Request) {
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
	webhookID := mux.Vars(r)["id"]
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
	URLToFind := "https://ana-api.myika.co/webhook/" + webhookID
	for _, bot := range allBots {
		if bot.WebhookURL == URLToFind {
			botToUse = bot
		}
	}

	//TODO: call other services for given bot based on body props
	data := jsonResponse{
		Msg:  fmt.Sprintf("Bot to use: \n %s", botToUse.String()),
		Body: webHookReq.Msg + "/" + webHookReq.Size + "/" + webHookReq.User,
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
}
