package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

func saveJsonToRedis() {
	data, err := ioutil.ReadFile("./fakeLong.json")
	if err != nil {
		fmt.Print(err)
	}

	var jStruct []RawOHLCGetResp
	json.Unmarshal(data, &jStruct)
	for _, c := range jStruct {
		fmt.Println(c)
	}
}
