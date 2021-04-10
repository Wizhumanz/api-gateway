package main

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandlerGetAllTrades(t *testing.T) {
	req := httptest.NewRequest("GET", "/trades?user="+"5632499082330112", nil)
	req.Header.Set("Authorization", "trader")
	w := httptest.NewRecorder()
	getAllTradesHandler(w, req)
	resp := w.Result()

	if resp.StatusCode != 200 {
		t.Error("Expected status code to equal 200")
	}

	//check body
	var listOfTrades []TradeAction
	dec := json.NewDecoder(strings.NewReader(decodeRespBody(resp)))
	err := dec.Decode(&listOfTrades)
	if err != nil {
		t.Error("Expected response body to be of type []TradeAction")
	}
	if len(listOfTrades) <= 0 {
		t.Error("Expected response of type []TradeAction to have length > 0")
	} else {
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
			if trade.Timestamp == "" {
				t.Error("Expected handler to return TradeAction structs with Timestamp")
			}
			if trade.Ticker == "" {
				t.Error("Expected handler to return TradeAction structs with Ticker")
			}
		}
	}
}
