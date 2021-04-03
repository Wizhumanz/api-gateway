package main

import (
	"io/ioutil"
	"net/http/httptest"
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
	body, _ := ioutil.ReadAll(resp.Body)
	// fmt.Println(body)
	if resp.StatusCode != 200 {
		t.Error("Expected status code to equal 200")
	}

	// fmt.Println(resp.Header.Get("Content-Type")
}
