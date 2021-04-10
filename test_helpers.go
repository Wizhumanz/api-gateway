package main

import (
	"bytes"
	"net/http"
)

func decodeRespBody(resp *http.Response) string {
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	newJsonStr := buf.String()
	return newJsonStr
}
