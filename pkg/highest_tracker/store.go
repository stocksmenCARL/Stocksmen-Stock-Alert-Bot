package highest_tracker

import (
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/m1k8/kronos/pkg/M1K8/harpe/pkg/db"
	"github.com/m1k8/kronos/pkg/M1K8/harpe/pkg/utils"
)

type AllAlerts struct {
	stocks  sync.Map
	shorts  sync.Map
	crypto  sync.Map
	options sync.Map
}

type AlertObj struct {
	Starting float32
	Peak     float32
	Closing  float32
	PctDiff  float32
}

type AL_TYPE int

const (
	STOCK_TYPE AL_TYPE = iota
	SHORT_TYPE
	CRYPTO_TYPE
	OPTION_TYPE
)

var all = sync.Map{}

func GetAlertsGuild(guildID, traderID string) *AllAlerts {
	if v, ok := all.Load(fmt.Sprintf("%s_%s", guildID, traderID)); !ok {
		newAlerts := &AllAlerts{
			stocks:  sync.Map{},
			shorts:  sync.Map{},
			crypto:  sync.Map{},
			options: sync.Map{},
		}
		all.Store(fmt.Sprintf("%s_%s", guildID, traderID), newAlerts)
		return newAlerts
	} else {
		return v.(*AllAlerts)
	}
}

func (a *AllAlerts) WriteResult(t AL_TYPE, alert db.Alert, newPrice float32) error {
	log.Println("writing Result")
	switch t {
	case SHORT_TYPE:
		sh := alert.(*db.Short)
		a.shorts.Store(sh.ShortTicker, &AlertObj{
			Starting: sh.ShortStarting,
			Peak:     newPrice,
			PctDiff:  alert.GetPctGain(newPrice),
			Closing:  0,
		})
		return nil
	case STOCK_TYPE:
		s := alert.(*db.Stock)
		a.stocks.Store(s.StockTicker, &AlertObj{
			Starting: s.StockStarting,
			Peak:     newPrice,
			PctDiff:  alert.GetPctGain(newPrice),
			Closing:  0,
		})
		return nil
	case CRYPTO_TYPE:
		c := alert.(*db.Crypto)
		a.crypto.Store(c.CryptoCoin, &AlertObj{
			Starting: c.CryptoStarting,
			Peak:     newPrice,
			PctDiff:  alert.GetPctGain(newPrice),
			Closing:  0,
		})
		return nil
	case OPTION_TYPE:
		o := alert.(*db.Option)
		niceStr := utils.NiceStr(o.OptionTicker, o.OptionContractType, o.OptionDay, o.OptionMonth, o.OptionYear, o.OptionStrike)
		a.options.Store(niceStr, &AlertObj{
			Starting: o.OptionStarting,
			Peak:     newPrice,
			PctDiff:  alert.GetPctGain(newPrice),
			Closing:  0,
		})
		return nil
	default:
		return errors.New("nope")
	}
}

func (a *AllAlerts) Clear() {
	a.stocks = sync.Map{}
	a.shorts = sync.Map{}
	a.crypto = sync.Map{}
	a.options = sync.Map{}
}

func (a *AllAlerts) GetWinners() int {
	total := 0
	a.stocks.Range(func(k interface{}, v interface{}) bool {
		vS := v.(*AlertObj)
		if vS.PctDiff > 5 {
			total++
		}
		return true
	})

	a.shorts.Range(func(k interface{}, v interface{}) bool {
		vS := v.(*AlertObj)
		if vS.PctDiff > 5 {
			total++
		}
		return true
	})

	a.crypto.Range(func(k interface{}, v interface{}) bool {
		vS := v.(*AlertObj)
		if vS.PctDiff > 5 {
			total++
		}
		return true
	})

	a.options.Range(func(k interface{}, v interface{}) bool {
		vS := v.(*AlertObj)
		if vS.PctDiff > 5 {
			total++
		}
		return true
	})

	return total
}

func (a *AllAlerts) GetAll() (stocks, shorts, crypto, opts map[string]*AlertObj, total int) {
	stocks = a.GetStocks()
	shorts = a.GetShorts()
	crypto = a.GetCrypto()
	opts = a.GetOptions()

	total = len(stocks) + len(shorts) + len(crypto) + len(opts)

	return
}

func (a *AllAlerts) GetCrypto() map[string]*AlertObj {
	subTotal := 0
	retMap := make(map[string]*AlertObj)
	a.crypto.Range(func(k interface{}, v interface{}) bool {
		subTotal++
		retMap[k.(string)] = v.(*AlertObj)
		return true
	})

	return retMap
}

func (a *AllAlerts) GetStocks() map[string]*AlertObj {
	subTotal := 0
	retMap := make(map[string]*AlertObj)
	a.stocks.Range(func(k interface{}, v interface{}) bool {
		subTotal++
		retMap[k.(string)] = v.(*AlertObj)
		return true
	})
	return retMap
}

func (a *AllAlerts) GetShorts() map[string]*AlertObj {
	subTotal := 0
	retMap := make(map[string]*AlertObj)
	a.shorts.Range(func(k interface{}, v interface{}) bool {
		subTotal++
		retMap[k.(string)] = v.(*AlertObj)
		return true
	})

	return retMap
}

func (a *AllAlerts) GetOptions() map[string]*AlertObj {
	subTotal := 0
	retMap := make(map[string]*AlertObj)
	a.options.Range(func(k interface{}, v interface{}) bool {
		subTotal++
		retMap[k.(string)] = v.(*AlertObj)
		return true
	})

	return retMap
}
