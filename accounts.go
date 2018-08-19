package gdax

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

const (
	TransferEntry = "transfer"
	MatchEntry    = "match"
	FeeEntry      = "fee"
	RebateEntry   = "rebate"
)

// An Account represents the user's account.
type Account struct {
	Id        *uuid.UUID `json:"id,string"`
	Currency  string     `json:"currency"`
	Balance   float64    `json:"balance,string"`
	Available float64    `json:"available,string"`
	Holds     float64    `json:"holds,string"`
	ProfileId string     `json:"profile_id,omitempty"`
}

// An AccountHold represents any holds that the user has.
type AccountHold struct {
	Id        *uuid.UUID `json:"id,string"`
	AccountId *uuid.UUID `json:"account_id,string"`
	CreatedAt time.Time  `json:"created_at,string"`
	UpdatedAt time.Time  `json:"updated_at,string"`
	Amount    float64    `json:"amount,string"`
	Type      string     `json:"type"`
	Ref       string     `json:"ref"`
}

// An AccountHistoryDetails represents information about past trades that the user has made.
type AccountHistoryDetails struct {
	OrderId   *uuid.UUID `json:"order_id,string"`
	TradeId   string     `json:"trade_id"`
	ProductId string     `json:"product_id"`
}

// An AccountHistory represents information about the past state(s) of the user's account.
type AccountHistory struct {
	Id        int64                 `json:"id"`
	CreatedAt time.Time             `json:"created_at,string"`
	Amount    float64               `json:"amount,string"`
	Balance   float64               `json:"balance,string"`
	Type      string                `json:"type"`
	Details   AccountHistoryDetails `json:"details"`
}

// An AccountCollection is an iterator of Accounts.
type AccountCollection struct {
	pageableCollection
}

// An AccountHistoryCollection is an iterator of AccountHistorys.
type AccountHistoryCollection struct {
	pageableCollection
	id *uuid.UUID
}

// An AccountHoldCollection is an iterator of AccountHoldCollections.
type AccountHoldCollection struct {
	pageableCollection
	id *uuid.UUID
}

// GetAccounts gets all associated Accounts.
func (accessInfo *AccessInfo) GetAccounts() *AccountCollection {
	accountCollection := AccountCollection{
		pageableCollection: accessInfo.newPageableCollection(false),
	}
	return &accountCollection
}

// GetAccount gets an Account with a specified accountId.
func (accessInfo *AccessInfo) GetAccount(accountId *uuid.UUID) (*Account, error) {
	// GET /accounts/<account-id>
	var account Account
	_, err := accessInfo.request(http.MethodGet, fmt.Sprintf("/accounts/%s", accountId), "", &account)
	if err != nil {
		return nil, err
	}
	return &account, err
}

// GetAccountHistory gets all AccountHistorys with a specified accountId.
func (accessInfo *AccessInfo) GetAccountHistory(accountId *uuid.UUID) *AccountHistoryCollection {
	accountHistoryCollection := AccountHistoryCollection{
		pageableCollection: accessInfo.newPageableCollection(true),
		id:                 accountId,
	}
	return &accountHistoryCollection
}

// GetAccountHolds gets all AcountHolds with a specified accountId.
func (accessInfo *AccessInfo) GetAccountHolds(accountId *uuid.UUID) *AccountHoldCollection {
	accountHoldCollection := AccountHoldCollection{
		pageableCollection: accessInfo.newPageableCollection(true),
		id:                 accountId,
	}
	return &accountHoldCollection
}

// HasNext determines if there is another Account in this iterator.
func (c *AccountCollection) HasNext() bool {
	// GET /accounts
	var accounts []Account
	return c.pageableCollection.hasNext(http.MethodGet, "/accounts", "", "", &accounts)
}

// HasNext determines if there is another AccountHistory in this iterator.
func (c *AccountHistoryCollection) HasNext() bool {
	// GET /accounts/<account-id>
	var accountHistory []AccountHistory
	return c.pageableCollection.hasNext(http.MethodGet, fmt.Sprintf("/accounts/%s/ledger", c.id), "", "", &accountHistory)
}

// HasNext determines if there is another AccountHold in this iterator.
func (c *AccountHoldCollection) HasNext() bool {
	// GET /accounts/<account-id>/holds
	var accountHolds []AccountHold
	return c.pageableCollection.hasNext(http.MethodGet, fmt.Sprintf("/accounts/%s/holds", c.id), "", "", &accountHolds)
}

// Next gets the next Account from the iterator.
func (c *AccountCollection) Next() (*Account, error) {
	account, err := c.pageableCollection.next()
	if err != nil {
		return nil, err
	}
	return account.Addr().Interface().(*Account), nil
}

// Next gets the next AccountHistory from the iterator.
func (c *AccountHistoryCollection) Next() (*AccountHistory, error) {
	history, err := c.pageableCollection.next()
	if err != nil {
		return nil, err
	}
	return history.Addr().Interface().(*AccountHistory), nil
}

// Next gets the next AccountHold from the iterator.
func (c *AccountHoldCollection) Next() (*AccountHold, error) {
	hold, err := c.pageableCollection.next()
	if err != nil {
		return nil, err
	}
	return hold.Addr().Interface().(*AccountHold), nil
}
