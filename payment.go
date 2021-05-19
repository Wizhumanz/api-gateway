package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/stripe/stripe-go/v71"
	"github.com/stripe/stripe-go/v71/checkout/session"
)

func createCheckoutSessionSecondTier(w http.ResponseWriter, req *http.Request) {
	//Payment processing
	stripe.Key = "sk_test_51IDiEqIjS4SHzVxyreZ8FjYJLU9DkBhK0ilRjCDJ9q4pTzHNJZ3rE79E0RY8rZzAJVsqMzhaki83AbHO4zOYvtFB00FxM7Tid0"

	setupCORS(&w, req)
	if (*req).Method == "OPTIONS" {
		return
	}

	domain := "http://localhost:3000"
	params := &stripe.CheckoutSessionParams{
		PaymentMethodTypes: stripe.StringSlice([]string{
			"card",
		}),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			&stripe.CheckoutSessionLineItemParams{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String(string(stripe.CurrencyUSD)),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String("Silver Tier"),
					},
					UnitAmount: stripe.Int64(14900),
				},
				Quantity: stripe.Int64(1),
			},
		},
		Mode:       stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL: stripe.String(domain + "/"),
		CancelURL:  stripe.String(domain + "/cancel.html"),
	}

	session, err := session.New(params)

	if err != nil {
		log.Printf("session.New: %v", err)
	}

	data := createCheckoutSessionResponse{
		SessionID: session.ID,
	}
	fmt.Println(session.ID)

	js, _ := json.Marshal(data)
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func createCheckoutSessionThirdTier(w http.ResponseWriter, req *http.Request) {
	//Payment processing
	stripe.Key = "sk_test_51IDiEqIjS4SHzVxyreZ8FjYJLU9DkBhK0ilRjCDJ9q4pTzHNJZ3rE79E0RY8rZzAJVsqMzhaki83AbHO4zOYvtFB00FxM7Tid0"

	setupCORS(&w, req)
	if (*req).Method == "OPTIONS" {
		return
	}

	domain := "http://localhost:3000"
	params := &stripe.CheckoutSessionParams{
		PaymentMethodTypes: stripe.StringSlice([]string{
			"card",
		}),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			&stripe.CheckoutSessionLineItemParams{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String(string(stripe.CurrencyUSD)),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String("Gold Tier"),
					},
					UnitAmount: stripe.Int64(33300),
				},
				Quantity: stripe.Int64(1),
			},
		},
		Mode:       stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL: stripe.String(domain + "/"),
		CancelURL:  stripe.String(domain + "/cancel.html"),
	}

	session, err := session.New(params)

	if err != nil {
		log.Printf("session.New: %v", err)
	}

	data := createCheckoutSessionResponse{
		SessionID: session.ID,
	}
	fmt.Println(session.ID)

	js, _ := json.Marshal(data)
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}
