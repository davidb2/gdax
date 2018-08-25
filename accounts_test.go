package gdax_test

import (
	"fmt"
	"net/http"
	"strconv"
	"testing"

	"github.com/google/uuid"
	"github.com/ljeabmreosn/gdax"
	"github.com/stretchr/testify/assert"
	gock "gopkg.in/h2non/gock.v1"
)

const (
	accountsJSON = `
		[
		    {
		        "id": "71452118-efc7-4cc4-8780-a5e22d4baa53",
		        "currency": "BTC",
		        "balance": "0.0000000000000000",
		        "available": "0.0000000000000000",
		        "hold": "0.0000000000000000",
		        "profile_id": "75da88c5-05bf-4f54-bc85-5c775bd68254"
		    },
		    {
		        "id": "e316cb9a-0808-4fd7-8914-97829c1925de",
		        "currency": "USD",
		        "balance": "80.2301373066930000",
		        "available": "79.2266348066930000",
		        "hold": "1.0035025000000000",
		        "profile_id": "75da88c5-05bf-4f54-bc85-5c775bd68254"
		    }
		]
	`
	accountJSON = `
		{
		    "id": "6cf2b1ba-3705-40e6-a41e-69be033514f7",
		    "balance": "1.100",
		    "holds": "0.100",
		    "available": "1.00",
		    "currency": "USD"
		}
	`
	accountHistoryJSON1 = `
		[
		    {
		        "id": 100,
		        "created_at": "2014-11-07T08:19:27.028459Z",
		        "amount": "0.001",
		        "balance": "239.669",
		        "type": "fee",
		        "details": {
		            "order_id": "d50ec984-77a8-460a-b958-66f114b0de9b",
		            "trade_id": "74",
		            "product_id": "BTC-USD"
		        }
		    }
		]
	`
	accountHistoryJSON2 = `
		[
		    {
		        "id": 100,
		        "created_at": "2014-11-07T08:19:29.028459Z",
		        "amount": "0.001",
		        "balance": "170.322",
		        "type": "fee",
		        "details": {
		            "order_id": "62087add-1eea-47fc-b79f-8cde52b458d6",
		            "trade_id": "75",
		            "product_id": "BTC-USD"
		        }
		    }
		]
	`
	accountHoldJSON1 = `
		[
		    {
		        "id": "82dcd140-c3c7-4507-8de4-2c529cd1a28f",
		        "account_id": "e0b3f39a-183d-453e-b754-0c13e5bab0b3",
		        "created_at": "2014-11-06T10:34:47.123456Z",
		        "updated_at": "2014-11-06T10:40:47.123456Z",
		        "amount": "4.23",
		        "type": "order",
		        "ref": "0a205de4-dd35-4370-a285-fe8fc375a273"
		    },
		    {
		        "id": "1fa18826-8f96-4640-b73a-752d85c69326",
		        "account_id": "e0b3f39a-183d-453e-b754-0c13e5bab0b3",
		        "created_at": "2014-11-06T10:34:47.123456Z",
		        "updated_at": "2014-11-06T10:40:47.123456Z",
		        "amount": "5.25",
		        "type": "order",
		        "ref": "ba2a968c-17f9-4fcb-90d7-eb6f2ac49538"
		    }
		]
	`
	accountHoldJSON2 = `
		[
		    {
		        "id": "e6b60c60-42ed-4329-a311-694d6c897d9b",
		        "account_id": "e0b3f39a-183d-453e-b754-0c13e5bab0b3",
		        "created_at": "2014-11-06T10:34:47.123456Z",
		        "updated_at": "2014-11-06T10:40:47.123456Z",
		        "amount": "6.34",
		        "type": "order",
		        "ref": "ba2a968c-17f9-4fcb-90d7-eb6f2ac49538"
		    }
		]
	`
)

func TestGetAccountsError(t *testing.T) {
	defer gock.Off()
	assert := assert.New(t)

	accessInfo, err := gdax.RetrieveAccessInfoFromEnvironmentVariables()
	assert.NoError(err)

	gock.New(gdax.EndPoint).
		Get("/accounts").
		Reply(http.StatusNotFound).
		BodyString(`{"message": "Account id not found"}`)

	accounts := accessInfo.GetAccounts()
	assert.True(accounts.HasNext())

	account, err := accounts.Next()
	assert.Error(err)
	assert.Nil(account)
	assert.Equal(err.Error(), "Account id not found")
}

func TestGetAccounts(t *testing.T) {
	defer gock.Off()
	assert := assert.New(t)

	accessInfo, err := gdax.RetrieveAccessInfoFromEnvironmentVariables()
	assert.NoError(err)

	gock.New(gdax.EndPoint).
		Get("/accounts").
		Reply(http.StatusOK).
		BodyString(accountsJSON)

	var ids = [...]string{"71452118-efc7-4cc4-8780-a5e22d4baa53", "e316cb9a-0808-4fd7-8914-97829c1925de"}

	for idx, accounts := 0, accessInfo.GetAccounts(); accounts.HasNext(); idx++ {
		account, err := accounts.Next()
		assert.NoError(err)

		parsedID, err := uuid.Parse(ids[idx])
		assert.NoError(err)

		assert.Equal(*account.ID, parsedID)
	}
}

func TestGetAccountError(t *testing.T) {
	defer gock.Off()
	assert := assert.New(t)

	accessInfo, err := gdax.RetrieveAccessInfoFromEnvironmentVariables()
	assert.NoError(err)

	const id = "6cf2b1ba-3705-40e6-a41e-69be033514f7"
	gock.New(gdax.EndPoint).
		Get(fmt.Sprintf("/accounts/%s", id)).
		Reply(http.StatusNotFound).
		BodyString(`{"message": "Account id not found"}`)

	parsedID, err := uuid.Parse(id)
	assert.NoError(err)

	account, err := accessInfo.GetAccount(&parsedID)
	assert.Error(err)
	assert.Nil(account)
	assert.Equal(err.Error(), "Account id not found")
}

func TestGetAccount(t *testing.T) {
	defer gock.Off()
	assert := assert.New(t)

	accessInfo, err := gdax.RetrieveAccessInfoFromEnvironmentVariables()
	assert.NoError(err)

	const id = "6cf2b1ba-3705-40e6-a41e-69be033514f7"
	gock.New(gdax.EndPoint).
		Get(fmt.Sprintf("/accounts/%s", id)).
		Reply(http.StatusOK).
		BodyString(accountJSON)

	parsedID, err := uuid.Parse(id)
	assert.NoError(err)

	account, err := accessInfo.GetAccount(&parsedID)
	assert.NoError(err)

	assert.Equal(*account.ID, parsedID)
}

func TestGetAccountHistoryError(t *testing.T) {
	defer gock.Off()
	assert := assert.New(t)

	accessInfo, err := gdax.RetrieveAccessInfoFromEnvironmentVariables()
	assert.NoError(err)

	const accountID = "6cf2b1ba-3705-40e6-a41e-69be033514f7"
	gock.New(gdax.EndPoint).
		Get(fmt.Sprintf("/accounts/%s/ledger", accountID)).
		Reply(http.StatusNotFound).
		BodyString(`{"message": "Account id not found"}`)

	parsedAccountID, err := uuid.Parse(accountID)
	assert.NoError(err)

	accountHistories := accessInfo.GetAccountHistory(&parsedAccountID)
	assert.True(accountHistories.HasNext())

	accountHistory, err := accountHistories.Next()
	assert.Error(err)
	assert.Nil(accountHistory)
	assert.Equal(err.Error(), "Account id not found")
}

func TestGetAccountHistory(t *testing.T) {
	defer gock.Off()
	assert := assert.New(t)

	accessInfo, err := gdax.RetrieveAccessInfoFromEnvironmentVariables()
	assert.NoError(err)

	var cursors = [...]int{10, 20}
	const accountID = "6cf2b1ba-3705-40e6-a41e-69be033514f7"
	gock.New(gdax.EndPoint).
		Get(fmt.Sprintf("/accounts/%s/ledger", accountID)).
		Reply(http.StatusOK).
		BodyString(accountHistoryJSON1).
		SetHeader("CB-AFTER", strconv.Itoa(cursors[0]))
	gock.New(gdax.EndPoint).
		Get(fmt.Sprintf("/accounts/%s/ledger", accountID)).
		MatchParam("after", strconv.Itoa(cursors[0])).
		Reply(http.StatusOK).
		BodyString(accountHistoryJSON2).
		SetHeader("CB-AFTER", strconv.Itoa(cursors[1]))
	gock.New(gdax.EndPoint).
		Get(fmt.Sprintf("/accounts/%s/ledger", accountID)).
		MatchParam("after", strconv.Itoa(cursors[1])).
		Reply(http.StatusOK).
		BodyString("[]")

	var orderIDs = [...]string{"d50ec984-77a8-460a-b958-66f114b0de9b", "62087add-1eea-47fc-b79f-8cde52b458d6"}

	parsedAccountID, err := uuid.Parse(accountID)
	assert.NoError(err)

	for idx, accountHistories := 0, accessInfo.GetAccountHistory(&parsedAccountID); accountHistories.HasNext(); idx++ {
		accountHistory, err := accountHistories.Next()
		assert.NoError(err)
		t.Log(accountHistory)

		parsedID, err := uuid.Parse(orderIDs[idx])
		assert.NoError(err)

		assert.Equal(*accountHistory.Details.OrderID, parsedID)
	}
}

func TestGetAccountHoldsError(t *testing.T) {
	defer gock.Off()
	assert := assert.New(t)

	accessInfo, err := gdax.RetrieveAccessInfoFromEnvironmentVariables()
	assert.NoError(err)

	const accountID = "6cf2b1ba-3705-40e6-a41e-69be033514f7"
	gock.New(gdax.EndPoint).
		Get(fmt.Sprintf("/accounts/%s/holds", accountID)).
		Reply(http.StatusNotFound).
		BodyString(`{"message": "Account id not found"}`)

	parsedAccountID, err := uuid.Parse(accountID)
	assert.NoError(err)

	accountHolds := accessInfo.GetAccountHolds(&parsedAccountID)
	assert.True(accountHolds.HasNext())

	accountHold, err := accountHolds.Next()
	assert.Error(err)
	assert.Nil(accountHold)
	assert.Equal(err.Error(), "Account id not found")
}

func TestGetAccountHolds(t *testing.T) {
	defer gock.Off()
	assert := assert.New(t)

	accessInfo, err := gdax.RetrieveAccessInfoFromEnvironmentVariables()
	assert.NoError(err)

	var cursors = [...]int{10, 20}
	const accountID = "e0b3f39a-183d-453e-b754-0c13e5bab0b3"
	gock.New(gdax.EndPoint).
		Get(fmt.Sprintf("/accounts/%s/holds", accountID)).
		Reply(http.StatusOK).
		BodyString(accountHoldJSON1).
		SetHeader("CB-AFTER", strconv.Itoa(cursors[0]))
	gock.New(gdax.EndPoint).
		Get(fmt.Sprintf("/accounts/%s/holds", accountID)).
		MatchParam("after", strconv.Itoa(cursors[0])).
		Reply(http.StatusOK).
		BodyString(accountHoldJSON2).
		SetHeader("CB-AFTER", strconv.Itoa(cursors[1]))
	gock.New(gdax.EndPoint).
		Get(fmt.Sprintf("/accounts/%s/holds", accountID)).
		MatchParam("after", strconv.Itoa(cursors[1])).
		Reply(http.StatusOK).
		BodyString("[]")

	var ids = [...]string{"82dcd140-c3c7-4507-8de4-2c529cd1a28f", "1fa18826-8f96-4640-b73a-752d85c69326", "e6b60c60-42ed-4329-a311-694d6c897d9b"}
	var amounts = [...]float64{4.23, 5.25, 6.34}

	parsedAccountID, err := uuid.Parse(accountID)
	assert.NoError(err)

	for idx, accountHolds := 0, accessInfo.GetAccountHolds(&parsedAccountID); accountHolds.HasNext(); idx++ {
		accountHold, err := accountHolds.Next()
		assert.NoError(err)
		t.Log(accountHold)

		parsedID, err := uuid.Parse(ids[idx])
		assert.NoError(err)

		assert.Equal(*accountHold.ID, parsedID)
		assert.Equal(accountHold.Amount, amounts[idx])
	}
}
