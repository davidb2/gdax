package gdax

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	Maker = "M"
	Taker = "T"
)

// A Fill represents a fill order.
type Fill struct {
	TradeId   int64      `json:"trade_id"`
	ProductId string     `json:"product_id"`
	Price     float64    `json:"price,string"`
	Size      float64    `json:"size,string"`
	OrderId   *uuid.UUID `json:"order_id,string"`
	CreatedAt *time.Time `json:"created_at,string"`
	Liquidity string     `json:"liquidity"`
	Fee       float64    `json:"fee,string"`
	Settled   bool       `json:"settled"`
	Side      string     `json:"side"`
}

// A FillCollection is an iterator of Fills.
type FillCollection struct {
	pageableCollection
	orderId   *uuid.UUID
	productId string
}

// GetFills gets all fills with the specified orderIds.
func (accessInfo *AccessInfo) GetFills(orderId ...*uuid.UUID) *FillCollection {
	return accessInfo.GetFillsForProduct("", orderId...)
}

// GetFillsForProduct gets all fills for a specified productId and specified orderIds.
func (accessInfo *AccessInfo) GetFillsForProduct(productId string, orderId ...*uuid.UUID) *FillCollection {
	var realOrderId *uuid.UUID
	if len(orderId) > 0 {
		realOrderId = orderId[0]
	}
	fillCollection := FillCollection{
		pageableCollection: accessInfo.newPageableCollection(true),
		orderId:            realOrderId,
		productId:          productId,
	}
	return &fillCollection
}

// HasNext determines if there is another Fill in this iterator.
func (c *FillCollection) HasNext() bool {
	// GET /fills
	var (
		orderParam   string
		productParam string
		fills        []Fill
	)

	if c.orderId != nil {
		orderParam = fmt.Sprintf("order_id=%s", c.orderId)
	}
	if c.productId != "" {
		productParam = fmt.Sprintf("product_id=%s", c.productId)
	}

	params := strings.Join(stringFilter([]string{orderParam, productParam}, notEmpty), "&")
	return c.pageableCollection.hasNext(http.MethodGet, "/fills", params, "", &fills)
}

// Next gets the next Fill from the iterator.
func (c *FillCollection) Next() (*Fill, error) {
	fill, err := c.pageableCollection.next()
	if err != nil {
		return nil, err
	}
	return fill.Addr().Interface().(*Fill), nil
}
