package main

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"math/rand"
	"strconv"
	"time"

	// "encoding/base64"
	"encoding/base64"
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
	Name       string `json:"name"`
	Email      string `json:"email"`
	Password   string `json:"password"`
	EncryptKey string
}

func (l User) String() string {
	r := ""
	v := reflect.ValueOf(l)
	typeOfL := v.Type()

	for i := 0; i < v.NumField(); i++ {
		r = r + fmt.Sprintf("%s: %v, ", typeOfL.Field(i).Name, v.Field(i).Interface())
	}
	return r
}

type Bot struct {
	KEY                     string `json:"KEY,omitempty"`
	Name                    string `json:"Name"`
	AggregateID             int    `json:"AggregateID,string"`
	UserID                  string `json:"UserID"`
	ExchangeConnection      string `json:"ExchangeConnection"`
	AccountRiskPercPerTrade string `json:"AccountRiskPercPerTrade"`
	AccountSizePercToTrade  string `json:"AccountSizePercToTrade"`
	IsActive                bool   `json:"IsActive,string"`
	IsArchived              bool   `json:"IsArchived,string"`
	Leverage                string `json:"Leverage"`
	WebhookURL              string `json:"WebhookURL"`
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

type TradeAction struct {
	KEY         string  `json:"KEY,omitempty"`
	Action      string  `json:"Action"`
	AggregateID int     `json:"AggregateID,string"`
	BotID       int     `json:"BotID"`
	OrderType   int     `json:"OrderType"`
	Size        float32 `json:"size"`
	TimeStamp   string  `json:"timeStamp"`
}

var googleProjectID = "myika-anastasia"
var redisHost = os.Getenv("REDISHOST")
var redisPort = os.Getenv("REDISPORT")
var redisAddr = fmt.Sprintf("%s:%s", redisHost, redisPort)
var rdb *redis.Client
var client *datastore.Client
var ctx context.Context

// helper funcs

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func generateEncryptKey(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

var iv = []byte{34, 12, 55, 11, 10, 39, 16, 47, 87, 53, 88, 98, 66, 40, 14, 05}

func encodeBase64(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}

func decodeBase64(s string) []byte {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return data
}

func encrypt(key, text string) string {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		panic(err)
	}
	plaintext := []byte(text)
	cfb := cipher.NewCFBEncrypter(block, iv)
	ciphertext := make([]byte, len(plaintext))
	cfb.XORKeyStream(ciphertext, plaintext)
	return encodeBase64(ciphertext)
}

func decrypt(key, text string) string {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		panic(err)
	}
	ciphertext := decodeBase64(text)
	cfb := cipher.NewCFBEncrypter(block, iv)
	plaintext := make([]byte, len(ciphertext))
	cfb.XORKeyStream(plaintext, ciphertext)
	return string(plaintext)
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

func setupCORS(w *http.ResponseWriter, req *http.Request) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*") //temp
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, auth")
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 16)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func authenticateUser(req loginReq) (bool, User) {
	// get user with email
	var userWithEmail User
	query := datastore.NewQuery("User").
		Filter("Email =", req.Email)
	t := client.Run(ctx, query)
	_, error := t.Next(&userWithEmail)
	if error != nil {
		// Handle error.
	}
	// check password hash and return
	return CheckPasswordHash(req.Password, userWithEmail.Password), userWithEmail
}

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
	authSuccess, _ := authenticateUser(newLoginReq)
	if authSuccess {
		data = jsonResponse{
			Msg:  "Successfully logged in!",
			Body: newLoginReq.Email,
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
		Email:    r.URL.Query()["user"][0],
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
	setupCORS(&w, r)
	if (*r).Method == "OPTIONS" {
		return
	}
	botResp := make([]Bot, 0)

	auth, _ := url.QueryUnescape(r.Header.Get("Authorization"))
	authReq := loginReq{
		Email:    r.URL.Query()["user"][0],
		Password: auth,
	}
	authSuccess, reqUser := authenticateUser(authReq)
	if len(r.URL.Query()["isActive"]) == 0 && !authSuccess {
		data := jsonResponse{Msg: "Authorization Invalid", Body: "Go away."}
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(data)
		return
	}

	fmt.Println(r.URL.Query()["user"][0])

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

		//decrypt props
		x.AccountRiskPercPerTrade = decrypt(reqUser.EncryptKey, x.AccountRiskPercPerTrade)
		x.AccountSizePercToTrade = decrypt(reqUser.EncryptKey, x.AccountSizePercToTrade)
		x.Leverage = decrypt(reqUser.EncryptKey, x.Leverage)
		x.Name = decrypt(reqUser.EncryptKey, x.Name)
		x.WebhookURL = decrypt(reqUser.EncryptKey, x.WebhookURL)

		botResp = append(botResp, x)
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(botResp)
}

// almost identical logic with create and update (event sourcing)
func addBot(w http.ResponseWriter, r *http.Request, isPutReq bool, reqBot Bot, reqUser User) {
	var newBot Bot
	newBot = reqBot

	// if updating bot, don't allow AggregateID change
	if isPutReq && (newBot.AggregateID != 0) {
		data := jsonResponse{Msg: "ID property of Bot is immutable.", Body: "Do not pass ID property in request body, instead pass in URL."}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(data)
		return
	}

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
	newBot.Name = encrypt(reqUser.EncryptKey, newBot.Name)
	newBot.AccountRiskPercPerTrade = encrypt(reqUser.EncryptKey, newBot.AccountRiskPercPerTrade)
	newBot.AccountSizePercToTrade = encrypt(reqUser.EncryptKey, newBot.AccountSizePercToTrade)
	newBot.Leverage = encrypt(reqUser.EncryptKey, newBot.Leverage)
	newBot.WebhookURL = encrypt(reqUser.EncryptKey, newBot.WebhookURL)

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
		Email:    r.URL.Query()["user"][0],
		Password: auth,
	}
	authSuccess, _ := authenticateUser(authReq)
	if !authSuccess {
		data := jsonResponse{Msg: "Authorization Invalid", Body: "Go away."}
		w.WriteHeader(http.StatusUnauthorized)
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

	addBot(w, r, true, botsResp[len(botsResp)-1], User{})
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
		Email:    newBot.UserID,
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

func main() {
	// initRedis()

	//init
	ctx = context.Background()
	var err error
	client, err = datastore.NewClient(ctx, googleProjectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	router := mux.NewRouter().StrictSlash(true)
	router.Methods("GET").Path("/").HandlerFunc(indexHandler)
	router.Methods("POST", "OPTIONS").Path("/login").HandlerFunc(loginHandler)
	router.Methods("POST", "OPTIONS").Path("/user").HandlerFunc(createNewUserHandler)
	router.Methods("GET").Path("/trades").HandlerFunc(getAllTradesHandler)
	router.Methods("GET").Path("/bots").HandlerFunc(getAllBotsHandler)
	router.Methods("POST").Path("/bot").HandlerFunc(createNewBotHandler)
	router.Methods("PUT").Path("/bot/{id}").HandlerFunc(updateBotHandler) //pass aggregate ID in URL

	port := os.Getenv("PORT")
	fmt.Println("api-gateway listening on port " + port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
