package main

//test strat to pass to backtest func, runs once per historical candlestick
func strat1(open, high, low, close []float64, relCandleIndex int, strategy *StrategySimulator, storage *interface{}) {
	if strategy.PosLongSize == 0 {
		//if two green candles in a row, buy
		if (close[0] > open[0]) && (close[1] > open[1]) {
			strategy.Buy(close[0], 69.69, true)
		}
	} else {
		//if two red candles in a row, sell
		if (open[0] > close[0]) && (open[1] > close[1]) {
			strategy.CloseLong(close[0], 69.69)
		}
	}
}

func runBacktest(
	userStrat func(float64, float64, float64, float64, int, *StrategySimulator, *interface{}),
) ([]CandlestickChartData, []ProfitCurveData, []SimulatedTradeData) {
	//TODO: get all candlestick data for selected backtest period
	data := []Candlestick{}
	strategySim := StrategySimulator{}
	strategySim.Init(1000)
	var storage interface{}

	for i, candle := range data {
		userStrat(candle.Open, candle.High, candle.Low, candle.Close, i, &strategySim, &storage)
		//TODO: build display data as well using strategySim
	}

	return nil, nil, nil
}
