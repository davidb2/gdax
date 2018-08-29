package gdax_test

import (
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/ljeabmreosn/gdax"
	"github.com/stretchr/testify/assert"
	gock "gopkg.in/h2non/gock.v1"
)

const (
	fillJSON1 = `
		[
		    {
		        "trade_id": 74,
		        "product_id": "BTC-USD",
		        "price": "10.00",
		        "size": "0.01",
		        "order_id": "d50ec984-77a8-460a-b958-66f114b0de9b",
		        "created_at": "2014-11-07T22:19:28.578544Z",
		        "liquidity": "T",
		        "fee": "0.00025",
		        "settled": true,
		        "side": "buy"
		    }
		]
	`
	fillJSON2 = `
		[
		    {
		        "trade_id": 74,
		        "product_id": "BTC-USD",
		        "price": "9.00",
		        "size": "0.02",
		        "order_id": "03a7a57f-c5d5-4e29-b7a1-118b3a6cc88d",
		        "created_at": "2014-11-07T22:19:28.578544Z",
		        "liquidity": "T",
		        "fee": "0.00025",
		        "settled": true,
		        "side": "buy"
		    }
		]
	`
)

func TestGetFillsError(t *testing.T) {
	defer gock.Off()
	assert := assert.New(t)

	accessInfo, err := gdax.RetrieveAccessInfoFromEnvironmentVariables()
	assert.NoError(err)

	const orderID = "6cf2b1ba-3705-40e6-a41e-69be033514f7"
	gock.New(gdax.EndPoint).
		Get("/fills").
		MatchParam("order_id", orderID).
		Reply(http.StatusNotFound).
		BodyString(`{"message": "Order id not found"}`)

	parsedOrderID, err := uuid.Parse(orderID)
	assert.NoError(err)

	fills := accessInfo.GetFills(&parsedOrderID)
	assert.True(fills.HasNext())

	fill, err := fills.Next()
	assert.Error(err)
	assert.Nil(fill)
	assert.Equal(err.Error(), "Order id not found")
}

func TestGetFills(t *testing.T) {
	defer gock.Off()
	assert := assert.New(t)

	accessInfo, err := gdax.RetrieveAccessInfoFromEnvironmentVariables()
	assert.NoError(err)

	var cursors = [...]int{10, 20}
	var orderIDs = [...]string{"d50ec984-77a8-460a-b958-66f114b0de9b", "03a7a57f-c5d5-4e29-b7a1-118b3a6cc88d"}
	gock.New(gdax.EndPoint).
		Get("/fills").
		MatchParam("order_id", strings.Join(orderIDs[:], ",")).
		Reply(http.StatusOK).
		BodyString(fillJSON1).
		SetHeader("CB-AFTER", strconv.Itoa(cursors[0]))
	gock.New(gdax.EndPoint).
		Get("/fills").
		MatchParam("order_id", strings.Join(orderIDs[:], ",")).
		MatchParam("after", strconv.Itoa(cursors[0])).
		Reply(http.StatusOK).
		BodyString(fillJSON2).
		SetHeader("CB-AFTER", strconv.Itoa(cursors[1]))
	gock.New(gdax.EndPoint).
		Get("/fills").
		MatchParam("order_id", strings.Join(orderIDs[:], ",")).
		MatchParam("after", strconv.Itoa(cursors[1])).
		Reply(http.StatusOK).
		BodyString("[]")

	var parsedOrderIDs []*uuid.UUID
	for _, orderID := range orderIDs {
		parsedOrderID, err := uuid.Parse(orderID)
		assert.NoError(err)
		parsedOrderIDs = append(parsedOrderIDs, &parsedOrderID)
	}

	for idx, fills := 0, accessInfo.GetFills(parsedOrderIDs[:]...); fills.HasNext(); idx++ {
		fill, err := fills.Next()
		assert.NoError(err)

		assert.Equal(*fill.OrderID, *parsedOrderIDs[idx])
	}
}

func TestGetFillsForProduct(t *testing.T) {
	defer gock.Off()
	assert := assert.New(t)

	accessInfo, err := gdax.RetrieveAccessInfoFromEnvironmentVariables()
	assert.NoError(err)

	var cursors = [...]int{10, 20}
	var orderIDs = [...]string{"d50ec984-77a8-460a-b958-66f114b0de9b", "03a7a57f-c5d5-4e29-b7a1-118b3a6cc88d"}
	gock.New(gdax.EndPoint).
		Get("/fills").
		MatchParam("order_id", strings.Join(orderIDs[:], ",")).
		Reply(http.StatusOK).
		BodyString(fillJSON1).
		SetHeader("CB-AFTER", strconv.Itoa(cursors[0]))
	gock.New(gdax.EndPoint).
		Get("/fills").
		MatchParam("order_id", strings.Join(orderIDs[:], ",")).
		MatchParam("after", strconv.Itoa(cursors[0])).
		Reply(http.StatusOK).
		BodyString(fillJSON2).
		SetHeader("CB-AFTER", strconv.Itoa(cursors[1]))
	gock.New(gdax.EndPoint).
		Get("/fills").
		MatchParam("order_id", strings.Join(orderIDs[:], ",")).
		MatchParam("after", strconv.Itoa(cursors[1])).
		Reply(http.StatusOK).
		BodyString("[]")

	var parsedOrderIDs []*uuid.UUID
	for _, orderID := range orderIDs {
		parsedOrderID, err := uuid.Parse(orderID)
		assert.NoError(err)
		parsedOrderIDs = append(parsedOrderIDs, &parsedOrderID)
	}

	for idx, fills := 0, accessInfo.GetFillsForProduct("BTC-USD", parsedOrderIDs[:]...); fills.HasNext(); idx++ {
		fill, err := fills.Next()
		assert.NoError(err)

		assert.Equal(*fill.OrderID, *parsedOrderIDs[idx])
		assert.Equal(fill.ProductID, "BTC-USD")
	}
}
