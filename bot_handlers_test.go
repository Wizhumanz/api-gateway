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

func TestHandlerGetAllBots(t *testing.T) {
	req := httptest.NewRequest("GET", "/bots?user="+"5632499082330112", nil)
	req.Header.Set("Authorization", "trader")
	w := httptest.NewRecorder()
	getAllBotsHandler(w, req)

	resp := w.Result()

	if resp.StatusCode != 200 {
		t.Error("Expected status code to equal 200")
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	newJsonStr := buf.String()
	// fmt.Println(newJsonStr)

	var listOfBots []Bot
	dec := json.NewDecoder(strings.NewReader(newJsonStr))
	err := dec.Decode(&listOfBots)
	if err != nil {
		t.Error("Expected response body to be of type []Bot")
	}
	// for i, bot := range listOfBots {
	// 	fmt.Println(i, bot.K.ID)
	// }
	if len(listOfBots) > 0 {
		for _, bot := range listOfBots {
			if bot.Name == "" {
				t.Error("Expected handler to return Bot structs with Name")
			}
			if bot.AggregateID == 0 {
				t.Error("Expected handler to return Bot structs with AggregateID")
			}
			if bot.UserID == "" {
				t.Error("Expected handler to return Bot structs with UserID")
			}
			if bot.ExchangeConnection == "" {
				t.Error("Expected handler to return Bot structs with ExchangeConnection")
			}
			if bot.AccountRiskPercPerTrade == "" {
				t.Error("Expected handler to return Bot structs with AccountRiskPercPerTrade")
			}
			if bot.AccountSizePercToTrade == "" {
				t.Error("Expected handler to return Bot structs with AccountSizePercToTrade")
			}
			if bot.Leverage == "" {
				t.Error("Expected handler to return Bot structs with Leverage")
			}
			// if bot.WebhookURL == "" {
			// 	t.Error("Expected handler to return Bot structs with WebhookURL")
			// }
			if bot.Timestamp == "" {
				t.Error("Expected handler to return Bot structs with Timestamp")
			}
			if bot.Ticker == "" {
				t.Error("Expected handler to return Bot structs with Ticker")
			}
			if bot.KEY == "" {
				t.Error("Expected handler to return Bot structs with KEY")
			}
			if bot.K.ID == 0 {
				t.Error("Expected handler to return Bot structs with DB key")
			}
		}
	}
}

func TestHandlerCreateNewBot(t *testing.T) {
	values := map[string]string{
		"Name":                    "TAYLOR BOT",
		"UserID":                  "5632499082330112",
		"ExchangeConnection":      "5634161670881280",
		"AccountRiskPercPerTrade": "5",
		"AccountSizePercToTrade":  "12",
		"IsActive":                "true",
		"IsArchived":              "false",
		"Leverage":                "69"}

	json_data, err := json.Marshal(values)

	if err != nil {
		log.Fatal(err)
	}

	req := httptest.NewRequest("POST", "/bot", bytes.NewBuffer(json_data))
	req.Header.Set("Authorization", "trader")
	w := httptest.NewRecorder()
	createNewBotHandler(w, req)

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

		query := datastore.NewQuery("Bot").Filter("Name =", "TAYLOR BOT")

		//run query
		tds := client.Run(ctx, query)
		var x User
		for {
			_, err := tds.Next(&x)
			if err == iterator.Done {
				break
			}
		}

		if x.Name != "TAYLOR BOT" {
			t.Error("Expected new Bot name to be defined")
		}

		key := datastore.IDKey("Bot", x.K.ID, nil)
		if err := client.Delete(ctx, key); err != nil {
			// TODO: Handle error.
			log.Fatal(err)
		}
	}
}
