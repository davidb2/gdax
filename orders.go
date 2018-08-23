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

// Order Constants
const (
	Buy  = "buy" // Side
	Sell = "sell"

	Limit  = "limit" // Type
	Market = "market"

	Loss  = "loss" // Stop
	Entry = "entry"

	GoodTillTime      = "GTT" // Order Policy
	GoodTillCancelled = "GTC"
	ImmediateOrCancel = "IOC"
	FillOrKill        = "FOK"

	DecreaseAndCancel = "dc" // Self-Trade Prevention
	CancelOldest      = "co"
	CancelNewest      = "cn"

	Open    = "open" // Status
	Pending = "pending"
	Active  = "active"
	Done    = "done"
	All     = "all"
)

// An Order represents an order.
type Order struct {
	Side        string      `json:"side"`
	ProductID   string      `json:"product_id"`
	Type        string      `json:"type,omitempty"`
	ClientOid   *uuid.UUID  `json:"client_oid,string,omitempty"`
	Stp         string      `json:"stp,omitempty"`
	Stop        string      `json:"stop,omitempty"`
	StopPrice   float64     `json:"stop_price,string,omitempty"`
	TimeInForce string      `json:"time_in_force,omitempty"`
	CancelAfter *DayHourMin `json:"cancel_after,string,omitempty"`
	Funds       float64     `json:"funds,string,omitempty"`

	// additional fields
	ID            *uuid.UUID `json:"id,string,omitempty"`
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
	productID string
}

// An UUIDCollection is an iterator of UUIDs.
type UUIDCollection struct {
	pageableCollection
	productID string
	orderID   *uuid.UUID
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

	orderJSON, err := json.Marshal(*order)
	if err != nil {
		return nil, err
	}

	_, err = accessInfo.request(http.MethodPost, "/orders", string(orderJSON), &orderResponse)
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

	orderJSON, err := json.Marshal(*order)
	if err != nil {
		return nil, err
	}

	_, err = accessInfo.request(http.MethodPost, "/orders", string(orderJSON), &orderResponse)
	if err != nil {
		return nil, err
	}

	if err := mergo.Merge(&orderResponse, *order); err != nil {
		return nil, err
	}

	return &orderResponse, err
}

// CancelOrder cancels an order with the specified orderID.
// Note that this function is lazy.
func (accessInfo *AccessInfo) CancelOrder(orderID *uuid.UUID) *UUIDCollection {
	uuidCollection := UUIDCollection{
		pageableCollection: accessInfo.newPageableCollection(false),
		orderID:            orderID,
	}
	return &uuidCollection
}

// CancelAllOrders cancels all orders.
// Note that this function is lazy.
func (accessInfo *AccessInfo) CancelAllOrders() *UUIDCollection {
	return accessInfo.CancelAllOrdersForProduct("")
}

// CancelAllOrdersForProduct cancels all orders with the specified productID.
// Note that this function is lazy.
func (accessInfo *AccessInfo) CancelAllOrdersForProduct(productID string) *UUIDCollection {
	uuidCollection := UUIDCollection{
		pageableCollection: accessInfo.newPageableCollection(false),
		productID:          productID,
	}
	return &uuidCollection
}

// GetOrder gets the order with the specified orderID.
func (accessInfo *AccessInfo) GetOrder(orderID *uuid.UUID) (*Order, error) {
	// GET /orders/<order-id>
	var order Order

	_, err := accessInfo.request(http.MethodGet, fmt.Sprintf("/orders/%s", orderID), "", &order)
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// GetOrders gets all orders with the given statuses.
func (accessInfo *AccessInfo) GetOrders(statuses ...string) *OrderCollection {
	return accessInfo.GetOrdersForProduct("", statuses...)
}

// GetOrdersForProduct gets all orders with the specified productID and specified statuses.
func (accessInfo *AccessInfo) GetOrdersForProduct(productID string, statuses ...string) *OrderCollection {
	updatedStatuses := statuses[:]
	if len(statuses) == 0 {
		updatedStatuses = append(updatedStatuses, All)
	}
	orderCollection := OrderCollection{
		pageableCollection: accessInfo.newPageableCollection(true),
		statuses:           updatedStatuses,
		productID:          productID,
	}
	return &orderCollection
}

// HasNext determines if there is another Order in this iterator.
func (c *OrderCollection) HasNext() bool {
	// GET /orders
	var orders []Order
	statusParams := strings.Join(stringMap(c.statuses, func(s string) string { return "status=" + s }), "&")
	productParams := ""
	if c.productID != "" {
		productParams = fmt.Sprintf("product_id=%s", c.productID)
	}
	return c.pageableCollection.hasNext(http.MethodGet, "/orders", strings.Join(stringFilter([]string{statusParams, productParams}, notEmpty), "&"), "", &orders)
}

// HasNext determines if there is another UUID in this iterator.
func (c *UUIDCollection) HasNext() bool {
	// DELETE /orders
	var (
		productIDParam string
		orderIDParam   string
		cancelledIDs   []uuid.UUID
	)
	if c.productID != "" {
		productIDParam = "product_id=" + c.productID
	}
	if c.orderID != nil {
		orderIDParam = "order_id=" + c.orderID.String()
	}
	params := strings.Join(stringFilter([]string{productIDParam, orderIDParam}, notEmpty), "&")
	return c.pageableCollection.hasNext(http.MethodDelete, "/orders", params, "", &cancelledIDs)
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
