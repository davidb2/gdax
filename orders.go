package gdax

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/imdario/mergo"
)

const (
	// side
	Buy  = "buy"
	Sell = "sell"

	// type
	Limit  = "limit"
	Market = "market"

	// stop
	Loss  = "loss"
	Entry = "entry"

	// order policy
	GoodTillTime      = "GTT"
	GoodTillCancelled = "GTC"
	ImmediateOrCancel = "IOC"
	FillOrKill        = "FOK"

	// self-trade prevention
	DecreaseAndCancel = "dc"
	CancelOldest      = "co"
	CancelNewest      = "cn"

	// status
	Open    = "open"
	Pending = "pending"
	Active  = "active"
	Done    = "done"
	All     = "all"
)

// An Order represents an order.
type Order struct {
	Side        string      `json:"side"`
	ProductId   string      `json:"product_id"`
	Type        string      `json:"type,omitempty"`
	ClientOid   *uuid.UUID  `json:"client_oid,string,omitempty"`
	Stp         string      `json:"stp,omitempty"`
	Stop        string      `json:"stop,omitempty"`
	StopPrice   float64     `json:"stop_price,string,omitempty"`
	TimeInForce string      `json:"time_in_force,omitempty"`
	CancelAfter *DayHourMin `json:"cancel_after,string,omitempty"`
	Funds       float64     `json:"funds,string,omitempty"`

	// additional fields
	Id            *uuid.UUID `json:"id,string,omitempty"`
	Price         float64    `json:"price,string,omitempty"`
	Size          float64    `json:"size,string,omitempty"`
	PostOnly      bool       `json:"post_only,omitempty"`
	CreatedAt     *time.Time `json:"created_at,string,omitempty"`
	FillFees      float64    `json:"fill_fees,string,omitempty"`
	FilledSize    float64    `json:"filled_size,string,omitempty"`
	ExecutedValue float64    `json:"executed_value,string,omitempty"`
	Status        string     `json:"status,omitempty"`
	Settled       bool       `json:"settled,omitempty"`
}

// An OrderCollection is an iterator of Orders.
type OrderCollection struct {
	pageableCollection
	statuses  []string
	productId string
}


// An UUIDCollection is an iterator of UUIDs.
type UUIDCollection struct {
	pageableCollection
	productId string
	orderId   *uuid.UUID
}

// PlaceMarketOrder places a market order.
func (accessInfo *AccessInfo) PlaceMarketOrder(order *Order) (*Order, error) {
	// POST /orders
	var orderResponse Order

	// fill in some more info about the order
	order.Type = Market
	if order.ClientOid == nil {
		clientOid := uuid.New()
		order.ClientOid = &clientOid
	}

	orderJson, err := json.Marshal(*order)
	if err != nil {
		return nil, err
	}

	_, err = accessInfo.request(http.MethodPost, "/orders", string(orderJson), &orderResponse)
	if err != nil {
		return nil, err
	}

	if err := mergo.Merge(&orderResponse, *order); err != nil {
		return nil, err
	}

	return &orderResponse, err
}

// PlaceLimitOrder places a limit order.
func (accessInfo *AccessInfo) PlaceLimitOrder(order *Order) (*Order, error) {
	// POST /orders
	var orderResponse Order

	// fill in some more info about the order
	order.Type = Limit
	if order.ClientOid == nil {
		clientOid := uuid.New()
		order.ClientOid = &clientOid
	}

	orderJson, err := json.Marshal(*order)
	if err != nil {
		return nil, err
	}

	_, err = accessInfo.request(http.MethodPost, "/orders", string(orderJson), &orderResponse)
	if err != nil {
		return nil, err
	}

	if err := mergo.Merge(&orderResponse, *order); err != nil {
		return nil, err
	}

	return &orderResponse, err
}

// CancelOrder cancels an order with the specified orderId.
// Note that this function is lazy.
func (accessInfo *AccessInfo) CancelOrder(orderId *uuid.UUID) *UUIDCollection {
	uuidCollection := UUIDCollection{
		pageableCollection: accessInfo.newPageableCollection(false),
		orderId:            orderId,
	}
	return &uuidCollection
}

// CancelAllOrders cancels all orders.
// Note that this function is lazy.
func (accessInfo *AccessInfo) CancelAllOrders() *UUIDCollection {
	return accessInfo.CancelAllOrdersForProduct("")
}

// CancelAllOrdersForProduct cancels all orders with the specified productId.
// Note that this function is lazy.
func (accessInfo *AccessInfo) CancelAllOrdersForProduct(productId string) *UUIDCollection {
	uuidCollection := UUIDCollection{
		pageableCollection: accessInfo.newPageableCollection(false),
		productId:          productId,
	}
	return &uuidCollection
}

// GetOrder gets the order with the specified orderId.
func (accessInfo *AccessInfo) GetOrder(orderId *uuid.UUID) (*Order, error) {
	// GET /orders/<order-id>
	var order Order

	_, err := accessInfo.request(http.MethodGet, fmt.Sprintf("/orders/%s", orderId), "", &order)
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// GetOrders gets all orders with the given statuses.
func (accessInfo *AccessInfo) GetOrders(statuses ...string) *OrderCollection {
	return accessInfo.GetOrdersForProduct("", statuses...)
}

// GetOrdersForProduct gets all orders with the specified productId and specified statuses.
func (accessInfo *AccessInfo) GetOrdersForProduct(productId string, statuses ...string) *OrderCollection {
	updatedStatuses := statuses[:]
	if len(statuses) == 0 {
		updatedStatuses = append(updatedStatuses, All)
	}
	orderCollection := OrderCollection{
		pageableCollection: accessInfo.newPageableCollection(true),
		statuses:           updatedStatuses,
		productId:          productId,
	}
	return &orderCollection
}

// HasNext determines if there is another Order in this iterator.
func (c *OrderCollection) HasNext() bool {
	// GET /orders
	var orders []Order
	statusParams := strings.Join(stringMap(c.statuses, func(s string) string { return "status=" + s }), "&")
	productParams := ""
	if c.productId != "" {
		productParams = fmt.Sprintf("product_id=%s", c.productId)
	}
	return c.pageableCollection.hasNext(http.MethodGet, "/orders", strings.Join(stringFilter([]string{statusParams, productParams}, notEmpty), "&"), "", &orders)
}

// HasNext determines if there is another UUID in this iterator.
func (c *UUIDCollection) HasNext() bool {
	// DELETE /orders
	var (
		productIdParam string
		orderIdParam   string
		cancelledIds   []uuid.UUID
	)
	if c.productId != "" {
		productIdParam = "product_id=" + c.productId
	}
	if c.orderId != nil {
		orderIdParam = "order_id=" + c.orderId.String()
	}
	params := strings.Join(stringFilter([]string{productIdParam, orderIdParam}, notEmpty), "&")
	return c.pageableCollection.hasNext(http.MethodDelete, "/orders", params, "", &cancelledIds)
}

// Next gets the next Order from the iterator.
func (c *OrderCollection) Next() (*Order, error) {
	order, err := c.pageableCollection.next()
	if err != nil {
		return nil, err
	}
	return order.Addr().Interface().(*Order), nil
}

// Next gets the next UUID from the iterator.
func (c *UUIDCollection) Next() (*uuid.UUID, error) {
	id, err := c.pageableCollection.next()
	if err != nil {
		return nil, err
	}
	return id.Addr().Interface().(*uuid.UUID), nil
}
