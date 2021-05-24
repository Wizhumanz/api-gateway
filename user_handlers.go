package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/gorilla/mux"
	"google.golang.org/api/iterator"
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

	addUser(w, r, false, newUser)
	// // create password hash
	// newUser.Password, _ = HashPassword(newUser.Password)
	// // create encrypt key of fixed length
	// rand.Seed(time.Now().UnixNano())

	// // create new listing in DB
	// kind := "User"
	// newUserKey := datastore.IncompleteKey(kind, nil)
	// addedKey, err := client.Put(ctx, newUserKey, &newUser)
	// if err != nil {
	// 	log.Fatalf("Failed to save User: %v", err)
	// }

	// // return
	// data := jsonResponse{
	// 	Msg:  "Added " + newUserKey.String(),
	// 	Body: fmt.Sprint(addedKey.ID),
	// }
	// w.Header().Set("Content-Type", "application/json")
	// w.WriteHeader(http.StatusCreated)
	// json.NewEncoder(w).Encode(data)
}

func updateUser(w http.ResponseWriter, r *http.Request) {
	setupCORS(&w, r)
	if (*r).Method == "OPTIONS" {
		return
	}

	if flag.Lookup("test.v") != nil {
		initDatastore()
	}

	usersResp := make([]User, 0)

	// auth, _ := url.QueryUnescape(r.Header.Get("Authorization"))
	// authReq := loginReq{
	// 	ID:       r.URL.Query()["user"][0],
	// 	Password: auth,
	// }
	// authSuccess, _ := authenticateUser(authReq)
	// if !authSuccess {
	// 	data := jsonResponse{Msg: "Authorization Invalid", Body: "Go away."}
	// 	w.WriteHeader(http.StatusUnauthorized)
	// 	json.NewEncoder(w).Encode(data)
	// 	return
	// }

	// if updating User, don't allow AggregateID change
	var reqUserData User
	err := json.NewDecoder(r.Body).Decode(&reqUserData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if reqUserData.AggregateID != 0 {
		data := jsonResponse{Msg: "ID property of User is immutable.", Body: "Do not pass ID property in request body, instead pass in URL."}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(data)
		return
	}

	//check if user already exists to update
	userToUpdateID, unescapeErr := url.QueryUnescape(mux.Vars(r)["id"]) //aggregate ID, not DB __key__
	if unescapeErr != nil {
		data := jsonResponse{Msg: "User ID Parse Error", Body: unescapeErr.Error()}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(data)
		return
	}
	int, _ := strconv.Atoi(userToUpdateID)
	query := datastore.NewQuery("User").
		Filter("AggregateID =", int)
	t := client.Run(ctx, query)
	for {
		var x User
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

		usersResp = append(usersResp, x)
	}

	//return if user to update doesn't exist
	putIDValid := len(usersResp) > 0
	if !putIDValid {
		data := jsonResponse{Msg: "User ID Invalid", Body: "User with provided ID does not exist."}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(data)
		return
	}

	//must assign aggregate ID from existing User
	reqUserData.AggregateID = usersResp[len(usersResp)-1].AggregateID

	addUser(w, r, true, reqUserData)
}

func addUser(w http.ResponseWriter, r *http.Request, isPutReq bool, reqUser User) {
	setupCORS(&w, r)
	if (*r).Method == "OPTIONS" {
		return
	}

	if flag.Lookup("test.v") != nil {
		initDatastore()
	}
	newUser := reqUser

	// if updating, name field not passed in JSON body, so must fill
	if isPutReq {
		newUser.AggregateID = reqUser.AggregateID
	} else {
		// else increment aggregate ID
		var x User
		//get highest aggregate ID
		query := datastore.NewQuery("User").
			Project("AggregateID").
			Order("-AggregateID")
		t := client.Run(ctx, query)
		_, error := t.Next(&x)
		if error != nil {
			// Handle error.
		}
		newUser.AggregateID = x.AggregateID + 1
	}
	// create password hash
	newUser.Password, _ = HashPassword(newUser.Password)
	// create encrypt key of fixed length
	rand.Seed(time.Now().UnixNano())

	//set timestamp
	newUser.Timestamp = time.Now().Format("2006-01-02_15:04:05_-0700")

	// create new listing in DB
	kind := "User"
	newUserKey := datastore.IncompleteKey(kind, nil)
	_, err := client.Put(ctx, newUserKey, &newUser)
	if err != nil {
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

func getUser(w http.ResponseWriter, r *http.Request) {
	setupCORS(&w, r)
	if (*r).Method == "OPTIONS" {
		return
	}

	if flag.Lookup("test.v") != nil {
		initDatastore()
	}

	email := r.URL.Query()["email"][0]

	if !(len(email) > 0) {
		data := jsonResponse{Msg: "IDs array param empty.", Body: "Pass ids property in json as array of strings."}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(data)
		return
	}
	// Get user from database
	query := datastore.NewQuery("User").
		Filter("Email =", email)
	t := client.Run(ctx, query)
	// for {
	var user User
	t.Next(&user)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

// func parseUserQueryRes(t *datastore.Iterator) []User {
// 	var usersResp []User
// 	for {
// 		var x User
// 		key, err := t.Next(&x)
// 		if key != nil {
// 			x.KEY = fmt.Sprint(key.ID)
// 		}
// 		if err == iterator.Done {
// 			break
// 		}

// 		//event sourcing (pick latest snapshot)
// 		if len(usersResp) == 0 {
// 			usersResp = append(usersResp, x)
// 		} else {
// 			//find User in existing array
// 			var exUser User
// 			for _, b := range usersResp {
// 				if b.AggregateID == x.AggregateID {
// 					exUser = b
// 				}
// 			}

// 			//if User exists, append row/entry with the latest timestamp
// 			if exUser.AggregateID != 0 || exUser.Timestamp != "" {
// 				//compare timestamps
// 				layout := "2006-01-02_15:04:05_-0700"
// 				existingUserTime, _ := time.Parse(layout, exUser.Timestamp)
// 				newUserTime, _ := time.Parse(layout, x.Timestamp)
// 				//if existing is older, remove it and add newer current listing; otherwise, do nothing
// 				if existingUserTime.Before(newUserTime) {
// 					//rm existing listing
// 					usersResp = deleteElementUser(usersResp, exUser)
// 					//append current listing
// 					usersResp = append(usersResp, x)
// 				}
// 			} else {
// 				//otherwise, just append newly decoded (so far unique) User
// 				usersResp = append(usersResp, x)
// 			}
// 		}
// 	}

// 	return usersResp
// }
