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
var redisHostMsngr = os.Getenv("REDISHOST_MSNGR")
var redisPortMsngr = os.Getenv("REDISPORT_MSNGR")
var redisPassMsngr = os.Getenv("REDISPASS_MSNGR")
var rdbMsngr *redis.Client
var redisHostChartmaster = os.Getenv("REDISHOST_CM")
var redisPortChartmaster = os.Getenv("REDISPORT_CM")
var redisPassChartmaster = os.Getenv("REDISPASS_CM")
var rdbChartmaster *redis.Client
var client *datastore.Client
var ctx context.Context

var periodDurationMap = map[string]time.Duration{}
var httpTimeFormat string

//websockets
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

//all connected clients (url map to *websocket.Conn)
var wsConnections map[string]*websocket.Conn
var wsConnectionsChartmaster map[string]*websocket.Conn

func main() {
	httpTimeFormat = "2006-01-02T15:04:05"

	rand.Seed(time.Now().UTC().UnixNano())

	wsConnections = make(map[string]*websocket.Conn)
	wsConnectionsChartmaster = make(map[string]*websocket.Conn)

	initRedis()
	initDatastore()

	// http.Handle("/", http.FileServer(http.Dir(".")))
	// http.HandleFunc("/create-checkout-session", createCheckoutSession)
	// addr := "localhost:4243"
	// log.Printf("Listening on %s", addr)
	// log.Fatal(http.ListenAndServe(addr, nil))

	periodDurationMap["1MIN"] = 1 * time.Minute
	periodDurationMap["2MIN"] = 2 * time.Minute
	periodDurationMap["3MIN"] = 3 * time.Minute
	periodDurationMap["4MIN"] = 4 * time.Minute
	periodDurationMap["5MIN"] = 5 * time.Minute
	periodDurationMap["6MIN"] = 6 * time.Minute
	periodDurationMap["10MIN"] = 10 * time.Minute
	periodDurationMap["15MIN"] = 15 * time.Minute
	periodDurationMap["20MIN"] = 20 * time.Minute
	periodDurationMap["30MIN"] = 30 * time.Minute
	periodDurationMap["1HRS"] = 1 * time.Hour
	periodDurationMap["2HRS"] = 2 * time.Hour
	periodDurationMap["3HRS"] = 3 * time.Hour
	periodDurationMap["4HRS"] = 4 * time.Hour
	periodDurationMap["6HRS"] = 6 * time.Hour
	periodDurationMap["8HRS"] = 8 * time.Hour
	periodDurationMap["12HRS"] = 12 * time.Hour
	periodDurationMap["1DAY"] = 24 * time.Hour
	periodDurationMap["2DAY"] = 48 * time.Hour

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
	router.Methods("GET", "OPTIONS").Path("/ws-cm/{id}").HandlerFunc(wsChartmasterConnectHandler)

	router.Methods("GET", "OPTIONS").Path("/stacked").HandlerFunc(stackedHandler)
	router.Methods("GET", "OPTIONS").Path("/pie").HandlerFunc(pieHandler)
	router.Methods("GET", "OPTIONS").Path("/scatter").HandlerFunc(scatterHandler)

	// router.Methods("GET", "OPTIONS").Path("/candlestick").HandlerFunc(indexChartmasterHandler)
	// router.Methods("GET", "OPTIONS").Path("/profitCurve").HandlerFunc(profitCurveHandler)
	// router.Methods("GET", "OPTIONS").Path("/simulatedTrades").HandlerFunc(simulatedTradesHandler)
	router.Methods("POST", "OPTIONS").Path("/backtest").HandlerFunc(backtestHandler)
	router.Methods("POST", "OPTIONS").Path("/shareresult").HandlerFunc(shareResultHandler)
	router.Methods("GET", "OPTIONS").Path("/getshareresult").HandlerFunc(getShareResultHandler)
	router.Methods("GET", "OPTIONS").Path("/getChartmasterTickers").HandlerFunc(getTickersHandler)
	router.Methods("GET", "OPTIONS").Path("/backtestHistory").HandlerFunc(getBacktestHistoryHandler)
	router.Methods("GET", "OPTIONS").Path("/backtestHistory/{id}").HandlerFunc(getBacktestResHandler)

	msngr.GoogleProjectID = "myika-anastasia"
	msngr.InitRedis(redisHostMsngr, redisPortMsngr, redisPassMsngr)

	port := os.Getenv("PORT")
	fmt.Println("api-gateway listening on port " + port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
