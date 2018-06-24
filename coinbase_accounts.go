package gdax

import (
	"net/http"

	"github.com/google/uuid"
)

// A CoinbaseAccount stores information about a specific coinbase account.
type CoinbaseAccount struct {
	Id       *uuid.UUID `json:"id,string"`
	Name     string     `json:"name"`
	Balance  float64    `json:"balance,string"`
	Currency string     `json:"currency"`
	Type     string     `json:"wallet"`
	Primary  bool       `json:"primary"`
	Active   bool       `json:"active"`
}

// A CoinbaseAccountCollection is an iterator of CoinbaseAccounts.
type CoinbaseAccountCollection struct {
	pageableCollection
	requestSent bool
	pages       [][]CoinbaseAccount
}

// GetCoinbaseAccounts gets all coinbase accounts.
func (accessInfo *AccessInfo) GetCoinbaseAccounts() *CoinbaseAccountCollection {
	coinbaseAccountCollection := CoinbaseAccountCollection{
		pageableCollection: accessInfo.newPageableCollection(false),
		requestSent:        false,
		pages:              nil,
	}
	return &coinbaseAccountCollection
}

// HasNext determines if there is another CoinbaseAccount in this iterator.
func (c *CoinbaseAccountCollection) HasNext() bool {
	// GET /coinbase-accounts
	var col []CoinbaseAccount
	return c.pageableCollection.hasNext(http.MethodGet, "/coinbase-accounts", "", "", &col)
}

// Next gets the next CoinbaseAccount from the iterator.
func (c *CoinbaseAccountCollection) Next() (*CoinbaseAccount, error) {
	account, err := c.pageableCollection.next()
	if err != nil {
		return nil, err
	}
	return account.Addr().Interface().(*CoinbaseAccount), nil
}
