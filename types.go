package main

import (
	"fmt"
	"reflect"
	"strconv"
	"time"

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

type CandlestickChartData struct {
	DateTime        string  `json:"DateTime"`
	Open            float64 `json:"Open"`
	High            float64 `json:"High"`
	Low             float64 `json:"Low"`
	Close           float64 `json:"Close"`
	StratEnterPrice float64 `json:"StratEnterPrice"`
	StratExitPrice  float64 `json:"StratExitPrice"`
	Label           string  `json:"Label"`
}

type ProfitCurveDataPoint struct {
	DateTime string  `json:"DateTime"`
	Equity   float64 `json:"Equity"`
}

type ProfitCurveData struct {
	Label string                 `json:"DataLabel"`
	Data  []ProfitCurveDataPoint `json:"Data"`
}

type SimulatedTradeDataPoint struct {
	DateTime      string  `json:"DateTime"`
	Direction     string  `json:"Direction"`
	EntryPrice    float64 `json:"EntryPrice"`
	ExitPrice     float64 `json:"ExitPrice"`
	PosSize       float64 `json:"PosSize"`
	RiskedEquity  float64 `json:"RiskedEquity"`
	RawProfitPerc float64 `json:"RawProfitPerc"`
}

type SimulatedTradeData struct {
	Label string                    `json:"DataLabel"`
	Data  []SimulatedTradeDataPoint `json:"Data"`
}

type Candlestick struct {
	DateTime    string
	PeriodStart string  `json:"time_period_start"`
	PeriodEnd   string  `json:"time_period_end"`
	TimeOpen    string  `json:"time_open"`
	TimeClose   string  `json:"time_close"`
	Open        float64 `json:"price_open"`
	High        float64 `json:"price_high"`
	Low         float64 `json:"price_low"`
	Close       float64 `json:"price_close"`
	Volume      float64 `json:"volume_traded"`
	TradesCount float64 `json:"trades_count"`
}

func (c *Candlestick) Create(redisData map[string]string) {
	c.Open, _ = strconv.ParseFloat(redisData["open"], 32)
	c.High, _ = strconv.ParseFloat(redisData["high"], 32)
	c.Low, _ = strconv.ParseFloat(redisData["low"], 32)
	c.Close, _ = strconv.ParseFloat(redisData["close"], 32)
	c.Volume, _ = strconv.ParseFloat(redisData["volume"], 32)
	c.TradesCount, _ = strconv.ParseFloat(redisData["tradesCount"], 32)
	c.TimeOpen = redisData["timeOpen"]
	c.TimeClose = redisData["timeClose"]
	c.PeriodStart = redisData["periodStart"]
	c.PeriodEnd = redisData["periodEnd"]
	t, timeErr := time.Parse(redisKeyTimeFormat, redisData["periodStart"])
	if timeErr != nil {
		fmt.Println(timeErr)
		return
	}
	c.DateTime = t.Format(httpTimeFormat)
}

type StrategySimulatorAction struct {
	Action string
	Price  float64
	SL     float64
}

type StrategySimulator struct {
	PosLongSize     float64
	PosShortSize    float64
	totalEquity     float64
	availableEquity float64
	Actions         map[int]StrategySimulatorAction //map bar index to action that occured at that index
}

func (strat *StrategySimulator) Init(e float64) {
	strat.totalEquity = e
	strat.availableEquity = e
	strat.Actions = map[int]StrategySimulatorAction{}
}

func (strat *StrategySimulator) GetEquity() float64 {
	return strat.totalEquity
}

func (strat *StrategySimulator) Buy(price, sl, orderSize float64, directionIsLong bool, cIndex int) {
	// if (orderSize * price) > strat.availableEquity {
	// 	log.Fatal(colorRed + "Order size exceeds available equity" + colorReset)
	// 	return
	// }

	strat.availableEquity = strat.availableEquity - (orderSize * price)

	if directionIsLong {
		strat.PosLongSize = orderSize
	} else {
		strat.PosShortSize = orderSize
	}

	strat.Actions[cIndex] = StrategySimulatorAction{
		Action: "ENTER",
		Price:  price,
		SL:     sl,
	}
}

func (strat *StrategySimulator) CloseLong(price, orderSize float64, cIndex int) {
	//close entire long
	if orderSize == 0 {
		strat.totalEquity = strat.availableEquity + (strat.PosLongSize * price)
		strat.PosLongSize = 0
	} else {
		strat.totalEquity = strat.availableEquity + (orderSize * price)
		strat.PosLongSize = strat.PosLongSize - orderSize
	}
	strat.availableEquity = strat.totalEquity

	strat.Actions[cIndex] = StrategySimulatorAction{
		Action: "SL",
		Price:  price,
	}
}

func (strat *StrategySimulator) CheckPositions(open, high, low, close float64, cIndex int) float64 {
	var sl float64
	if strat.PosLongSize > 0 {
		//get SL
		for _, act := range strat.Actions {
			if act.Action == "ENTER" {
				sl = act.SL
				break
			}
		}
		//check SL
		if low <= sl || close <= sl || open <= sl || high <= sl {
			strat.CloseLong(close, 0, cIndex)
		}

		//TODO: check TP
	}

	return sl
}
