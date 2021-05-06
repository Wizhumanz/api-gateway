package main

import (
	"context"
	"math/rand"
	"time"

	"fmt"

	"log"
	"net/http"

	"os"

	"cloud.google.com/go/datastore"
	"gitlab.com/myikaco/msngr"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var colorReset = "\033[0m"
var colorRed = "\033[31m"
var colorGreen = "\033[32m"
var colorYellow = "\033[33m"
var colorBlue = "\033[34m"
var colorPurple = "\033[35m"
var colorCyan = "\033[36m"
var colorWhite = "\033[37m"

var googleProjectID = "myika-anastasia"
var redisHost = os.Getenv("REDISHOST")
var redisPort = os.Getenv("REDISPORT")
var redisPass = os.Getenv("REDISPASS")
var rdb *redis.Client
var client *datastore.Client
var ctx context.Context

//websockets
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

//all connected clients (url map to *websocket.Conn)
var wsConnections map[string]*websocket.Conn

func main() {
	// go saveJsonToRedis()

	rand.Seed(time.Now().UTC().UnixNano())

	wsConnections = make(map[string]*websocket.Conn)

	initRedis()
	initDatastore()
	// http.Handle("/", http.FileServer(http.Dir(".")))
	// http.HandleFunc("/create-checkout-session", createCheckoutSession)
	// addr := "localhost:4243"
	// log.Printf("Listening on %s", addr)
	// log.Fatal(http.ListenAndServe(addr, nil))

	router := mux.NewRouter().StrictSlash(true)
	router.Methods("GET", "OPTIONS").Path("/").HandlerFunc(indexHandler)
	router.Methods("POST", "OPTIONS").Path("/login").HandlerFunc(loginHandler)
	router.Methods("POST", "OPTIONS").Path("/user").HandlerFunc(createNewUserHandler)
	router.Methods("POST", "OPTIONS").Path("/trade").HandlerFunc(createNewTradeHandler)
	router.Methods("GET", "OPTIONS").Path("/trades").HandlerFunc(getAllTradesHandler)
	router.Methods("GET", "OPTIONS").Path("/bots").HandlerFunc(getAllBotsHandler)
	router.Methods("GET", "OPTIONS").Path("/bot").HandlerFunc(getBotHandler)
	router.Methods("GET", "OPTIONS").Path("/webhooks").HandlerFunc(getAllWebhookConnectionHandler)
	router.Methods("GET", "OPTIONS").Path("/webhook").HandlerFunc(getWebhookConnectionHandler)
	router.Methods("POST", "OPTIONS").Path("/bot").HandlerFunc(createNewBotHandler)
	router.Methods("POST", "OPTIONS").Path("/webhook").HandlerFunc(createNewWebhookConnectionHandler)
	router.Methods("PUT", "OPTIONS").Path("/bot/{id}").HandlerFunc(updateBotHandler)
	router.Methods("DELETE", "OPTIONS").Path("/bot/{id}").HandlerFunc(deleteBotHandler)
	router.Methods("GET", "OPTIONS").Path("/exchanges").HandlerFunc(getAllExchangeConnectionsHandler)
	router.Methods("POST", "OPTIONS").Path("/exchange").HandlerFunc(createNewExchangeConnectionHandler)
	router.Methods("DELETE", "OPTIONS").Path("/exchange/{id}").HandlerFunc(deleteExchangeConnectionHandler)
	router.Methods("POST", "OPTIONS").Path("/create-checkout-session").HandlerFunc(createCheckoutSession)

	router.Methods("POST", "OPTIONS").Path("/webhook/{id}").HandlerFunc(tvWebhookHandler)
	router.Methods("GET", "OPTIONS").Path("/ws/{id}").HandlerFunc(wsConnectHandler)

	router.Methods("GET", "OPTIONS").Path("/stacked").HandlerFunc(stackedHandler)
	router.Methods("GET", "OPTIONS").Path("/pie").HandlerFunc(pieHandler)
	router.Methods("GET", "OPTIONS").Path("/scatter").HandlerFunc(scatterHandler)

	router.Methods("GET", "OPTIONS").Path("/candlestick").HandlerFunc(indexChartmasterHandler)
	router.Methods("GET", "OPTIONS").Path("/profitCurve").HandlerFunc(profitCurveHandler)
	router.Methods("GET", "OPTIONS").Path("/simulatedTrades").HandlerFunc(simulatedTradesHandler)

	msngr.GoogleProjectID = "myika-anastasia"
	msngr.InitRedis()

	port := os.Getenv("PORT")
	fmt.Println("api-gateway listening on port " + port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
