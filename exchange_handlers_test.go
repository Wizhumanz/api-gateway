package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http/httptest"
	"strings"
	"testing"

	"cloud.google.com/go/datastore"
	"google.golang.org/api/iterator"
)

func TestHandlerGetAllExchangeConnections(t *testing.T) {
	req := httptest.NewRequest("GET", "/trades?user="+"5632499082330112", nil)
	req.Header.Set("Authorization", "trader")
	w := httptest.NewRecorder()
	getAllExchangeConnectionsHandler(w, req)
	resp := w.Result()

	if resp.StatusCode != 200 {
		t.Error("Expected status code to equal 200")
	}

	//check body
	var listOfExchanges []ExchangeConnection
	dec := json.NewDecoder(strings.NewReader(decodeRespBody(resp)))
	err := dec.Decode(&listOfExchanges)
	if err != nil {
		t.Error("Expected response body to be of type []ExchangeConnection")
	}

	if len(listOfExchanges) <= 0 {
		t.Error("Expected length of response []ExchangeConnection to be > 0")
	} else {
		for _, exchange := range listOfExchanges {
			if exchange.Name == "" {
				t.Error("Expected handler to return ExchangeConnection structs with Name")
			}
			if exchange.APIKey == "" {
				t.Error("Expected handler to return ExchangeConnection structs with APIKey")
			}
			if exchange.UserID == "" {
				t.Error("Expected handler to return ExchangeConnection structs with UserID")
			}
			if exchange.Timestamp == "" {
				t.Error("Expected handler to return ExchangeConnection structs with Timestamp")
			}
			if exchange.KEY == "" {
				t.Error("Expected handler to return ExchangeConnection structs with KEY")
			}
			if exchange.K.ID == 0 {
				t.Error("Expected handler to return ExchangeConnection structs with DB key")
			}
		}
	}
}

func TestHandlerCreateNewExchangeConnection(t *testing.T) {
	values := map[string]string{
		"Name":      "Test Exchange",
		"APIKey":    "hcuid27495hf727erer98974hfh2f9",
		"UserID":    "5632499082330112",
		"IsDeleted": "false"}

	json_data, err := json.Marshal(values)

	if err != nil {
		log.Fatal(err)
	}

	req := httptest.NewRequest("POST", "/exchange", bytes.NewBuffer(json_data))
	req.Header.Set("Authorization", "trader")
	w := httptest.NewRecorder()
	createNewExchangeConnectionHandler(w, req)

	resp := w.Result()

	if resp.StatusCode != 201 {
		t.Error("Expected status code to equal 201")
	} else {
		ctx := context.Background()
		client, err := datastore.NewClient(ctx, googleProjectID)
		if err != nil {
			// TODO: Handle error.
			log.Fatal(err)
		}

		query := datastore.NewQuery("ExchangeConnection").Filter("Name =", "Test Exchange")

		//run query
		t := client.Run(ctx, query)
		var x ExchangeConnection
		for {
			_, err := t.Next(&x)
			if err == iterator.Done {
				break
			}
		}
		key := datastore.IDKey("ExchangeConnection", x.K.ID, nil)
		if err := client.Delete(ctx, key); err != nil {
			// TODO: Handle error.
			log.Fatal(err)
		}
	}
}

/*
func TestHandlerDeleteExchangeConnection(t *testing.T) {
	// query := datastore.NewQuery("ExchangeConnection").Filter("Name =", "Doge Exchange").Filter("IsDeleted =", "false")

	// //run query
	// g := client.Run(ctx, query)
	// var x ExchangeConnection
	// for {
	// 	_, err := g.Next(&x)
	// 	if err == iterator.Done {
	// 		break
	// 	}
	// }

	req := httptest.NewRequest("DELETE", "/exchange/5418958039547904?user=5632499082330112", nil)
	req.Header.Set("Authorization", "trader")
	w := httptest.NewRecorder()
	deleteExchangeConnectionHandler(w, req)

	resp := w.Result()

	if resp.StatusCode != 201 {
		t.Error("Expected status code to equal 201")
	} else {
		// ctx := context.Background()
		// client, err := datastore.NewClient(ctx, googleProjectID)
		// if err != nil {
		// 	// TODO: Handle error.
		// 	log.Fatal(err)
		// }

		// query := datastore.NewQuery("ExchangeConnection").Filter("Name =", "Doge Exchange")

		// //run query
		// t := client.Run(ctx, query)
		// var c ExchangeConnection
		// for {
		// 	_, err := t.Next(&c)
		// 	if err == iterator.Done {
		// 		break
		// 	}
		// }
		// key := datastore.IDKey("ExchangeConnection", c.K.ID, nil)
		// if err := client.Delete(ctx, key); err != nil {
		// 	// TODO: Handle error.
		// 	log.Fatal(err)
		// }
	}
}
*/
