package main

import (

	// "io"

	// "net/http"
	"net/http/httptest"
	"testing"
)

func TestIndex(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	indexHandler(w, req)

	resp := w.Result()
	// body, _ := ioutil.ReadAll(resp.Body)
	// fmt.Println(resp.Header.Get("Content-Type"))
	if resp.StatusCode != 200 {
		t.Error("Expected status code to equal 200")
	}
}
