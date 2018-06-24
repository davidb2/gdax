package gdax

import (
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"strconv"
	"time"
)

const (
	SubscriptionsType = "subscriptions"
	HeartbeatType     = "heartbeat"
	TickerType        = "ticker"
	Level2Type        = "level2"
	L2UpdateType      = "l2update"
	SnapshotType      = "snapshot"
	UserType          = "user"
	MatchesType       = "matches"
	MatchType         = "match"
	FullType          = "full"
	ErrorType         = "error"
	SubscribeType     = "subscribe"
	addr              = "wss://ws-feed.gdax.com"
)

// interfaces

type Message interface {
	MessageType() string
}

// structs

type message struct {
	Type      string `json:"type"`
	ProductId string `json:"product_id,,omitempty"`
}

type Bid struct {
	Price float64 `json:"price"`
	Size  float64 `json:"size"`
}

type Ask struct {
	Price float64 `json:"price"`
	Size  float64 `json:"size"`
}

type Change struct {
	Side  string  `json:"side"`
	Price float64 `json:"price"`
	Size  float64 `json:"size"`
}

type Error struct {
	message
	Message string `json:"message"`
}

type Subscription struct {
	Type       string   `json:"type"`
	Channels   []string `json:"channels"`
	ProductIds []string `json:"product_ids"`
}

type Heartbeat struct {
	message
	Sequence    int64      `json:"sequence"`
	LastTradeId int64      `json:"last_trade_id"`
	Time        *time.Time `json:"time,string"`
}

type Ticker struct {
	message
	TradeId   int64      `json:"trade_id"`
	Sequence  int64      `json:"sequence"`
	Time      *time.Time `json:"time,string"`
	ProductId string     `json:"product_id"`
	Price     float64    `json:"price,string"`
	Side      string     `json:"side"`
	LastSize  float64    `json:"last_size,string"`
	BestBid   float64    `json:"best_bid,string"`
	BestAsk   float64    `json:"best_ask,string"`
}

type Snapshot struct {
	message
	ProductId string `json:"product_id"`
	Bids      []Bid  `json:"bids"`
	Asks      []Ask  `json:"asks"`
}

type L2Update struct {
	message
	Changes []Change `json:"changes"`
}

type Match struct {
	message
	Time         *time.Time `json:"time,string"`
	Sequence     int64      `json:"sequence"`
	TradeId      int64      `json:"trade_id"`
	MakerOrderId *uuid.UUID `json:"maker_order_id,string"`
	TakerOrderId *uuid.UUID `json:"taker_order_id,string"`
	Size         float64    `json:"size,string"`
	Price        float64    `json:"price,string"`
	Side         string     `json:"side"`
}

// struct functions

func (err Error) Error() string {
	return err.Message
}

func (m message) MessageType() string {
	return m.Type
}

func (m *L2Update) UnmarshalJSON(b []byte) error {
	var fields map[string]interface{}
	if err := json.Unmarshal(b, &fields); err != nil {
		return err
	}
	for key, val := range fields {
		switch key {
		case "type":
			m.Type = val.(string)
		case "product_id":
			m.ProductId = val.(string)
		case "changes":
			for _, e := range val.([]interface{}) {
				side := e.([]interface{})[0].(string)
				price, err := strconv.ParseFloat(e.([]interface{})[1].(string), 64)
				if err != nil {
					return err
				}
				size, err := strconv.ParseFloat(e.([]interface{})[2].(string), 64)
				if err != nil {
					return err
				}
				m.Changes = append(m.Changes, Change{Side: side, Price: price, Size: size})
			}
		}
	}
	return nil
}

func (m *Snapshot) UnmarshalJSON(b []byte) error {
	var fields map[string]interface{}
	if err := json.Unmarshal(b, &fields); err != nil {
		return err
	}
	for key, val := range fields {
		switch key {
		case "type":
			m.Type = val.(string)
		case "product_id":
			m.ProductId = val.(string)
		case "bids":
			for _, e := range val.([]interface{}) {
				price, err := strconv.ParseFloat(e.([]interface{})[0].(string), 64)
				if err != nil {
					return err
				}
				size, err := strconv.ParseFloat(e.([]interface{})[1].(string), 64)
				if err != nil {
					return err
				}
				m.Bids = append(m.Bids, Bid{Price: price, Size: size})
			}
		case "asks":
			for _, e := range val.([]interface{}) {
				price, err := strconv.ParseFloat(e.([]interface{})[0].(string), 64)
				if err != nil {
					return err
				}
				size, err := strconv.ParseFloat(e.([]interface{})[1].(string), 64)
				if err != nil {
					return err
				}
				m.Asks = append(m.Asks, Ask{Price: price, Size: size})
			}
		}
	}
	return nil
}

// functions

func Feed(s *Subscription, messageHandler func(Message)) error {
	body, err := json.Marshal(*s)
	if err != nil {
		return err
	}
	messageType := make(chan string)
	jsonString := make(chan []byte)
	errorChan := make(chan error)
	if err = createWebsocketConnection(addr, body, messageType, jsonString, errorChan); err != nil {
		return err
	}
	for {
		if err := <-errorChan; err != nil {
			return err
		}
		messageTypeInstance := <-messageType
		jsonInstance := <-jsonString
		switch messageTypeInstance {
		case HeartbeatType:
			var heartbeat Heartbeat
			if err := json.Unmarshal(jsonInstance, &heartbeat); err != nil {
				return err
			}
			messageHandler(heartbeat)
		case TickerType:
			var ticker Ticker
			if err := json.Unmarshal(jsonInstance, &ticker); err != nil {
				return err
			}
			messageHandler(ticker)
		case L2UpdateType:
			var l2update L2Update
			if err := json.Unmarshal(jsonInstance, &l2update); err != nil {
				return err
			}
			messageHandler(l2update)
		case SnapshotType:
			var snapshot Snapshot
			if err := json.Unmarshal(jsonInstance, &snapshot); err != nil {
				return err
			}
			messageHandler(snapshot)
		case MatchType:
			var match Match
			if err := json.Unmarshal(jsonInstance, &match); err != nil {
				return err
			}
			messageHandler(match)
		case ErrorType:
			var e Error
			if err := json.Unmarshal(jsonInstance, &e); err != nil {
				return err
			}
			messageHandler(e)
			return errors.New(e.Message)
		}
	}
	return nil
}
