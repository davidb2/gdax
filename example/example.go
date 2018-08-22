package main

import (
	"flag"
	"io/ioutil"
	"log"
	"math/big"

	"github.com/google/uuid"
	"github.com/ljeabmreosn/gdax"
)

var logToStdout bool
var linearRegression *LinearRegression

type LinearRegression struct {
	sumX      *big.Float
	sumY      *big.Float
	sumXY     *big.Float
	sumXX     *big.Float
	n         *big.Float
	slope     *big.Float
	intercept *big.Float
}

func NewLinearRegression() *LinearRegression {
	return &LinearRegression{
		sumX:      big.NewFloat(0.0),
		sumY:      big.NewFloat(0.0),
		sumXY:     big.NewFloat(0.0),
		sumXX:     big.NewFloat(0.0),
		n:         big.NewFloat(0.0),
		slope:     big.NewFloat(0.0),
		intercept: big.NewFloat(0.0),
	}
}

func (lr *LinearRegression) AddPoint(x, y float64) {
	xf := big.NewFloat(x)
	yf := big.NewFloat(y)
	af := big.NewFloat(0.0)
	bf := big.NewFloat(0.0)
	cf := big.NewFloat(0.0)
	denom := big.NewFloat(0.0)
	lr.sumX.Add(lr.sumX, xf)
	lr.sumY.Add(lr.sumY, yf)
	lr.sumXY.Add(lr.sumXY, af.Mul(xf, yf))
	lr.sumXX.Add(lr.sumXY, af.Mul(xf, xf))
	lr.n.Add(lr.n, big.NewFloat(0.0))
	denom.Sub(af.Mul(lr.n, lr.sumXX), bf.Mul(lr.sumX, lr.sumX))
	lr.slope.Quo(cf.Sub(af.Mul(lr.n, lr.sumXY), bf.Mul(lr.sumX, lr.sumY)), denom)
	lr.intercept.Quo(cf.Sub(af.Mul(lr.sumXX, lr.sumY), bf.Mul(lr.sumXY, lr.sumX)), denom)
}

func (lr *LinearRegression) GetCoefficients() (*big.Float, *big.Float) {
	return lr.slope, lr.intercept
}

func ParseFlags() {
	flag.BoolVar(&logToStdout, "logtostdout", false, "determine whether or not log")
	flag.Parse()
}

func performLinearRegression() error {
	err := gdax.Feed(&gdax.Subscription{
		Type:       gdax.SubscribeType,
		Channels:   []string{gdax.MatchesType},
		ProductIds: []string{"BTC-USD"},
	},
		func(message gdax.Message) {
			switch message.MessageType() {
			case gdax.ErrorType:
				log.Printf("error: %+v\n", message.(gdax.Error))
			case gdax.MatchType:
				match := message.(gdax.Match)
				linearRegression.AddPoint(float64(match.Time.Unix()), match.Price)
				a, b := linearRegression.GetCoefficients()
				log.Println(a, b, match.Price)
			default:
				log.Printf("unknown: %+v\n", message)
			}
		})
	return err
}

func main() {
	ParseFlags()
	if !logToStdout {
		log.SetOutput(ioutil.Discard)
	}

	accessInfo, err := gdax.RetrieveAccessInfoFromFile("keys.json")
	if err != nil {
		log.Panic(err)
	}
	accountId, err := uuid.Parse("7f757a16-7f1f-4985-8244-f1ff2e803b33")
	log.Println(accountId)
	if err != nil {
		log.Panic(err)
	}
	for accounts := accessInfo.GetAccountHistory(&accountId); accounts.HasNext(); {
		history, err := accounts.Next()
		if err != nil {
			log.Panic(err)
		}
		log.Printf("%+v\n", history)
	}
	// cancelledOrders, err := accessInfo.CancelAllOrders()
	// if err != nil {
	//   log.Panic(err)
	// }
	// log.Println("Cancelled all orders", cancelledOrders)
	// orderResponse, err := accessInfo.PlaceLimitOrder(&gdax.Order{
	//   Side: gdax.Buy,
	//   ProductId: "BTC-USD",
	//   Price: 700,
	//   Size: 0.01,
	// })
	// if err != nil {
	//   log.Panic(err)
	// }
	// log.Printf("%+v\n", orderResponse)

	// time.Sleep(0 * time.Second)
	// for orders := accessInfo.GetOrdersForProduct("BTC-USD"); orders.HasNext(); {
	//   order, err := orders.Next()
	//   if err != nil {
	//     log.Panic(err)
	//   }
	//   log.Println("order", order.Id)
	//   o, err := accessInfo.GetOrder(order.Id)
	//   if err != nil {
	//     log.Panic(err)
	//   }
	//   log.Println("order confirmation", o)
	// }

	// for fills := accessInfo.GetFills(); fills.HasNext(); {
	//   fill, err := fills.Next()
	//   if err != nil {
	//     log.Panic(err)
	//   }
	//   log.Println("fill", fill)
	// }

	// for cas := accessInfo.GetCoinbaseAccounts(); cas.HasNext(); {
	//   ca, err := cas.Next()
	//   if err != nil {
	//     log.Panic(err)
	//   }
	//   log.Println("coinbase account", ca)
	// }

	// for accounts := accessInfo.GetAccounts(); accounts.HasNext(); {
	//   account, err := accounts.Next()
	//   if err != nil {
	//     log.Panic(err)
	//   }
	//   log.Printf("account: %+v\n", account)
	// }

	// startDate, err := time.Parse(time.RFC3339, "2014-11-01T00:00:00.000Z")
	// if err != nil {
	//   log.Panic(err)
	// }
	// endDate, err := time.Parse(time.RFC3339, "2014-11-30T23:59:59.000Z")
	// if err != nil {
	//   log.Panic(err)
	// }
	// report, err := accessInfo.CreateReport(&gdax.Report{
	//   Type: gdax.Fills,
	//   StartDate: &startDate,
	//   EndDate: &endDate,
	// })
	// if err != nil {
	//   log.Panic(err)
	// }
	// log.Printf("report: %+v\n", report)

	// linearRegression = NewLinearRegression()
	// if err := performLinearRegression(); err != nil {
	// 	panic(err)
	// }

}
