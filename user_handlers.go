package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"cloud.google.com/go/datastore"
)

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
