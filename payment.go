package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/stripe/stripe-go/v71"
	portalsession "github.com/stripe/stripe-go/v71/billingportal/session"
	"github.com/stripe/stripe-go/v71/checkout/session"
	"github.com/stripe/stripe-go/webhook"
)

// func createCheckoutSessionSecondTier(w http.ResponseWriter, req *http.Request) {
// 	//Payment processing
// 	stripe.Key = "sk_test_51IDiEqIjS4SHzVxyreZ8FjYJLU9DkBhK0ilRjCDJ9q4pTzHNJZ3rE79E0RY8rZzAJVsqMzhaki83AbHO4zOYvtFB00FxM7Tid0"

// 	setupCORS(&w, req)
// 	if (*req).Method == "OPTIONS" {
// 		return
// 	}

// 	domain := "http://localhost:3000"
// 	params := &stripe.CheckoutSessionParams{
// 		PaymentMethodTypes: stripe.StringSlice([]string{
// 			"card",
// 		}),
// 		LineItems: []*stripe.CheckoutSessionLineItemParams{
// 			&stripe.CheckoutSessionLineItemParams{
// 				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
// 					Currency: stripe.String(string(stripe.CurrencyUSD)),
// 					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
// 						Name: stripe.String("Silver Tier"),
// 					},
// 					UnitAmount: stripe.Int64(14900),
// 				},
// 				Quantity: stripe.Int64(1),
// 			},
// 		},
// 		Mode:       stripe.String(string(stripe.CheckoutSessionModePayment)),
// 		SuccessURL: stripe.String(domain + "/"),
// 		CancelURL:  stripe.String(domain + "/cancel.html"),
// 	}

// 	session, err := session.New(params)

// 	if err != nil {
// 		log.Printf("session.New: %v", err)
// 	}

// 	data := createCheckoutSessionResponse{
// 		SessionID: session.ID,
// 	}
// 	fmt.Println(session.ID)

// 	js, _ := json.Marshal(data)
// 	w.Header().Set("Content-Type", "application/json")
// 	w.Write(js)
// }

// func createCheckoutSessionThirdTier(w http.ResponseWriter, req *http.Request) {
// 	//Payment processing
// 	stripe.Key = "sk_test_51IDiEqIjS4SHzVxyreZ8FjYJLU9DkBhK0ilRjCDJ9q4pTzHNJZ3rE79E0RY8rZzAJVsqMzhaki83AbHO4zOYvtFB00FxM7Tid0"

// 	setupCORS(&w, req)
// 	if (*req).Method == "OPTIONS" {
// 		return
// 	}

// 	domain := "http://localhost:3000"
// 	params := &stripe.CheckoutSessionParams{
// 		PaymentMethodTypes: stripe.StringSlice([]string{
// 			"card",
// 		}),
// 		LineItems: []*stripe.CheckoutSessionLineItemParams{
// 			&stripe.CheckoutSessionLineItemParams{
// 				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
// 					Currency: stripe.String(string(stripe.CurrencyUSD)),
// 					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
// 						Name: stripe.String("Gold Tier"),
// 					},
// 					UnitAmount: stripe.Int64(33300),
// 				},
// 				Quantity: stripe.Int64(1),
// 			},
// 		},
// 		Mode:       stripe.String(string(stripe.CheckoutSessionModePayment)),
// 		SuccessURL: stripe.String(domain + "/"),
// 		CancelURL:  stripe.String(domain + "/cancel.html"),
// 	}

// 	session, err := session.New(params)

// 	if err != nil {
// 		log.Printf("session.New: %v", err)
// 	}

// 	data := createCheckoutSessionResponse{
// 		SessionID: session.ID,
// 	}
// 	fmt.Println(session.ID)

// 	js, _ := json.Marshal(data)
// 	w.Header().Set("Content-Type", "application/json")
// 	w.Write(js)
// }

// // Set your secret key. Remember to switch to your live secret key in production.
// // See your keys here: https://dashboard.stripe.com/apikeys

func handleSetup(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, struct {
		PublishableKey string `json:"publishableKey"`
		BasicPrice     string `json:"basicPrice"`
		ProPrice       string `json:"proPrice"`
	}{
		PublishableKey: os.Getenv("STRIPE_PUBLISHABLE_KEY"),
		BasicPrice:     os.Getenv("BASIC_PRICE_ID"),
		ProPrice:       os.Getenv("PRO_PRICE_ID"),
	}, nil)
}

func handleCreateCheckoutSession(w http.ResponseWriter, r *http.Request) {
	fmt.Println("1")

	stripe.Key = "sk_test_51IDiEqIjS4SHzVxyreZ8FjYJLU9DkBhK0ilRjCDJ9q4pTzHNJZ3rE79E0RY8rZzAJVsqMzhaki83AbHO4zOYvtFB00FxM7Tid0"

	setupCORS(&w, r)
	if (*r).Method == "OPTIONS" {
		return
	}

	if r.Method != "POST" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Price string `json:"priceId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, nil, err)
		log.Printf("json.NewDecoder.Decode: %v", err)
		return
	}
	domain := "http://localhost:3000"
	params := &stripe.CheckoutSessionParams{
		SuccessURL: stripe.String(domain + "/success?session_id={CHECKOUT_SESSION_ID}"),
		CancelURL:  stripe.String(domain + "/canceled"),
		PaymentMethodTypes: stripe.StringSlice([]string{
			"card",
		}),
		Mode: stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(req.Price),
				Quantity: stripe.Int64(1),
			},
		},
	}

	s, err := session.New(params)
	if err != nil {
		writeJSON(w, nil, err)
		return
	}

	writeJSON(w, struct {
		SessionID string `json:"sessionId"`
	}{
		SessionID: s.ID,
	}, nil)
}

func handleCheckoutSession(w http.ResponseWriter, r *http.Request) {
	fmt.Println("2")

	stripe.Key = "sk_test_51IDiEqIjS4SHzVxyreZ8FjYJLU9DkBhK0ilRjCDJ9q4pTzHNJZ3rE79E0RY8rZzAJVsqMzhaki83AbHO4zOYvtFB00FxM7Tid0"

	setupCORS(&w, r)
	if (*r).Method == "OPTIONS" {
		return
	}

	if r.Method != "GET" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	sessionID := r.URL.Query().Get("sessionId")
	s, err := session.Get(sessionID, nil)
	writeJSON(w, s, err)
}

func handleCustomerPortal(w http.ResponseWriter, r *http.Request) {
	fmt.Println("3")
	stripe.Key = "sk_test_51IDiEqIjS4SHzVxyreZ8FjYJLU9DkBhK0ilRjCDJ9q4pTzHNJZ3rE79E0RY8rZzAJVsqMzhaki83AbHO4zOYvtFB00FxM7Tid0"

	setupCORS(&w, r)
	if (*r).Method == "OPTIONS" {
		return
	}
	if r.Method != "POST" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		SessionID string `json:"sessionId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, nil, err)
		log.Printf("json.NewDecoder.Decode: %v", err)
		return
	}

	// For demonstration purposes, we're using the Checkout session to retrieve the customer ID.
	// Typically this is stored alongside the authenticated user in your database.
	sessionID := req.SessionID
	s, err := session.Get(sessionID, nil)
	if err != nil {
		writeJSON(w, nil, err)
		return
	}
	// The URL to which the user is redirected when they are done managing
	// billing in the portal.
	// returnURL := os.Getenv("DOMAIN")
	// fmt.Println("returnURL")
	params := &stripe.BillingPortalSessionParams{
		Customer:  stripe.String(s.Customer.ID),
		ReturnURL: stripe.String("http://localhost:3000/payment"),
	}
	ps, _ := portalsession.New(params)

	writeJSON(w, struct {
		URL string `json:"url"`
	}{
		URL: ps.URL,
	}, nil)
}

func handleWebhook(w http.ResponseWriter, r *http.Request) {
	stripe.Key = "sk_test_51IDiEqIjS4SHzVxyreZ8FjYJLU9DkBhK0ilRjCDJ9q4pTzHNJZ3rE79E0RY8rZzAJVsqMzhaki83AbHO4zOYvtFB00FxM7Tid0"
	fmt.Println("4")

	if r.Method != "POST" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Printf("ioutil.ReadAll: %v", err)
		return
	}

	event, err := webhook.ConstructEvent(b, r.Header.Get("Stripe-Signature"), "whsec_cPs6tMcNZSQg11DhoW4G5VKlNaGzlck6")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Printf("webhook.ConstructEvent: %v", err)
		return
	}

	switch event.Type {
	case "checkout.session.completed":
		// Payment is successful and the subscription is created.
		// You should provision the subscription and save the customer ID to your database.
		fmt.Println("checkout session completed")
	case "invoice.paid":
		// Continue to provision the subscription as payments continue to be made.
		// Store the status in your database and check when a user accesses your service.
		// This approach helps you avoid hitting rate limits.
		fmt.Println("invoive paid")
	case "invoice.payment_failed":
		// The payment failed or the customer does not have a valid payment method.
		// The subscription becomes past_due. Notify your customer and send them to the
		// customer portal to update their payment information.
		fmt.Println("invoice payment failed")
	case "customer.subscription.updated":
		fmt.Println("customer.subscription.updated")

		if event.Data.Object["cancel_at_period_end"] == false {
			updateUserTier(event.Data.Object["items"].(map[string]interface{})["data"].([]interface{})[0].(map[string]interface{})["plan"].(map[string]interface{})["amount"].(float64))
		}

	case "customer.subscription.deleted":
		fmt.Println("customer.subscription.deleted")
		updateUserCancellation(true)

	default:
		// unhandled event type
		fmt.Println("default")
	}
}

func updateUserTier(tier float64) {
	currentUser[0].Tier = tier

	currentUser[0].Timestamp = time.Now().Format("2006-01-02_15:04:05_-0700")

	kind := "User"
	newUserKey := datastore.IncompleteKey(kind, nil)
	_, err := client.Put(ctx, newUserKey, &currentUser[0])
	if err != nil {
		log.Fatalf("Failed to save User: %v", err)
	}
}

func updateUserCancellation(cancel bool) {
	currentUser[0].Cancellation = cancel

	currentUser[0].Timestamp = time.Now().Format("2006-01-02_15:04:05_-0700")

	kind := "User"
	newUserKey := datastore.IncompleteKey(kind, nil)
	_, err := client.Put(ctx, newUserKey, &currentUser[0])
	if err != nil {
		log.Fatalf("Failed to save User: %v", err)
	}
}

type errResp struct {
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

func writeJSON(w http.ResponseWriter, v interface{}, err error) {
	var respVal interface{}
	if err != nil {
		msg := err.Error()
		var serr *stripe.Error
		if errors.As(err, &serr) {
			msg = serr.Msg
		}
		w.WriteHeader(http.StatusBadRequest)
		var e errResp
		e.Error.Message = msg
		respVal = e
	} else {
		respVal = v
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(respVal); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("json.NewEncoder.Encode: %v", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := io.Copy(w, &buf); err != nil {
		log.Printf("io.Copy: %v", err)
		return
	}
}
