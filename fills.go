package gdax

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Fills
const (
	Maker = "M"
	Taker = "T"
)

// A Fill represents a fill order.
type Fill struct {
	TradeID   int64      `json:"trade_id"`
	ProductID string     `json:"product_id"`
	Price     float64    `json:"price,string"`
	Size      float64    `json:"size,string"`
	OrderID   *uuid.UUID `json:"order_id,string"`
	CreatedAt *time.Time `json:"created_at,string"`
	Liquidity string     `json:"liquidity"`
	Fee       float64    `json:"fee,string"`
	Settled   bool       `json:"settled"`
	Side      string     `json:"side"`
}

// A FillCollection is an iterator of Fills.
type FillCollection struct {
	pageableCollection
	orderIDs  []*uuid.UUID
	productID string
}

// GetFills gets all fills with the specified orderIDs.
func (accessInfo *AccessInfo) GetFills(orderIDs ...*uuid.UUID) *FillCollection {
	return accessInfo.GetFillsForProduct("", orderIDs...)
}

// GetFillsForProduct gets all fills for a specified productID and specified orderIDs.
func (accessInfo *AccessInfo) GetFillsForProduct(productID string, orderIDs ...*uuid.UUID) *FillCollection {
	fillCollection := FillCollection{
		pageableCollection: accessInfo.newPageableCollection(true),
		orderIDs:           orderIDs,
		productID:          productID,
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

	if c.orderIDs != nil {
		unparsedOrderIDs := make([]string, len(c.orderIDs))
		for idx, orderID := range c.orderIDs {
			unparsedOrderIDs[idx] = orderID.String()
		}
		orderParam = fmt.Sprintf("order_id=%s", strings.Join(unparsedOrderIDs, ","))
	}
	if c.productID != "" {
		productParam = fmt.Sprintf("product_id=%s", c.productID)
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
