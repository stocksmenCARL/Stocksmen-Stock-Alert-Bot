package handlers

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/m1k8/kronos/pkg/M1K8/harpe/pkg/db"
	"github.com/m1k8/kronos/pkg/M1K8/harpe/pkg/utils"
	stonks "github.com/m1k8/kronos/pkg/M1K8/nabu/pkg/fetcher"
	"github.com/m1k8/kronos/pkg/discord"
	"github.com/m1k8/kronos/pkg/worker"
)

func getAlertEmbed(s *discordgo.Session, channelID, mention string, args ...*discordgo.ApplicationCommandInteractionDataOption) *discordgo.MessageEmbed {
	desc := ""
	amount := len(args)

	if amount < 1 {
		discord.SendError(s, channelID, errors.New("incorrect number of args"))
		return nil
	}

	desc += fmt.Sprintf("***%v*** says: ", mention)

	if amount >= 1 {
		for i := 0; i < amount; i++ {
			desc += " " + args[i].StringValue()
		}
	}

	return discord.GetMessageEmbed(utils.Neutral, mention, desc)
}

func refresh(s *discordgo.Session, channelID, guildID string) string {
	d := db.NewDB(guildID)

	stocks, shorts, crypto, options, err := d.RefreshFromDB()
	respStr := ""

	if err != nil {
		return err.Error()
	}

	respStr += "__Stocks__\n"
	for _, v := range stocks {
		if d.GetExitChan("s_"+v.StockTicker) == nil {
			exitChan := make(chan bool, 1)
			go worker.StockTracker(s, channelID, v.StockTicker, v.TraderID, v.Caller, fmt.Sprintf("%d", v.StockExpiry), d, nil, d.SetAndReturnNewExitChan("s_"+v.StockTicker, exitChan))

			respStr += v.StockTicker + "\n"
		}
	}

	respStr += "__Shorts__\n"
	for _, v := range shorts {
		if d.GetExitChan("sh_"+v.ShortTicker) == nil {
			exitChan := make(chan bool, 1)
			go worker.ShortTracker(s, channelID, v.ShortTicker, v.TraderID, v.Caller, fmt.Sprintf("%d", v.ShortExpiry), d, nil, d.SetAndReturnNewExitChan("sh_"+v.ShortTicker, exitChan))
			respStr += v.ShortTicker + "\n"
		}
	}

	respStr += "__Options__\n"
	for _, v := range options {
		if d.GetExitChan(v.OptionUid) == nil {
			exitChan := make(chan bool, 1)
			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			go worker.OptionsTracker(s, channelID, v.Caller, v.OptionTicker, v.TraderID, v.OptionContractType, v.OptionDay, v.OptionMonth, v.OptionYear, v.OptionStrike, d, nil, d.SetAndReturnNewExitChan(v.OptionUid, exitChan))
			time.Sleep(1 + time.Duration(r.Intn(7))*time.Second)
			respStr += v.OptionUid + "\n"
		}
	}

	respStr += "__Crypto__\n"
	for _, v := range crypto {
		if d.GetExitChan("c_"+v.CryptoCoin) == nil {
			exitChan := make(chan bool, 1)
			go worker.CryptoTracker(s, channelID, v.CryptoCoin, v.TraderID, v.Caller, fmt.Sprintf("%d", v.CryptoExpiry), d, nil, d.SetAndReturnNewExitChan("c_"+v.CryptoCoin, exitChan))
			respStr += v.CryptoCoin + "\n"
		}
	}

	return respStr
}

func all(s *discordgo.Session, channelID, guildID string) []string {
	d := db.NewDB(guildID)
	allStocks, allShorts, allCrypto, allOptions, err := d.GetAll()

	if err != nil {
		return nil
	}

	fetcher := stonks.NewFetcher()

	var (
		stockStr  string
		stockStrD string
		shortStrD string
		stockStrL string
		shortStrL string
		shortStr  string
		cryptoStr string
		optiStrL  string
		optiStrD  string
		optiStr   string
	)

	if len(allStocks) > 0 {
		stockStr = "\n**__Stocks__**\n"

		for _, v := range allStocks {
			OGPrice := v.StockStarting
			newPrice, err := fetcher.GetStock(v.StockTicker)

			tradeType := ""

			switch v.ChannelType {
			case discord.DAY:
				tradeType = "Day Trade"

				if err != nil {
					stockStrD += fmt.Sprintf("Unable to fetch current price for **%v**. Starting price: *$%.2f* - %v", v.StockTicker, OGPrice, tradeType) + "\n"
				} else {
					stockStrD += fmt.Sprintf("**%v** starting price: *$%.2f*, current price: *$%.2f* - %v", v.StockTicker, OGPrice, newPrice, tradeType)
					if v.StockEPt != 0 {
						stockStrD += fmt.Sprintf(" Exit PT: $%.2f ", v.StockEPt)
					}
					if v.StockSPt != 0 {
						stockStrD += fmt.Sprintf(" Scale PT: $%.2f ", v.StockSPt)
					}
					stockStrD += "\n"
				}
			case discord.SWING:
				tradeType = "Long Trade"

				if err != nil {
					stockStrL += fmt.Sprintf("Unable to fetch current price for **%v**. Starting price: *$%.2f* - %v", v.StockTicker, OGPrice, tradeType) + "\n"
				} else {
					stockStrL += fmt.Sprintf("**%v** starting price: *$%.2f*, current price: *$%.2f* - %v", v.StockTicker, OGPrice, newPrice, tradeType)
					if v.StockEPt != 0 {
						stockStrL += fmt.Sprintf(" Exit PT: $%.2f ", v.StockEPt)
					}
					if v.StockSPt != 0 {
						stockStrL += fmt.Sprintf(" Scale PT: $%.2f ", v.StockSPt)
					}
					stockStrD += "\n"
				}
			case discord.WATCHLIST:
				tradeType = "Watchlist"
			}

		}
	}

	if len(allShorts) > 0 {
		shortStr += "\n**__Shorts__**\n"

		for _, v := range allShorts {
			OGPrice := v.ShortStarting

			newPrice, err := fetcher.GetStock(v.ShortTicker)

			tradeType := ""

			switch v.ChannelType {
			case discord.DAY:
				tradeType = "Day Trade"
				if err != nil {
					shortStrD += fmt.Sprintf("Unable to fetch current price for **%v**. Starting price: *$%.2f* - %v", v.ShortTicker, OGPrice, tradeType) + "\n"
				} else {
					shortStrD += fmt.Sprintf("**%v** starting price: *$%.2f*, current price: *$%.2f* - %v", v.ShortTicker, OGPrice, newPrice, tradeType)
					if v.ShortEPt != 0 {
						shortStrD += fmt.Sprintf(" Exit PT: $%.2f ", v.ShortEPt)
					}
					if v.ShortSPt != 0 {
						shortStrD += fmt.Sprintf(" Scale PT: $%.2f ", v.ShortSPt)
					}
					shortStrD += "\n"
				}
			case discord.SWING:
				tradeType = "Long Trade"
				if err != nil {
					shortStrL += fmt.Sprintf("Unable to fetch current price for **%v**. Starting price: *$%.2f* - %v", v.ShortTicker, OGPrice, tradeType) + "\n"
				} else {
					shortStrL += fmt.Sprintf("**%v** starting price: *$%.2f*, current price: *$%.2f* - %v", v.ShortTicker, OGPrice, newPrice, tradeType)
					if v.ShortEPt != 0 {
						shortStrL += fmt.Sprintf(" Exit PT: $%.2f ", v.ShortEPt)
					}
					if v.ShortSPt != 0 {
						shortStrL += fmt.Sprintf(" Scale PT: $%.2f ", v.ShortSPt)
					}
					shortStrL += "\n"
				}
			case discord.WATCHLIST:
				tradeType = "Watchlist"
			}

		}
	}

	if len(allCrypto) > 0 {
		cryptoStr += "\n**__Crypto__**\n"

		for _, v := range allCrypto {
			OGPrice := v.CryptoStarting

			newPrice, err := fetcher.GetCrypto(v.CryptoCoin, false)
			tradeType := ""

			switch v.ChannelType {
			case discord.DAY:
				tradeType = "Day Trade"
			case discord.SWING:
				tradeType = "Long Trade"
			case discord.WATCHLIST:
				tradeType = "Watchlist"
			}

			if err != nil {
				cryptoStr += fmt.Sprintf("Unable to fetch current price for **%v**. Starting price: *$%.8f* - %v", v.CryptoCoin, OGPrice, tradeType) + "\n"
			} else {
				cryptoStr += fmt.Sprintf("**%v** starting price: *$%.8f*, current price: *$%.8f* - %v", v.CryptoCoin, OGPrice, newPrice, tradeType)
				if v.CryptoEPt != 0 {
					cryptoStr += fmt.Sprintf(" Exit PT: $%.2f ", v.CryptoEPt)
				}
				if v.CryptoSPt != 0 {
					cryptoStr += fmt.Sprintf(" Scale PT: $%.2f ", v.CryptoSPt)
				}
				cryptoStr += "\n"
			}
		}
	}

	if len(allOptions) > 0 {
		optiStr += "\n**__Options__**\n"

		for _, v := range allOptions {
			OGPrice := v.OptionStarting
			ticker, cType, day, month, year, price, err := db.SplitOptionsCode(v.OptionUid)
			prettyOID := utils.NiceStr(ticker, cType, day, month, year, price)
			if err != nil {
				log.Println(fmt.Sprintf("Unable to parse options code for %v. It is likely in a corrupt state and should be removed: %v.", prettyOID, err.Error()))
			}

			newPrice, _, err := fetcher.GetOption(ticker, cType, day, month, year, price, 0)

			tradeType := ""

			switch v.ChannelType {
			case discord.DAY:
				tradeType = "Day Trade"
				if err != nil {
					log.Println(fmt.Sprintf("Unable to get option %v. Crash?: %v.", prettyOID, err.Error()))
					optiStrD += fmt.Sprintf("Unable to fetch current price for **%v**. Starting price: *$%.2f* - %v", prettyOID, OGPrice, tradeType) + "\n"
				} else {
					optiStrD += fmt.Sprintf("**%v** starting price: *$%.2f*, current price: *$%.2f* - %v", prettyOID, OGPrice, newPrice, tradeType) + "\n"
				}
			case discord.SWING:
				tradeType = "Long Trade"
				if err != nil {
					log.Println(fmt.Sprintf("Unable to get option %v. Crash?: %v.", prettyOID, err.Error()))
					optiStrL += fmt.Sprintf("Unable to fetch current price for **%v**. Starting price: *$%.2f* - %v", prettyOID, OGPrice, tradeType) + "\n"
				} else {
					optiStrL += fmt.Sprintf("**%v** starting price: *$%.2f*, current price: *$%.2f* - %v", prettyOID, OGPrice, newPrice, tradeType) + "\n"
				}
			case discord.WATCHLIST:
				tradeType = "Watchlist"
			}

		}
	}

	fmt.Println("stockStr is " + stockStr)
	fmt.Println("stockStrL is " + stockStrL)
	fmt.Println("stockStrD is " + stockStrD)
	fmt.Println("shortStr is " + shortStr)
	fmt.Println("shortStrL is " + shortStrL)
	fmt.Println("shortStrD is " + shortStrD)
	fmt.Println("cryptoStr is " + cryptoStr)
	fmt.Println("optiStr is " + optiStr)
	fmt.Println("optiStrL is " + optiStrL)
	fmt.Println("optiStrD is " + optiStrD)

	if optiStr == "" && stockStr == "" && shortStr == "" && cryptoStr == "" {
		return []string{"**No alerts to show**"}
	} else {
		respStrs := make([]string, 0)
		if !db.IsTradingHours() {
			respStrs = append(respStrs, "**Warning** - currently outside of market hours - data may not be accurate\n")
		}
		if stockStr != "" {
			respStrs = append(respStrs, stockStr)
			if shortStrD != "" {
				respStrs = append(respStrs, stockStrD, "\n")
			}
			if shortStrL != "" {
				respStrs = append(respStrs, stockStrL, "\n")
			}
			respStrs = append(respStrs, "\n")
		}
		if shortStr != "" {
			respStrs = append(respStrs, shortStr)
			if shortStrL != "" {
				respStrs = append(respStrs, shortStrL, "\n")
			}
			if shortStrD != "" {
				respStrs = append(respStrs, shortStrD, "\n")
			}
			respStrs = append(respStrs, "\n")
		}
		if cryptoStr != "" {
			respStrs = append(respStrs, cryptoStr)
		}
		if optiStr != "" {
			respStrs = append(respStrs, optiStr)
			if optiStrL != "" {
				respStrs = append(respStrs, optiStrL, "\n")
			}
			if optiStrD != "" {
				respStrs = append(respStrs, optiStrD, "\n")
			}
			respStrs = append(respStrs, "\n")
		}
		return respStrs
	}
}
