package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http/httptest"
	"testing"

	"cloud.google.com/go/datastore"
	"google.golang.org/api/iterator"
)

func TestHandlerCreateNewUser(t *testing.T) {
	values := map[string]string{"name": "JOHN DOE", "email": "VEGGIE@VEGGIE.COM", "password": "supersoaker"}
	json_data, err := json.Marshal(values)

	if err != nil {
		log.Fatal(err)
	}

	req := httptest.NewRequest("POST", "/user", bytes.NewBuffer(json_data))
	w := httptest.NewRecorder()
	createNewUserHandler(w, req)

	resp := w.Result()

	if resp.StatusCode != 201 {
		t.Error("Expected status code to equal 201")
	} else {
		ctx := context.Background()
		client, err := datastore.NewClient(ctx, googleProjectID)
		if err != nil {
			// TODO: Handle error.
			log.Fatal(err)
		}

		query := datastore.NewQuery("User").Filter("Name =", "JOHN DOE")

		//run query
		tds := client.Run(ctx, query)
		var x User
		for {
			_, err := tds.Next(&x)
			if err == iterator.Done {
				break
			}
		}

		if x.Name != "JOHN DOE" {
			t.Error("Expected new User name to be defined")
		}

		//cleanup DEL user
		key := datastore.IDKey("User", x.K.ID, nil)
		if err := client.Delete(ctx, key); err != nil {
			// TODO: Handle error.
			log.Fatal(err)
		}
	}
}
