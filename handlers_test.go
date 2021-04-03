package main

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
)

// func TestHandlerCreateNewBot(t *testing.T) {
// 	type ColorGroup struct {
// 		ID     int
// 		Name   string
// 		Colors []string
// 	}
// 	group := ColorGroup{
// 		ID:     1,
// 		Name:   "Reds",
// 		Colors: []string{"Crimson", "Red", "Ruby", "Maroon"},
// 	}
// 	b, err := json.Marshal(group)
// }

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
			if bot.WebhookURL == "" {
				t.Error("Expected handler to return Bot structs with WebhookURL")
			}
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

	// fmt.Println(resp.Header.Get("Content-Type")

}
func TestHandlerGetAllTrades(t *testing.T) {
	req := httptest.NewRequest("GET", "/trades?user="+"5632499082330112", nil)
	req.Header.Set("Authorization", "trader")
	w := httptest.NewRecorder()
	getAllTradesHandler(w, req)

	resp := w.Result()

	if resp.StatusCode != 200 {
		t.Error("Expected status code to equal 200")
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	newJsonStr := buf.String()
	// fmt.Println(newJsonStr)

	var listOfTrades []TradeAction
	dec := json.NewDecoder(strings.NewReader(newJsonStr))
	err := dec.Decode(&listOfTrades)
	if err != nil {
		t.Error("Expected response body to be of type []TradeAction")
	}
	// for i, bot := range listOfBots {
	// 	fmt.Println(i, bot.K.ID)
	// }
	if len(listOfTrades) > 0 {
		for _, trade := range listOfTrades {
			if trade.Action == "" {
				t.Error("Expected handler to return TradeAction structs with Action")
			}
			if trade.AggregateID == 0 {
				t.Error("Expected handler to return TradeAction structs with AggregateID")
			}
			if trade.BotID == "" {
				t.Error("Expected handler to return TradeAction structs with BotID")
			}
			if trade.OrderType != 0 && trade.OrderType != 1 {
				t.Error("Expected handler to return TradeAction structs with OrderType")
			}
			if trade.Size == 0 {
				t.Error("Expected handler to return TradeAction structs with Size")
			}
			if trade.TimeStamp == "" {
				t.Error("Expected handler to return TradeAction structs with TimeStamp")
			}
			if trade.Ticker == "" {
				t.Error("Expected handler to return TradeAction structs with Ticker")
			}
			if trade.Exchange == "" {
				t.Error("Expected handler to return TradeAction structs with Exchange")
			}
			if trade.KEY == "" {
				t.Error("Expected handler to return TradeAction structs with KEY")
			}
		}
	}

	// fmt.Println(resp.Header.Get("Content-Type")

}

func TestHandlerGetAllExchangeConnections(t *testing.T) {
	req := httptest.NewRequest("GET", "/trades?user="+"5632499082330112", nil)
	req.Header.Set("Authorization", "trader")
	w := httptest.NewRecorder()
	getAllExchangeConnectionsHandler(w, req)

	resp := w.Result()

	if resp.StatusCode != 200 {
		t.Error("Expected status code to equal 200")
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	newJsonStr := buf.String()
	// fmt.Println(newJsonStr)

	var listOfExchanges []ExchangeConnection
	dec := json.NewDecoder(strings.NewReader(newJsonStr))
	err := dec.Decode(&listOfExchanges)
	if err != nil {
		t.Error("Expected response body to be of type []ExchangeConnection")
	}
	// for i, bot := range listOfBots {
	// 	fmt.Println(i, bot.K.ID)
	// }
	if len(listOfExchanges) > 0 {
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

	// fmt.Println(resp.Header.Get("Content-Type")

}
