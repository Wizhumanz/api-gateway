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

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
)

var googleProjectID = "myika-anastasia"
var redisHost = os.Getenv("REDISHOST")
var redisPort = os.Getenv("REDISPORT")
var redisAddr = fmt.Sprintf("%s:%s", redisHost, redisPort)
var rdb *redis.Client
var client *datastore.Client
var ctx context.Context

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	// initRedis()
	initDatastore()

	router := mux.NewRouter().StrictSlash(true)
	router.Methods("GET").Path("/").HandlerFunc(indexHandler)
	router.Methods("POST", "OPTIONS").Path("/login").HandlerFunc(loginHandler)
	router.Methods("POST", "OPTIONS").Path("/user").HandlerFunc(createNewUserHandler)
	router.Methods("GET").Path("/trades").HandlerFunc(getAllTradesHandler)
	router.Methods("GET").Path("/bots").HandlerFunc(getAllBotsHandler)
	router.Methods("POST").Path("/bot").HandlerFunc(createNewBotHandler)
	router.Methods("PUT").Path("/bot/{id}").HandlerFunc(updateBotHandler)
	router.Methods("GET").Path("/exchanges").HandlerFunc(getAllExchangeConnectionsHandler)
	router.Methods("POST").Path("/exchange").HandlerFunc(createNewExchangeConnectionHandler)
	router.Methods("DELETE").Path("/exchange/{id}").HandlerFunc(deleteExchangeConnectionHandler)

	router.Methods("POST").Path("/webhook/{id}").HandlerFunc(tvWebhookHandler)

	port := os.Getenv("PORT")
	fmt.Println("api-gateway listening on port " + port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
