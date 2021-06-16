package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"runtime"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/gorilla/mux"
	"google.golang.org/api/iterator"
)

func getBotHandler(w http.ResponseWriter, r *http.Request) {
	setupCORS(&w, r)
	if (*r).Method == "OPTIONS" {
		return
	}

	if flag.Lookup("test.v") != nil {
		initDatastore()
	}

	var botsRes []Bot

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

	//get query string ids
	rawIDs := r.URL.Query()["ids"][0]
	batchReqIDs := strings.Split(rawIDs, " ")

	if !(len(batchReqIDs) > 0) {
		data := jsonResponse{Msg: "IDs array param empty.", Body: "Pass ids property in json as array of strings."}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(data)
		return
	}

	//get the bot from DB
	for _, id := range batchReqIDs {
		intID, _ := strconv.Atoi(id)
		key := datastore.IDKey("Bot", int64(intID), nil)
		query := datastore.NewQuery("Bot").
			Filter("__key__ =", key)
		t := client.Run(ctx, query)
		botRes := parseBotsQueryRes(t)
		botsRes = append(botsRes, botRes[0])
	}

	// return
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(botsRes)
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
	authSuccess, _ := authenticateUser(authReq)
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
	botsResp = parseBotsQueryRes(t)

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
	newBot.AccountRiskPercPerTrade = encrypt(newBot.AccountRiskPercPerTrade)
	newBot.AccountSizePercToTrade = encrypt(newBot.AccountSizePercToTrade)
	newBot.Leverage = encrypt(newBot.Leverage)

	//set timestamp
	newBot.Timestamp = time.Now().Format("2006-01-02_15:04:05_-0700")
	if isPutReq {
		newBot.CreationDate = reqBot.CreationDate
	} else {
		newBot.CreationDate = newBot.Timestamp
	}

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
			x.AccountRiskPercPerTrade = decrypt(x.AccountRiskPercPerTrade)
		}
		if isBase64(x.AccountSizePercToTrade) {
			x.AccountSizePercToTrade = decrypt(x.AccountSizePercToTrade)
		}
		if isBase64(x.Leverage) {
			x.Leverage = decrypt(x.Leverage)
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

	newQuery := datastore.NewQuery("Bot").Filter("AggregateID =", reqBotData.AggregateID)
	g := client.Run(ctx, newQuery)
	previousBot := parseBotsQueryRes(g)

	if previousBot[0].IsActive != reqBotData.IsActive {
		if reqBotData.IsActive {
			activateBot(reqBotData)
			_, file, line, _ := runtime.Caller(0)
			go Log("activate", fmt.Sprintf("<%v> %v", line, file))
		} else {
			shutdownBot(reqBotData)
			_, file, line, _ := runtime.Caller(0)
			go Log("shutdown", fmt.Sprintf("<%v> %v", line, file))
		}
	}

	if previousBot[0].AccountRiskPercPerTrade != reqBotData.AccountRiskPercPerTrade ||
		previousBot[0].AccountSizePercToTrade != reqBotData.AccountSizePercToTrade ||
		previousBot[0].Leverage != reqBotData.Leverage ||
		previousBot[0].Ticker != reqBotData.Ticker ||
		previousBot[0].ExchangeConnection != reqBotData.ExchangeConnection {
		editBot(reqBotData)
		_, file, line, _ := runtime.Caller(0)
		go Log("edit", fmt.Sprintf("<%v> %v", line, file))
	}

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
