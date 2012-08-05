package matcher

import (
	"testing"
)

const (
	stockId = 1
	trader1 = 1
	trader2 = 2
	trader3 = 3
)

type responseVals struct {
	price        int64
	amount       uint32
	tradeId      uint32
	counterParty uint32
}

func verifyResponse(t *testing.T, r *Response, vals responseVals) {
	price := vals.price
	amount := vals.amount
	tradeId := vals.tradeId
	counterParty := vals.counterParty
	if r.TradeId != tradeId {
		t.Errorf("Expecting %d trade-id, got %d instead", tradeId, r.TradeId)
	}
	if r.Amount != amount {
		t.Errorf("Expecting %d amount, got %d instead", amount, r.Amount)
	}
	if r.Price != price {
		t.Errorf("Expecting %d price, got %d instead", price, r.Price)
	}
	if r.CounterParty != counterParty {
		t.Errorf("Expecting %d counter party, got %d instead", counterParty, r.CounterParty)
	}
}

func responseChan() chan *Response {
	return make(chan *Response, 256)
}

func responseFunc(rc chan *Response) func(*Response) {
	return func(response *Response) {
		rc <- response
	}
}

func TestMidPoint(t *testing.T) {
	midpoint(t, 1, 1, 1)
	midpoint(t, 2, 1, 1)
	midpoint(t, 3, 1, 2)
	midpoint(t, 4, 1, 2)
	midpoint(t, 5, 1, 3)
	midpoint(t, 6, 1, 3)
	midpoint(t, 20, 10, 15)
	midpoint(t, 21, 10, 15)
	midpoint(t, 22, 10, 16)
	midpoint(t, 23, 10, 16)
	midpoint(t, 24, 10, 17)
	midpoint(t, 25, 10, 17)
	midpoint(t, 26, 10, 18)
	midpoint(t, 27, 10, 18)
	midpoint(t, 28, 10, 19)
	midpoint(t, 29, 10, 19)
	midpoint(t, 30, 10, 20)
}

func midpoint(t *testing.T, bPrice, sPrice, expected int64) {
	result := price(bPrice, sPrice)
	if result != expected {
		t.Errorf("price(%d,%d) does not equal %d, got %d instead.", bPrice, sPrice, expected, result)
	}
}

// Basic test matches lonely buy/sell trade pair which match exactly
func TestSimpleMatch(t *testing.T) {
	m := NewMatcher(stockId)
	addLowBuys(m, 5)
	addHighSells(m, 10)
	trader1Chan := responseChan()
	trader2Chan := responseChan()
	// Add Buy
	costData := CostData{Price: 7, Amount: 1}
	tradeData := TradeData{TraderId: trader1, TradeId: 1, StockId: stockId}
	m.AddBuy(NewBuy(costData, tradeData, responseFunc(trader1Chan)))
	// Add sell
	costData = CostData{Price: 7, Amount: 1}
	tradeData = TradeData{TraderId: trader2, TradeId: 2, StockId: stockId}
	m.AddSell(NewSell(costData, tradeData, responseFunc(trader2Chan)))
	// Verify
	verifyResponse(t, <-trader1Chan, responseVals{price: -7, amount: 1, tradeId: 1, counterParty: trader2})
	verifyResponse(t, <-trader2Chan, responseVals{price: 7, amount: 1, tradeId: 2, counterParty: trader1})
}

// Test matches one buy order to two separate sells
func TestDoubleSellMatch(t *testing.T) {
	m := NewMatcher(stockId)
	addLowBuys(m, 5)
	addHighSells(m, 10)
	trader1Chan := responseChan()
	trader2Chan := responseChan()
	trader3Chan := responseChan()
	// Add Buy
	costData := CostData{Price: 7, Amount: 2}
	tradeData := TradeData{TraderId: trader1, TradeId: 1, StockId: stockId}
	m.AddBuy(NewBuy(costData, tradeData, responseFunc(trader1Chan)))
	// Add Sell
	costData = CostData{Price: 7, Amount: 1}
	tradeData = TradeData{TraderId: trader2, TradeId: 2, StockId: stockId}
	m.AddSell(NewSell(costData, tradeData, responseFunc(trader2Chan)))
	// Verify
	verifyResponse(t, <-trader1Chan, responseVals{price: -7, amount: 1, tradeId: 1, counterParty: trader2})
	verifyResponse(t, <-trader2Chan, responseVals{price: 7, amount: 1, tradeId: 2, counterParty: trader1})
	// Add Sell
	costData = CostData{Price: 7, Amount: 1}
	tradeData = TradeData{TraderId: trader3, TradeId: 3, StockId: stockId}
	m.AddSell(NewSell(costData, tradeData, responseFunc(trader3Chan)))
	// Verify
	verifyResponse(t, <-trader1Chan, responseVals{price: -7, amount: 1, tradeId: 1, counterParty: trader3})
	verifyResponse(t, <-trader3Chan, responseVals{price: 7, amount: 1, tradeId: 3, counterParty: trader1})
}

// Test matches two buy orders to one sell
func TestDoubleBuyMatch(t *testing.T) {
	m := NewMatcher(stockId)
	addLowBuys(m, 5)
	addHighSells(m, 10)
	trader1Chan := responseChan()
	trader2Chan := responseChan()
	trader3Chan := responseChan()
	// Add Sell
	costData := CostData{Price: 7, Amount: 2}
	tradeData := TradeData{TraderId: trader1, TradeId: 1, StockId: stockId}
	m.AddSell(NewSell(costData, tradeData, responseFunc(trader1Chan)))
	// Add Buy
	costData = CostData{Price: 7, Amount: 1}
	tradeData = TradeData{TraderId: trader2, TradeId: 2, StockId: stockId}
	m.AddBuy(NewBuy(costData, tradeData, responseFunc(trader2Chan)))
	verifyResponse(t, <-trader1Chan, responseVals{price: 7, amount: 1, tradeId: 1, counterParty: trader2})
	verifyResponse(t, <-trader2Chan, responseVals{price: -7, amount: 1, tradeId: 2, counterParty: trader1})
	// Add Buy
	costData = CostData{Price: 7, Amount: 1}
	tradeData = TradeData{TraderId: trader3, TradeId: 3, StockId: stockId}
	m.AddBuy(NewBuy(costData, tradeData, responseFunc(trader3Chan)))
	verifyResponse(t, <-trader1Chan, responseVals{price: 7, amount: 1, tradeId: 1, counterParty: trader3})
	verifyResponse(t, <-trader3Chan, responseVals{price: -7, amount: 1, tradeId: 3, counterParty: trader1})
}

// Test matches lonely buy/sell pair, with same quantity, uses the mid-price point for trade price
func TestMidPrice(t *testing.T) {
	m := NewMatcher(stockId)
	addLowBuys(m, 5)
	addHighSells(m, 10)
	trader1Chan := responseChan()
	trader2Chan := responseChan()
	// Add Buy
	costData := CostData{Price: 9, Amount: 1}
	tradeData := TradeData{TraderId: trader1, TradeId: 1, StockId: stockId}
	m.AddBuy(NewBuy(costData, tradeData, responseFunc(trader1Chan)))
	// Add Sell
	costData = CostData{Price: 6, Amount: 1}
	tradeData = TradeData{TraderId: trader2, TradeId: 1, StockId: stockId}
	m.AddSell(NewSell(costData, tradeData, responseFunc(trader2Chan)))
	verifyResponse(t, <-trader1Chan, responseVals{price: -7, amount: 1, tradeId: 1, counterParty: trader2})
	verifyResponse(t, <-trader2Chan, responseVals{price: 7, amount: 1, tradeId: 1, counterParty: trader1})
}

// Test matches lonely buy/sell pair, sell > quantity, and uses the mid-price point for trade price
func TestMidPriceBigSell(t *testing.T) {
	m := NewMatcher(stockId)
	addLowBuys(m, 5)
	addHighSells(m, 10)
	trader1Chan := responseChan()
	trader2Chan := responseChan()
	// Add Buy
	costData := CostData{Price: 9, Amount: 1}
	tradeData := TradeData{TraderId: trader1, TradeId: 1, StockId: stockId}
	m.AddBuy(NewBuy(costData, tradeData, responseFunc(trader1Chan)))
	// Add Sell
	costData = CostData{Price: 6, Amount: 10}
	tradeData = TradeData{TraderId: trader2, TradeId: 1, StockId: stockId}
	m.AddSell(NewSell(costData, tradeData, responseFunc(trader2Chan)))
	// Verify
	verifyResponse(t, <-trader1Chan, responseVals{price: -7, amount: 1, tradeId: 1, counterParty: trader2})
	verifyResponse(t, <-trader2Chan, responseVals{price: 7, amount: 1, tradeId: 1, counterParty: trader1})
}

// Test matches lonely buy/sell pair, buy > quantity, and uses the mid-price point for trade price
func TestMidPriceBigBuy(t *testing.T) {
	m := NewMatcher(stockId)
	addLowBuys(m, 5)
	addHighSells(m, 10)
	trader1Chan := responseChan()
	trader2Chan := responseChan()
	// Add Buy
	costData := CostData{Price: 9, Amount: 10}
	tradeData := TradeData{TraderId: trader1, TradeId: 1, StockId: stockId}
	m.AddBuy(NewBuy(costData, tradeData, responseFunc(trader1Chan)))
	// Add Sell
	costData = CostData{Price: 6, Amount: 1}
	tradeData = TradeData{TraderId: trader2, TradeId: 1, StockId: stockId}
	m.AddSell(NewSell(costData, tradeData, responseFunc(trader2Chan)))
	verifyResponse(t, <-trader1Chan, responseVals{price: -7, amount: 1, tradeId: 1, counterParty: trader2})
	verifyResponse(t, <-trader2Chan, responseVals{price: 7, amount: 1, tradeId: 1, counterParty: trader1})
}

func addLowBuys(m *M, highestPrice int64) {
	buys := mkBuys(10, 1, highestPrice)
	for _, buy := range buys {
		m.AddBuy(buy)
	}
}

func addHighSells(m *M, lowestPrice int64) {
	sells := mkSells(10, lowestPrice, lowestPrice+10000)
	for _, sell := range sells {
		m.AddSell(sell)
	}
}