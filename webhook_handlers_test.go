package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http/httptest"
	"testing"
)

func checkHandlerResp(
	t *testing.T,
	body map[string]string,
	expectedRespCode int,
	httpMethod string,
	reqUrl string) {
	json_data, err := json.Marshal(body)
	if err != nil {
		log.Fatal(err)
	}

	req := httptest.NewRequest(httpMethod, reqUrl, bytes.NewBuffer(json_data))

	fmt.Println(req.RequestURI)
	fmt.Println(req.URL)

	req.Header.Set("Authorization", "trader")
	w := httptest.NewRecorder()
	tvWebhookHandler(w, req)

	resp := w.Result()

	//DEBUG
	fmt.Println(resp.StatusCode)
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(string(bytes))

	if resp.StatusCode != expectedRespCode {
		t.Error("Expected status code to equal " + fmt.Sprint(expectedRespCode))
	}
}

func TestTvWebhookHandler(t *testing.T) {
	// TODO: pass /wenhook/{id} not working
	// a := map[string]string{
	// 	"User":            "5632499082330112",
	// 	"Ticker":          "DOGEBTC",
	// 	"TradeActionType": "ENTER",
	// 	"Size":            "420.69",
	// }
	b := map[string]string{
		"User":            "5632499082330112",
		"TradeActionType": "ENTER",
		"Size":            "420.69",
	}

	// checkHandlerResp(t, a, 200, "POST", "/webhook/"+"9947727057786880520585822462569327265417975548950560271651548923300901316622722332392179461653417345")
	checkHandlerResp(t, b, 400, "POST", "/webhook/9947727057786880520585822462569327265417975548950560271651548923300901316622722332392179461653417345")
}
