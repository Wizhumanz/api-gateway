package main

import (
	"os"
	"reflect"
	"testing"

	"cloud.google.com/go/datastore"
)

type testPair struct {
	Inputs   []interface{}
	Expected interface{}
}

var testData []testPair

func setup() {
	testData = []testPair{
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
	for _, pair := range testData {
		arg1 := pair.Inputs[0].([]Bot)
		arg2 := pair.Inputs[1].(Bot)
		expected := pair.Expected.([]Bot)
		res := deleteElement(arg1, arg2)
		if !reflect.DeepEqual(res, expected) {
			t.Errorf("Expected deleteElement(%s, %s) to equal %s, actual result = %s", arg1, arg2, expected, res)
		}
	}
}
