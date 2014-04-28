/*
  btcrobot is a Bitcoin, Litecoin and Altcoin trading bot written in golang,
  it features multiple trading methods using technical analysis.

  Disclaimer:

  USE AT YOUR OWN RISK!

  The author of this project is NOT responsible for any damage or loss caused
  by this software. There can be bugs and the bot may not Tick as expected
  or specified. Please consider testing it first with paper trading /
  backtesting on historical data. Also look at the code to see what how
  it's working.

  Weibo:http://weibo.com/bocaicfa
*/

package strategy

import (
	. "common"
	. "config"
	"fmt"
	// "math/rand"
	"sort"
	"strconv"
	"time"
)

type CollectorStrgatey struct {
	lastSellBot int
	lastBuyTop  int

	lastSellTop int
	lastBuyBot  int

	KeepApart int
	Dist      int

	Count int
	Unit  string
}

func init() {
	collector := new(CollectorStrgatey)
	Register("collector", collector)

	collector.lastSellTop, _ = strconv.Atoi(CollectorOption["lastSellTop"])
	collector.lastSellBot, _ = strconv.Atoi(CollectorOption["lastSellBot"])
	collector.lastBuyTop, _ = strconv.Atoi(CollectorOption["lastBuyTop"])
	collector.lastBuyBot, _ = strconv.Atoi(CollectorOption["lastBuyBot"])
	collector.Dist, _ = strconv.Atoi(CollectorOption["Dist"])
	collector.KeepApart, _ = strconv.Atoi(CollectorOption["KeepApart"])

	collector.Count, _ = strconv.Atoi(CollectorOption["Count"])
	collector.Unit = CollectorOption["Unit"]

	fmt.Println("Asign collector:", collector)
}

type ByPrice []HBOrder

func (a ByPrice) Len() int      { return len(a) }
func (a ByPrice) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByPrice) Less(i, j int) bool {

	m, err := strconv.ParseFloat(a[i].Order_price, 64)
	if err != nil {
		m = 0
	}

	n, err := strconv.ParseFloat(a[j].Order_price, 64)
	if err != nil {
		n = 0
	}
	return n < m
}

func (collector *CollectorStrgatey) patchMargin(nowSellBot int, nowBuyTop int) {
	for i := 0; i < collector.lastBuyTop-nowBuyTop; i += collector.Dist {
		p := nowBuyTop + i + collector.Dist

		for {
			buyId := gTradeAPI.Buy(strconv.FormatFloat(float64(p)+0.27, 'f', 2, 64), collector.Unit)
			if buyId != "0" {
				fmt.Printf("Buy %s success %d \n", buyId, p)
				collector.Count += 1
				break
			} else {
				fmt.Printf("Buy...")
			}
		}
	}

	for i := 0; i < nowSellBot-collector.lastSellBot; i += collector.Dist {
		p := nowSellBot - i - collector.Dist

		for {
			sellId := gTradeAPI.Sell(strconv.FormatFloat(float64(p)+0.73, 'f', 2, 64), collector.Unit)
			if sellId != "0" {
				fmt.Printf("Sell %s success %d\n", sellId, p)
				collector.Count += 1
				break
			} else {
				fmt.Printf("Sell...")
			}
		}
	}

	CollectorOption["lastSellTop"] = strconv.Itoa(collector.lastSellTop)
	CollectorOption["lastSellBot"] = strconv.Itoa(collector.lastSellBot)
	CollectorOption["lastBuyTop"] = strconv.Itoa(collector.lastBuyTop)
	CollectorOption["lastBuyBot"] = strconv.Itoa(collector.lastBuyBot)
	CollectorOption["Dist"] = strconv.Itoa(collector.Dist)
	CollectorOption["KeepApart"] = strconv.Itoa(collector.KeepApart)

	CollectorOption["Count"] = strconv.Itoa(collector.Count)

	CollectorOption["Unit"] = collector.Unit

	err := SaveCollectorOption()
	if err != nil {
		fmt.Println("写入Collector配置数据失败")
		return
	}

}

func GetIntOrderPrice(op string) (price int) {
	p, err := strconv.ParseFloat(op, 64)
	if err != nil {
		p = 0
		fmt.Println("price is not float:", op)
	}

	price = int(p)
	return price
}

//collector strategy
func (collector *CollectorStrgatey) Tick(records []Record) bool {
	accout_info, _ := gTradeAPI.GetAccount()
	fmt.Println(accout_info)

	// for i := 0; i < 10; i++ {
	// 	p := 1000.27 + float64(rand.Intn(100))

	// 	buyId := gTradeAPI.Buy(strconv.FormatFloat(p, 'f', 2, 64), "0.001")

	// 	fmt.Println(buyId, p)
	// }

	// for i := 0; i < 10; i++ {
	// 	p := 8000.73 + float64(rand.Intn(100))

	// 	sellId := gTradeAPI.Sell(strconv.FormatFloat(p, 'f', 2, 64), "0.001")

	// 	fmt.Println(sellId, p)
	// }

	ret, orders := true, []HBOrder{}
	for {
		ret, orders = gTradeAPI.GetOrders()
		if ret {
			time.Sleep(time.Second)
			ret2, orders2 := gTradeAPI.GetOrders()

			if ret2 && len(orders2) == len(orders) {
				break
			}
		}
	}
	sort.Sort(ByPrice(orders))
	isNoOrders, isNoBuyOrders, isNoSellOrders := true, true, true
	for _, od := range orders {
		isNoOrders = false

		if od.Type == 1 {
			isNoBuyOrders = false
		}

		if od.Type == 2 {
			isNoSellOrders = false
		}
	}

	if collector.Dist == 0 {
		collector.Dist = 5
	}

	// if price > collector.lastSellTop || price < collector.lastBuyBot {
	// 	fmt.Println("Price out range, noop...")
	// 	return false
	// }

	if isNoOrders {
		fmt.Println("No sell order, or no buy order")

		if collector.lastSellTop == 0 ||
			collector.lastSellBot == 0 ||
			collector.lastBuyTop == 0 ||
			collector.lastBuyBot == 0 { // if start:

			price := 0
			for {
				ret, orderbook := gTradeAPI.GetOrderBook()
				if ret {
					// fmt.Println(orderbook.Asks)
					// fmt.Println(orderbook.Bids)
					if len(orderbook.Asks) < 1 || len(orderbook.Bids) < 1 {
						fmt.Println("No Price")
						return false
					}
					pa := orderbook.Asks[9].Price
					pb := orderbook.Bids[0].Price

					price = int((pa+pb)/2) / collector.Dist * collector.Dist
					fmt.Println("Price", price, pa, pb)
					break
				}
			}

			if price < 100 {
				fmt.Println("Price", price)
				return false
			}

			collector.lastSellTop = price + collector.Dist*collector.KeepApart
			collector.lastSellBot = price + collector.Dist

			collector.lastBuyTop = price
			collector.lastBuyBot = price - collector.Dist*collector.KeepApart
		}

		collector.patchMargin(collector.lastSellTop, collector.lastBuyBot)

		fmt.Println("=== Init OK, let's go! ===")
		return false
	}

	nowSellTop := GetIntOrderPrice(orders[0].Order_price)
	nowBuyBot := GetIntOrderPrice(orders[len(orders)-1].Order_price)

	nowSellBot, nowBuyTop := 0, 0

	if isNoSellOrders {
		nowBuyTop = GetIntOrderPrice(orders[0].Order_price)
		nowSellBot = collector.lastSellTop
	} else if isNoBuyOrders {
		nowBuyTop = collector.lastBuyBot
		nowSellBot = GetIntOrderPrice(orders[len(orders)-1].Order_price)
	} else {
		for i, od := range orders {
			if od.Type == 1 {
				nowBuyTop = GetIntOrderPrice(od.Order_price)
				nowSellBot = GetIntOrderPrice(orders[i-1].Order_price)
				break
			}
		}
	}

	fmt.Println("=== Now: ===", collector.Count)
	fmt.Println(nowSellTop,
		nowSellBot,
		nowBuyTop,
		nowBuyBot)

	collectorlastSellBot := nowSellBot - (collector.lastBuyTop - nowBuyTop)

	collector.lastBuyTop = nowBuyTop + (nowSellBot - collector.lastSellBot)

	collector.lastSellBot = collectorlastSellBot

	fmt.Println("=== Next: ===")
	fmt.Println(collector.lastSellTop,
		collector.lastSellBot,
		collector.lastBuyTop,
		collector.lastBuyBot)

	collector.patchMargin(nowSellBot, nowBuyTop)

	// for {
	// 	ret, orders = gTradeAPI.GetOrders()
	// 	if ret {
	// 		break
	// 	}
	// }
	// sort.Sort(ByPrice(orders))
	// fmt.Println("Orders:")
	// for _, od := range orders {
	// 	fmt.Println(od)
	// }

	// for _, od := range orders {
	// 	fmt.Println(od.Id,
	// 		od.Type,
	// 		od.Order_price,
	// 		od.Order_amount,
	// 		od.Processed_amount,
	// 		od.Order_time,
	// 	)

	// 	sellId := strconv.Itoa(od.Id)
	// 	for {
	// 		if gTradeAPI.CancelOrder(sellId) {
	// 			fmt.Printf("cancel %s success \n", sellId)
	// 			break
	// 		} else {
	// 			fmt.Printf("cancel %s falied \n", sellId)
	// 		}
	// 	}
	// }

	fmt.Println("=== Collected OK, Next time will be better! ===")
	// 实现自己的策略
	return false
}
