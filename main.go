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

var googleProjectID = "myika-anastasia"
var redisHost = os.Getenv("REDISHOST")
var redisPort = os.Getenv("REDISPORT")
var redisAddr = fmt.Sprintf("%s:%s", redisHost, redisPort)
var rdb *redis.Client
var client *datastore.Client
var ws *websocket.Conn
var ctx context.Context

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	initDatastore()

	router := mux.NewRouter().StrictSlash(true)
	router.Methods("GET", "OPTIONS").Path("/").HandlerFunc(indexHandler)
	router.Methods("POST", "OPTIONS").Path("/login").HandlerFunc(loginHandler)
	router.Methods("POST", "OPTIONS").Path("/user").HandlerFunc(createNewUserHandler)
	router.Methods("GET", "OPTIONS").Path("/trades").HandlerFunc(getAllTradesHandler)
	router.Methods("GET", "OPTIONS").Path("/bots").HandlerFunc(getAllBotsHandler)
	router.Methods("GET", "OPTIONS").Path("/webhooks").HandlerFunc(getAllWebhookConnectionHandler)
	router.Methods("GET", "OPTIONS").Path("/webhook").HandlerFunc(getWebhookConnectionHandler)
	router.Methods("POST", "OPTIONS").Path("/bot").HandlerFunc(createNewBotHandler)
	router.Methods("POST", "OPTIONS").Path("/webhook").HandlerFunc(createNewWebhookConnectionHandler)
	router.Methods("PUT", "OPTIONS").Path("/bot/{id}").HandlerFunc(updateBotHandler)
	router.Methods("DELETE", "OPTIONS").Path("/bot/{id}").HandlerFunc(deleteBotHandler)
	router.Methods("GET", "OPTIONS").Path("/exchanges").HandlerFunc(getAllExchangeConnectionsHandler)
	router.Methods("POST", "OPTIONS").Path("/exchange").HandlerFunc(createNewExchangeConnectionHandler)
	router.Methods("DELETE", "OPTIONS").Path("/exchange/{id}").HandlerFunc(deleteExchangeConnectionHandler)

	router.Methods("POST", "OPTIONS").Path("/webhook/{id}").HandlerFunc(tvWebhookHandler)
	router.Methods("GET", "OPTIONS").Path("/ws").HandlerFunc(wsConnectHandler)

	msngr.GoogleProjectID = "myika-anastasia"
	msngr.InitRedis()
	msngr.InitDatastore()

	port := os.Getenv("PORT")
	fmt.Println("api-gateway listening on port " + port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
