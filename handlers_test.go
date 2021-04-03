package main

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandlerGetAllBots(t *testing.T) {
	// type ColorGroup struct {
	// 	ID     int
	// 	Name   string
	// 	Colors []string
	// }
	// group := ColorGroup{
	// 	ID:     1,
	// 	Name:   "Reds",
	// 	Colors: []string{"Crimson", "Red", "Ruby", "Maroon"},
	// }
	// b, err := json.Marshal(group)

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
			if bot.K.ID == 0 {
				t.Error("Expected handler to return Bot structs with DB key")
			}
		}
	}

	// fmt.Println(resp.Header.Get("Content-Type")
}
