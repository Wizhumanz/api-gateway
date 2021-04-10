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
	//send req
	req := httptest.NewRequest("GET", "/bots?user="+"5632499082330112", nil)
	req.Header.Set("Authorization", "trader")
	w := httptest.NewRecorder()
	getAllBotsHandler(w, req)
	resp := w.Result()

	//check status code
	if resp.StatusCode != 200 {
		t.Error("Expected status code to equal 200")
	}

	//check resp body
	var listOfBots []Bot
	dec := json.NewDecoder(strings.NewReader(decodeRespBody(resp)))
	err := dec.Decode(&listOfBots)
	if err != nil {
		t.Error("Expected response body to be of type []Bot")
	}

	if len(listOfBots) <= 0 {
		t.Error("Expected response of type []Bot to have length > 0")
	} else {
		for _, bot := range listOfBots {
			if bot.Name == "" {
				t.Error("Expected handler to return Bot with Name")
			}
			if bot.AggregateID == 0 {
				t.Error("Expected handler to return Bot with AggregateID")
			}
			if bot.UserID == "" {
				t.Error("Expected handler to return Bot with UserID")
			}
			if bot.ExchangeConnection == "" {
				t.Error("Expected handler to return Bot with ExchangeConnection")
			}
			if bot.AccountRiskPercPerTrade == "" {
				t.Error("Expected handler to return Bot with AccountRiskPercPerTrade")
			}
			if bot.AccountSizePercToTrade == "" {
				t.Error("Expected handler to return Bot with AccountSizePercToTrade")
			}
			if bot.Leverage == "" {
				t.Error("Expected handler to return Bot with Leverage")
			}
			if bot.Timestamp == "" {
				t.Error("Expected handler to return Bot with Timestamp")
			}
			if bot.Ticker == "" {
				t.Error("Expected handler to return Bot with Ticker")
			}
			if bot.KEY == "" {
				t.Error("Expected handler to return Bot with KEY")
			}
		}
	}
}

func TestHandlerCreateNewBot(t *testing.T) {
	values := map[string]string{
		"Name":                    "TAYLOR BOT",
		"UserID":                  "5632499082330112",
		"ExchangeConnection":      "5634161670881280",
		"AccountRiskPercPerTrade": "5.69",
		"AccountSizePercToTrade":  "12.69",
		"IsActive":                "true",
		"IsArchived":              "false",
		"Leverage":                "69",
	}
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
		//check added bot
		ctx := context.Background()
		client, err := datastore.NewClient(ctx, googleProjectID)
		if err != nil {
			// TODO: Handle error.
			log.Fatal(err)
		}
		query := datastore.NewQuery("Bot").Filter("Name =", "TAYLOR BOT")

		//run query
		tds := client.Run(ctx, query)
		var x Bot
		for {
			_, err := tds.Next(&x)
			if err == iterator.Done {
				break
			}
		}

		if x.Name != values["Name"] {
			t.Error("Expected new Bot name to be defined")
		}
		if x.UserID != values["UserID"] {
			t.Error("Expected new Bot UserID to be defined")
		}
		if x.ExchangeConnection != values["ExchangeConnection"] {
			t.Error("Expected new Bot ExchangeConnection to be defined")
		}
		if x.AccountRiskPercPerTrade != encrypt(values["AccountRiskPercPerTrade"]) {
			t.Error("Expected new Bot AccountRiskPercPerTrade to be defined")
		}
		if x.AccountSizePercToTrade != encrypt(values["AccountSizePercToTrade"]) {
			t.Error("Expected new Bot AccountSizePercToTrade to be defined")
		}
		var compIsActive string
		if x.IsActive {
			compIsActive = "true"
		} else {
			compIsActive = "false"
		}
		if compIsActive != values["IsActive"] {
			t.Error("Expected new Bot IsActive to be defined")
		}
		var compIsArchived string
		if x.IsArchived {
			compIsArchived = "true"
		} else {
			compIsArchived = "false"
		}
		if compIsArchived != values["IsArchived"] {
			t.Error("Expected new Bot IsArchived to be defined")
		}
		if x.Leverage != encrypt(values["Leverage"]) {
			t.Error("Expected new Bot Leverage to be defined")
		}

		//cleanup: del bot
		key := datastore.IDKey("Bot", x.K.ID, nil)
		if err := client.Delete(ctx, key); err != nil {
			// TODO: Handle error.
			log.Fatal(err)
		}
	}
}
