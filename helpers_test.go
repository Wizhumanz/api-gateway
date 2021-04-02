package main

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"cloud.google.com/go/datastore"
)

type testPair struct {
	Inputs   []interface{}
	Expected interface{}
}

var testDelElemData []testPair
var testBase64CheckData []testPair

func setup() {
	testDelElemData = []testPair{
		{
			Inputs: []interface{}{
				[]Bot{
					{K: datastore.IDKey("TEST", 1, nil)},
					{K: datastore.IDKey("TEST", 2, nil)},
					{K: datastore.IDKey("TEST", 3, nil)},
				},
				Bot{K: datastore.IDKey("TEST", 2, nil)},
			},
			Expected: []Bot{
				{K: datastore.IDKey("TEST", 1, nil)},
				{K: datastore.IDKey("TEST", 3, nil)},
			},
		},
		{
			Inputs: []interface{}{
				[]Bot{
					{K: datastore.IDKey("TEST", 23, nil)},
					{K: datastore.IDKey("TEST", 90, nil)},
					{K: datastore.IDKey("TEST", 56, nil)},
				},
				Bot{K: datastore.IDKey("TEST", 56, nil)},
			},
			Expected: []Bot{
				{K: datastore.IDKey("TEST", 23, nil)},
				{K: datastore.IDKey("TEST", 90, nil)},
			},
		},
	}

	testBase64CheckData = []testPair{
		{
			Inputs: []interface{}{
				base64.StdEncoding.EncodeToString([]byte("string")),
			},
			Expected: true,
		},
		{
			Inputs: []interface{}{
				base64.StdEncoding.EncodeToString([]byte("anotherString")),
			},
			Expected: true,
		},
		{
			Inputs: []interface{}{
				"Simon",
			},
			Expected: false,
		},
		{
			Inputs: []interface{}{
				"Testing SUCKS",
			},
			Expected: false,
		},
	}
}

func shutdown() {
	//shutdown
}

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	shutdown()
	os.Exit(code)
}

func TestHelperDeleteElement(t *testing.T) {
	for _, pair := range testDelElemData {
		arg1 := pair.Inputs[0].([]Bot)
		arg2 := pair.Inputs[1].(Bot)
		expected := pair.Expected.([]Bot)
		res := deleteElement(arg1, arg2)

		arg1J, _ := json.MarshalIndent(arg1, "", "  ")
		arg2J, _ := json.MarshalIndent(arg2, "", "  ")
		expectedJ, _ := json.MarshalIndent(expected, "", "  ")
		resJ, _ := json.MarshalIndent(res, "", "  ")
		if !reflect.DeepEqual(res, expected) {
			t.Errorf("Expected deleteElement(%v, %v) to equal %v, actual result = %v",
				string(arg1J),
				string(arg2J),
				string(expectedJ),
				string(resJ))
		}
	}
}

func TestHelperBase64Check(t *testing.T) {
	for _, pair := range testBase64CheckData {
		arg1 := pair.Inputs[0].(string)
		expected := pair.Expected.(bool)
		res := isBase64(arg1)
		if !reflect.DeepEqual(res, expected) {
			t.Errorf("Expected isBase64(%s) to equal %v, actual result = %v", arg1, expected, res)
		}
	}
}
