package gdax

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

const (
	// entry types
	TransferEntry = "transfer"
	MatchEntry    = "match"
	FeeEntry      = "fee"
	RebateEntry   = "rebate"
)

type Account struct {
	Id        *uuid.UUID `json:"id,string"`
	Currency  string     `json:"currency"`
	Balance   float64    `json:"balance,string"`
	Available float64    `json:"available,string"`
	Holds     float64    `json:"holds,string"`
	ProfileId string     `json:"profile_id,omitempty"`
}

type AccountHold struct {
	Id        *uuid.UUID `json:"id,string"`
	AccountId *uuid.UUID `json:"account_id,string"`
	CreatedAt time.Time  `json:"created_at,string"`
	UpdatedAt time.Time  `json:"updated_at,string"`
	Amount    float64    `json:"amount,string"`
	Type      string     `json:"type"`
	Ref       string     `json:"ref"`
}

type AccountHistoryDetails struct {
	OrderId   *uuid.UUID `json:"order_id,string"`
	TradeId   string     `json:"trade_id"`
	ProductId string     `json:"product_id"`
}

type AccountHistory struct {
	Id        int64                 `json:"id"`
	CreatedAt time.Time             `json:"created_at,string"`
	Amount    float64               `json:"amount,string"`
	Balance   float64               `json:"balance,string"`
	Type      string                `json:"type"`
	Details   AccountHistoryDetails `json:"details"`
}

type AccountCollection struct {
	pageableCollection
}

type AccountHistoryCollection struct {
	pageableCollection
	id *uuid.UUID
}

type AccountHoldCollection struct {
	pageableCollection
	id *uuid.UUID
}

func (accessInfo *AccessInfo) GetAccounts() *AccountCollection {
	accountCollection := AccountCollection{
		pageableCollection: accessInfo.newPageableCollection(false),
	}
	return &accountCollection
}

func (accessInfo *AccessInfo) GetAccount(accountId *uuid.UUID) (*Account, error) {
	// GET /accounts/<account-id>
	var account Account
	_, err := accessInfo.request(http.MethodGet, fmt.Sprintf("/accounts/%s", accountId), "", &account)
	if err != nil {
		return nil, err
	}
	return &account, err
}

func (accessInfo *AccessInfo) GetAccountHistory(accountId *uuid.UUID) *AccountHistoryCollection {
	accountHistoryCollection := AccountHistoryCollection{
		pageableCollection: accessInfo.newPageableCollection(true),
		id:                 accountId,
	}
	return &accountHistoryCollection
}

func (accessInfo *AccessInfo) GetAccountHolds(accountId *uuid.UUID) *AccountHoldCollection {
	accountHoldCollection := AccountHoldCollection{
		pageableCollection: accessInfo.newPageableCollection(true),
		id:                 accountId,
	}
	return &accountHoldCollection
}

func (c *AccountCollection) HasNext() bool {
	// GET /accounts
	var accounts []Account
	return c.pageableCollection.hasNext(http.MethodGet, "/accounts", "", "", &accounts)
}

func (c *AccountHistoryCollection) HasNext() bool {
	// GET /accounts/<account-id>
	var accountHistory []AccountHistory
	return c.pageableCollection.hasNext(http.MethodGet, fmt.Sprintf("/accounts/%s/ledger", c.id), "", "", &accountHistory)
}

func (c *AccountHoldCollection) HasNext() bool {
	// GET /accounts/<account-id>/holds
	var accountHolds []AccountHold
	return c.pageableCollection.hasNext(http.MethodGet, fmt.Sprintf("/accounts/%s/holds", c.id), "", "", &accountHolds)
}

func (c *AccountCollection) Next() (*Account, error) {
	account, err := c.pageableCollection.next()
	if err != nil {
		return nil, err
	}
	return account.Addr().Interface().(*Account), nil
}

func (c *AccountHistoryCollection) Next() (*AccountHistory, error) {
	history, err := c.pageableCollection.next()
	if err != nil {
		return nil, err
	}
	return history.Addr().Interface().(*AccountHistory), nil
}

func (c *AccountHoldCollection) Next() (*AccountHold, error) {
	hold, err := c.pageableCollection.next()
	if err != nil {
		return nil, err
	}
	return hold.Addr().Interface().(*AccountHold), nil
}
