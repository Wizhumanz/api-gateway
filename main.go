package main

import (
	"context"
	// "encoding/base64"
	"encoding/json"
	"fmt"

	// "io"
	"log"
	"net/http"

	"net/url"
	"os"
	"reflect"

	"cloud.google.com/go/datastore"
	"google.golang.org/api/iterator"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

// API types

type jsonResponse struct {
	Msg  string `json:"message"`
	Body string `json:"body"`
}

//for unmarshalling JSON to bools
type JSONBool bool

func (bit *JSONBool) UnmarshalJSON(b []byte) error {
	txt := string(b)
	*bit = JSONBool(txt == "1" || txt == "true")
	return nil
}

type loginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type User struct {
	Name        string `json:"name"`
	Email       string `json:"email"`
	AccountType string `json:"type"`
	Password    string `json:"password"`
}

type Bot struct {
	KEY                     string  `json:"KEY,omitempty"`
	AggregateID             int     `json:"AggregateID,string"`
	UserID                  string  `json:"UserID"`
	ExchangeConnection      string  `json:"ExchangeConnection"`
	AccountRiskPercPerTrade float32 `json:"AccountRiskPercPerTrade,string"`
	AccountSizePercToTrade  float32 `json:"AccountSizePercToTrade,string"`
	IsActive                bool    `json:"IsActive,string"`
	IsArchived              bool    `json:"IsArchived,string"`
	Leverage                int     `json:"Leverage,string"`
	WebhookUrl              string  `json:"WebhookURL"`
}

type TradeAction struct {
	KEY         string  `json:"KEY,omitempty"`
	Action      string  `json:"Action"`
	AggregateID int     `json:"AggregateID,string"`
	BotID       int     `json:"BotID"`
	OrderType   int     `json:"OrderType"`
	Size        float32 `json:"size"`
	TimeStamp   string  `json:"timeStamp"`
}

func (l Bot) String() string {
	r := ""
	v := reflect.ValueOf(l)
	typeOfL := v.Type()

	for i := 0; i < v.NumField(); i++ {
		r = r + fmt.Sprintf("%s: %v, ", typeOfL.Field(i).Name, v.Field(i).Interface())
	}
	return r
}

var googleProjectID = "myika-anastasia"
var redisHost = os.Getenv("REDISHOST")
var redisPort = os.Getenv("REDISPORT")
var redisAddr = fmt.Sprintf("%s:%s", redisHost, redisPort)
var rdb *redis.Client

// helper funcs

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 2)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func authenticateUser(req loginReq) bool {
	// get user with email
	ctx := context.Background()
	client, err := datastore.NewClient(ctx, googleProjectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	var userWithEmail User
	query := datastore.NewQuery("User").
		Filter("Email =", req.Email)
	t := client.Run(ctx, query)
	_, error := t.Next(&userWithEmail)
	if error != nil {
		// Handle error.
	}

	// check password hash and return
	return CheckPasswordHash(req.Password, userWithEmail.Password)
}

func initRedis() {
	if redisHost == "" {
		redisHost = "127.0.0.1"
		fmt.Println("Env var nil, using redis dev address -- " + redisHost)
	}
	if redisPort == "" {
		redisPort = "6379"
		fmt.Println("Env var nil, using redis dev port -- " + redisPort)
	}
	fmt.Println("Connected to Redis on " + redisHost + ":" + redisPort)
	rdb = redis.NewClient(&redis.Options{
		Addr: redisHost + ":" + redisPort,
	})
}

// route handlers

func indexHandler(w http.ResponseWriter, r *http.Request) {
	var data jsonResponse
	w.Header().Set("Content-Type", "application/json")
	data = jsonResponse{Msg: "Anastasia API Gateway", Body: "Ready"}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
	// w.Write([]byte(`{"msg": "привет сука"}`))
}

func getAllTradesHandler(w http.ResponseWriter, r *http.Request) {
	tradesResp := make([]TradeAction, 0)

	// authReq := loginReq{
	// 	Email:    r.URL.Query()["user"][0],
	// 	Password: r.Header.Get("auth"),
	// }
	// if !authenticateUser(authReq) {
	// 	data := jsonResponse{Msg: "Authorization Invalid", Body: "Go away."}
	// 	w.WriteHeader(http.StatusUnauthorized)
	// 	json.NewEncoder(w).Encode(data)
	// 	return
	// }

	//configs before running query
	ctx := context.Background()
	client, err := datastore.NewClient(ctx, googleProjectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	var query *datastore.Query
	userIDParam := r.URL.Query()["user"][0]
	query = datastore.NewQuery("TradeAction").Filter("UserID =", userIDParam)

	//run query
	t := client.Run(ctx, query)
	for {
		var x TradeAction
		key, err := t.Next(&x)
		if key != nil {
			x.KEY = key.Name
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
	botResp := make([]Bot, 0)

	// authReq := loginReq{
	// 	Email:    r.URL.Query()["user"][0],
	// 	Password: r.Header.Get("auth"),
	// }
	// if len(r.URL.Query()["isActive"]) == 0 && !authenticateUser(authReq) {
	// 	data := jsonResponse{Msg: "Authorization Invalid", Body: "Go away."}
	// 	w.WriteHeader(http.StatusUnauthorized)
	// 	json.NewEncoder(w).Encode(data)
	// 	return
	// }

	//configs before running query
	ctx := context.Background()
	client, err := datastore.NewClient(ctx, googleProjectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
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
	for {
		var x Bot
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
		fmt.Println(x.String())
		botResp = append(botResp, x)
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(botResp)
}

// almost identical logic with create and update (event sourcing)
func addBot(w http.ResponseWriter, r *http.Request, isPutReq bool, botToUpdate Bot) {
	var newBot Bot

	// decode data
	err := json.NewDecoder(r.Body).Decode(&newBot)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	authReq := loginReq{
		Email:    newBot.UserID,
		Password: r.Header.Get("auth"),
	}
	// for PUT req, user already authenticated outside this function
	if !isPutReq && !authenticateUser(authReq) {
		data := jsonResponse{Msg: "Authorization Invalid", Body: "Go away."}
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(data)
		return
	}

	// if updating bot, don't allow AggregateID change
	if isPutReq && (&newBot.AggregateID != nil) {
		data := jsonResponse{Msg: "ID property of Bot is immutable.", Body: "Do not pass ID property in request body, instead pass in URL."}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(data)
		return
	}
	// if updating, name field not passed in JSON body, so must fill
	if isPutReq {
		newBot.AggregateID = botToUpdate.AggregateID
	}

	// create new bot in DB
	ctx := context.Background()
	clientAdd, err := datastore.NewClient(ctx, googleProjectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	kind := "Bot"
	newBotKey := datastore.IncompleteKey(kind, nil)

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
	botsResp := make([]Bot, 0)

	//auth
	// authReq := loginReq{
	// 	Email:    r.URL.Query()["user"][0],
	// 	Password: r.Header.Get("auth"),
	// }
	// if !authenticateUser(authReq) {
	// 	data := jsonResponse{Msg: "Authorization Invalid", Body: "Go away."}
	// 	w.WriteHeader(http.StatusUnauthorized)
	// 	json.NewEncoder(w).Encode(data)
	// 	return
	// }

	//get bot with aggregate ID
	ctx := context.Background()
	client, err := datastore.NewClient(ctx, googleProjectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	//check if bot already exists to update
	botToUpdateID, unescapeErr := url.QueryUnescape(mux.Vars(r)["id"]) //aggregate ID, not DB __key__
	if unescapeErr != nil {
		data := jsonResponse{Msg: "Bot ID Parse Error", Body: unescapeErr.Error()}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(data)
		return
	}
	query := datastore.NewQuery("Bot").
		Filter("AggregateID =", botToUpdateID)
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
			x.KEY = key.Name
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

	addBot(w, r, true, botsResp[len(botsResp)-1])
}

func createNewBotHandler(w http.ResponseWriter, r *http.Request) {
	addBot(w, r, false, Bot{}) //empty Bot struct passed just for compiler
}

func main() {
	initRedis()

	router := mux.NewRouter().StrictSlash(true)
	router.Methods("GET").Path("/").HandlerFunc(indexHandler)
	router.Methods("GET").Path("/trades").HandlerFunc(getAllTradesHandler)
	router.Methods("GET").Path("/bots").HandlerFunc(getAllBotsHandler)
	router.Methods("POST").Path("/bot").HandlerFunc(createNewBotHandler)
	router.Methods("PUT").Path("/bot/{id}").HandlerFunc(updateBotHandler) //pass aggregate ID in URL

	port := os.Getenv("PORT")
	fmt.Println("api-gateway listening on port " + port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
