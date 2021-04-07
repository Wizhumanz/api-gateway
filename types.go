package main

import (
	"fmt"
	"reflect"

	"cloud.google.com/go/datastore"
)

// API types

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
	Msg  string `json:"message"`
	Size string `json:"size"`
	User string `json:"user"`
}

type User struct {
	K          *datastore.Key `datastore:"__key__"`
	Name       string         `json:"name"`
	Email      string         `json:"email"`
	Password   string         `json:"password"`
	EncryptKey string
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
	KEY         string  `json:"KEY,omitempty"`
	UserID      string  `json:"UserID,omitempty"`
	Action      string  `json:"Action"`
	AggregateID int     `json:"AggregateID,string"`
	BotID       string  `json:"BotID"`
	OrderType   int     `json:"OrderType"`
	Size        float32 `json:"Size"`
	Timestamp   string  `json:"Timestamp"`
	Ticker      string  `json:"Ticker"`
	Exchange    string  `json:"Exchange"`
}

type Bot struct {
	KEY                     string         `json:"KEY,omitempty"`
	K                       *datastore.Key `datastore:"__key__"`
	Name                    string         `json:"Name,omitempty"`
	AggregateID             int            `json:"AggregateID,string,omitempty"`
	UserID                  string         `json:"UserID,omitempty"`
	ExchangeConnection      string         `json:"ExchangeConnection,omitempty"`
	AccountRiskPercPerTrade string         `json:"AccountRiskPercPerTrade,omitempty"`
	AccountSizePercToTrade  string         `json:"AccountSizePercToTrade,omitempty"`
	IsActive                bool           `json:"IsActive,string,omitempty"`
	IsArchived              bool           `json:"IsArchived,string,omitempty"`
	Leverage                string         `json:"Leverage,omitempty"`
	Timestamp               string         `json:"Timestamp,omitempty"`
	Ticker                  string         `json:"Ticker,omitempty"`
	WebhookConnectionID     string         `json:"WebhookConnectionID"`
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
	URL         string         `json:"URL"`
	Name        string         `json:"Name"`
	Description string         `json:"Description"`
	IsPublic    bool           `json:"IsPublic,string"`
}
