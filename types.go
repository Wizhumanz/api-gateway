package main

import (
	"fmt"
	"reflect"

	"cloud.google.com/go/datastore"
)

// API types
type createCheckoutSessionResponse struct {
	SessionID string `json:"id"`
}

type jsonResponse struct {
	Msg  string `json:"message"`
	Body string `json:"body"`
}

//for unmarshalling JSON to bools
type JSONBool bool

func (bit *JSONBool) UnmarshalJSON(b []byte) error {
	txt := string(b)
	*bit = JSONBool(txt == "1" || txt == "true")
	return nil
}

type loginReq struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type webHookRequest struct {
	User            string `json:"User"`
	Ticker          string `json:"Ticker"`
	Direction       string `json:"Direction"`
	TradeActionType string `json:"TradeActionType"` // ENTER, EXIT, SL, TP
	Size            string `json:"Size"`
}

type User struct {
	K        *datastore.Key `datastore:"__key__"`
	Name     string         `json:"name"`
	Email    string         `json:"email"`
	Password string         `json:"password"`
}

func (l User) String() string {
	r := ""
	v := reflect.ValueOf(l)
	typeOfL := v.Type()

	for i := 0; i < v.NumField(); i++ {
		r = r + fmt.Sprintf("%s: %v, ", typeOfL.Field(i).Name, v.Field(i).Interface())
	}
	return r
}

type TradeAction struct {
	KEY         string  `json:"KEY"`
	UserID      string  `json:"UserID"`
	Action      string  `json:"Action"`
	AggregateID int     `json:"AggregateID,string"`
	BotID       string  `json:"BotID"`
	Direction   string  `json:"Direction"` //LONG or SHORT
	Size        float32 `json:"Size,string"`
	Timestamp   string  `json:"Timestamp"`
	Ticker      string  `json:"Ticker"`
	Exchange    string  `json:"Exchange"`
}

type Bot struct {
	KEY                     string         `json:"KEY"`
	K                       *datastore.Key `datastore:"__key__"`
	Name                    string         `json:"Name"`
	AggregateID             int            `json:"AggregateID,string"`
	UserID                  string         `json:"UserID"`
	ExchangeConnection      string         `json:"ExchangeConnection"`
	AccountRiskPercPerTrade string         `json:"AccountRiskPercPerTrade"`
	AccountSizePercToTrade  string         `json:"AccountSizePercToTrade"`
	IsActive                bool           `json:"IsActive,string"`
	IsArchived              bool           `json:"IsArchived,string"`
	Leverage                string         `json:"Leverage"`
	Timestamp               string         `json:"Timestamp"`
	Ticker                  string         `json:"Ticker"`
	WebhookConnectionID     string         `json:"WebhookConnectionID"`
	CreationDate            string         `json:"CreationDate"`
}

func (l Bot) String() string {
	r := ""
	v := reflect.ValueOf(l)
	typeOfL := v.Type()

	for i := 0; i < v.NumField(); i++ {
		r = r + fmt.Sprintf("%s: %v, ", typeOfL.Field(i).Name, v.Field(i).Interface())
	}
	return r
}

type ExchangeConnection struct {
	K         *datastore.Key `datastore:"__key__"`
	KEY       string         `json:"KEY"`
	Name      string         `json:"Name"`
	APIKey    string         `json:"APIKey"`
	UserID    string         `json:"UserID"`
	IsDeleted bool           `json:"IsDeleted,string"`
	Timestamp string         `json:"Timestamp"`
}

type WebhookConnection struct {
	K           *datastore.Key `datastore:"__key__"`
	KEY         string         `json:"KEY"`
	URL         string         `json:"URL"`
	Name        string         `json:"Name"`
	Description string         `json:"Description"`
	IsPublic    bool           `json:"IsPublic,string"`
}

func (l WebhookConnection) String() string {
	r := ""
	v := reflect.ValueOf(l)
	typeOfL := v.Type()

	for i := 0; i < v.NumField(); i++ {
		r = r + fmt.Sprintf("%s: %v, ", typeOfL.Field(i).Name, v.Field(i).Interface())
	}
	return r
}

type ScatterData struct {
	Profit   float64 `json:"Profit"`
	Duration float64 `json:"Duration"`
	Size     int     `json:"Size"`
	Leverage int     `json:"Leverage"`
	Time     int     `json:"Time"`
}

type ChartmasterData struct {
	Date  string  `json:"Date"`
	Open  float64 `json:"Open"`
	High  float64 `json:"High"`
	Low   float64 `json:"Low"`
	Close float64 `json:"Close"`
}
