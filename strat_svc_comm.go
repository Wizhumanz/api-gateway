package main

import (
	"encoding/json"
	"fmt"
	"time"

	"gitlab.com/myikaco/msngr"
)

func activateBot(bot Bot) {
	// add new trade info into stream (triggers other services)
	jsonBot, _ := json.Marshal(bot)

	msgs := []string{}
	msgs = append(msgs, "Timestamp")
	msgs = append(msgs, time.Now().Format("2006-01-02_15:04:05_-0700"))
	msgs = append(msgs, "BotStreamName")
	msgs = append(msgs, fmt.Sprint(bot.KEY))
	msgs = append(msgs, "CMD")
	msgs = append(msgs, "Activate")
	msgs = append(msgs, "Bot")
	msgs = append(msgs, string(jsonBot))

	botStreamMsgs := []string{}
	botStreamMsgs = append(botStreamMsgs, "Timestamp")
	botStreamMsgs = append(botStreamMsgs, time.Now().Format("2006-01-02_15:04:05_-0700"))
	botStreamMsgs = append(botStreamMsgs, "CMD")
	botStreamMsgs = append(botStreamMsgs, "INIT")

	msngr.AddToStream(fmt.Sprint(bot.KEY), botStreamMsgs)
	msngr.AddToStream("activeBots", msgs)
}

func shutdownBot(bot Bot) {
	// add new trade info into stream (triggers other services)
	msgs := []string{}
	msgs = append(msgs, "Timestamp")
	msgs = append(msgs, time.Now().Format("2006-01-02_15:04:05_-0700"))
	msgs = append(msgs, "BotStreamName")
	msgs = append(msgs, fmt.Sprint(bot.KEY))
	msgs = append(msgs, "CMD")
	msgs = append(msgs, "Deactivate")

	botStreamMsgs := []string{}
	botStreamMsgs = append(botStreamMsgs, "Timestamp")
	botStreamMsgs = append(botStreamMsgs, time.Now().Format("2006-01-02_15:04:05_-0700"))
	botStreamMsgs = append(botStreamMsgs, "CMD")
	botStreamMsgs = append(botStreamMsgs, "SHUTDOWN")

	msngr.AddToStream(fmt.Sprint(bot.KEY), botStreamMsgs)
	msngr.AddToStream("activeBots", msgs)
}

func editBot(bot Bot) {
	// add new trade info into stream (triggers other services)
	botStreamMsgs := []string{}
	botStreamMsgs = append(botStreamMsgs, "Timestamp")
	botStreamMsgs = append(botStreamMsgs, time.Now().Format("2006-01-02_15:04:05_-0700"))
	botStreamMsgs = append(botStreamMsgs, "CMD")
	botStreamMsgs = append(botStreamMsgs, "EDIT")
	botStreamMsgs = append(botStreamMsgs, "Leverage")
	botStreamMsgs = append(botStreamMsgs, bot.Leverage)
	botStreamMsgs = append(botStreamMsgs, "Risk")
	botStreamMsgs = append(botStreamMsgs, bot.AccountRiskPercPerTrade)
	botStreamMsgs = append(botStreamMsgs, "Size")
	botStreamMsgs = append(botStreamMsgs, bot.AccountSizePercToTrade)
	botStreamMsgs = append(botStreamMsgs, "Ticker")
	botStreamMsgs = append(botStreamMsgs, bot.Ticker)
	botStreamMsgs = append(botStreamMsgs, "Exchange")
	botStreamMsgs = append(botStreamMsgs, bot.ExchangeConnection)
	msngr.AddToStream(fmt.Sprint(bot.KEY), botStreamMsgs)
}
