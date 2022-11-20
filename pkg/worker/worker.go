package worker

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/m1k8/kronos/pkg/M1K8/harpe/pkg/db"
	"github.com/m1k8/kronos/pkg/M1K8/harpe/pkg/utils"
	"github.com/m1k8/kronos/pkg/M1K8/nabu/pkg/background"
	"github.com/m1k8/kronos/pkg/discord"
	"github.com/m1k8/kronos/pkg/highest_tracker"
	"github.com/m1k8/kronos/pkg/twitter"
)

func CryptoTracker(s *discordgo.Session, channelID, coin, traderID, caller, expiry string, d *db.DB, ref *discordgo.MessageReference, exitChan chan bool) {

	b := background.NewBG(d.Guild, d)
	responseChannel := make(chan background.Response)
	coinDb, err := d.GetCrypto(coin, traderID)
	if err != nil {
		log.Println(err)
		return
	}

	go b.CheckCryptoPriceInBG(responseChannel, coin, traderID, expiry, d.Guild, caller, exitChan)

	for {
		select {
		case msg := <-responseChannel:
			switch msg.Type {

			case background.Exit:
				return

			case background.Error:
				log.Println(msg.Message)
				return

			case background.Expired:
				discord.AlertExpired(s, d.Guild, channelID, caller, coin, coinDb.ChannelType)
				err = d.RemoveCrypto(coin, traderID)
				if err != nil {
					log.Println(err)
				}

			case background.New_High:
				err = d.CryptoSetNewHigh(coin, traderID, msg.Price)
				if err != nil {
					log.Println(err)
				}
				if coinDb.ChannelType == discord.DAY {
					al := highest_tracker.GetAlertsGuild(coinDb.CryptoGuildID, traderID)

					if msg.Price >= coinDb.CryptoHighest {
						log.Println("Storing Crypto High for Tracker")
						err = al.WriteResult(highest_tracker.CRYPTO_TYPE, coinDb, msg.Price)
						if err != nil {
							log.Println(err)
						}
					}

				}

			case background.POI:
				discord.POI(s, d.Guild, channelID, coin, caller, coinDb.CryptoPoI, coinDb.ChannelType, ref)
				err = d.CryptoPOIHit(coin, traderID)
				if err != nil {
					log.Println(err)
				}

			case background.PT1:
				discord.AlertCryptoPT(s, d.Guild, channelID, caller, coin, coinDb.CryptoSPt, coinDb.ChannelType)

			case background.PT2:
				discord.AlertCryptoPTExit(s, d.Guild, channelID, caller, coin, coinDb.CryptoSPt, coinDb.ChannelType)
				err = d.CryptoSetNewHigh(coin, traderID, msg.Price)
				if err != nil {
					log.Println(err)
				}
				if coinDb.ChannelType == discord.DAY {
					al := highest_tracker.GetAlertsGuild(coinDb.CryptoGuildID, traderID)

					if msg.Price >= coinDb.CryptoHighest {
						log.Println("Storing Crypto High for Tracker")
						err = al.WriteResult(highest_tracker.CRYPTO_TYPE, coinDb, msg.Price)
						if err != nil {
							log.Println(err)
						}
					}
				}

				err = d.RemoveCrypto(coin, traderID)
				if err != nil {
					log.Println(err)
				}
			case background.SL:
				discord.StopLoss(s, d.Guild, channelID, coin, caller, coinDb.CryptoStop, coinDb.ChannelType, ref)
				err = d.RemoveCrypto(coin, traderID)
				if err != nil {
					log.Println(err)
				}

			case background.Price:
				discord.AlertCrypto(s, d.Guild, channelID, caller, coin, msg.PctGain, msg.Price, coinDb.ChannelType, ref)

				if msg.PctGain > 50 {
					twtClient, err := twitter.GetTwitterClient(d.Guild)
					if err != nil {
						log.Println(fmt.Errorf("%v - %w", coin, err))
					} else {
						go twtClient.SendDelayedTweet("Crypto", coin, coinDb.CryptoCallTime, msg.PctGain)
					}
				}
			}

		}
	}
}

func StockTracker(s *discordgo.Session, channelID, ticker, traderID, caller, expiry string, d *db.DB, ref *discordgo.MessageReference, exitChan chan bool) {

	b := background.NewBG(d.Guild, d)
	responseChannel := make(chan background.Response)
	tickerDb, err := d.GetStock(ticker, traderID)
	if err != nil {
		log.Println(err)
		return
	}

	go b.CheckStockPriceInBG(responseChannel, ticker, traderID, expiry, d.Guild, caller, exitChan)

	for {
		select {
		case msg := <-responseChannel:
			switch msg.Type {

			case background.Exit:
				return

			case background.Error:
				log.Println(err)
				return

			case background.Expired:
				discord.AlertExpired(s, d.Guild, channelID, caller, ticker, tickerDb.ChannelType)
				err = d.RemoveStock(ticker, traderID)
				if err != nil {
					log.Println(err)
				}

			case background.New_High:
				d.StockSetNewHigh(ticker, traderID, msg.Price)
				if tickerDb.ChannelType == discord.DAY {
					al := highest_tracker.GetAlertsGuild(tickerDb.StockGuildID, traderID)

					if msg.Price >= tickerDb.StockHighest {
						log.Println("Storing Stock High for Tracker")
						err = al.WriteResult(highest_tracker.STOCK_TYPE, tickerDb, msg.Price)
						if err != nil {
							log.Println(err)
						}
					}
				}

			case background.POI:
				discord.POI(s, d.Guild, channelID, ticker, caller, tickerDb.StockPoI, tickerDb.ChannelType, ref)
				err = d.StockPOIHit(ticker, traderID)
				if err != nil {
					log.Println(err)
				}

			case background.PT1:
				discord.AlertStockPT(s, d.Guild, channelID, caller, ticker, tickerDb.StockSPt, tickerDb.ChannelType)

			case background.PT2:
				discord.AlertStockPTExit(s, d.Guild, channelID, caller, ticker, tickerDb.StockSPt, tickerDb.ChannelType)
				err = d.StockSetNewHigh(ticker, traderID, msg.Price)
				if err != nil {
					log.Println(err)
				}
				if tickerDb.ChannelType == discord.DAY {
					al := highest_tracker.GetAlertsGuild(tickerDb.StockGuildID, traderID)

					if msg.Price >= tickerDb.StockHighest {
						log.Println("Storing Stock High for Tracker")
						err = al.WriteResult(highest_tracker.STOCK_TYPE, tickerDb, msg.Price)
						if err != nil {
							log.Println(err)
						}
					}
				}

				err = d.RemoveStock(ticker, traderID)
				if err != nil {
					log.Println(err)
				}
			case background.SL:
				discord.StopLoss(s, d.Guild, channelID, ticker, caller, tickerDb.StockStop, tickerDb.ChannelType, ref)
				err = d.RemoveStock(ticker, traderID)
				if err != nil {
					log.Println(err)
				}

			case background.Price:
				discord.AlertStock(s, d.Guild, channelID, caller, ticker, msg.PctGain, msg.Price, tickerDb.ChannelType, ref)

				if msg.PctGain > 50 {
					twtClient, err := twitter.GetTwitterClient(d.Guild)
					if err != nil {
						log.Println(fmt.Errorf("%v - %w", ticker, err))
					} else {
						go twtClient.SendDelayedTweet("Stock", ticker, tickerDb.StockCallTime, msg.PctGain)
					}
				}
			}

		}
	}
}

func ShortTracker(s *discordgo.Session, channelID, ticker, traderID, caller, expiry string, d *db.DB, ref *discordgo.MessageReference, exitChan chan bool) {

	b := background.NewBG(d.Guild, d)
	responseChannel := make(chan background.Response)
	tickerDb, err := d.GetShort(ticker, traderID)
	if err != nil {
		log.Println(err)
		return
	}

	go b.CheckShortPriceInBG(responseChannel, ticker, traderID, expiry, d.Guild, caller, exitChan)

	for {
		select {
		case msg := <-responseChannel:
			switch msg.Type {

			case background.Exit:
				return

			case background.Error:
				log.Println(msg.Message)
				return

			case background.Expired:
				discord.AlertExpired(s, d.Guild, channelID, caller, ticker, tickerDb.ChannelType)
				err = d.RemoveShort(ticker, traderID)
				if err != nil {
					log.Println(err)
				}

			case background.New_High:
				err = d.ShortSetNewHigh(ticker, traderID, msg.Price)
				if err != nil {
					log.Println(err)
				}
				if tickerDb.ChannelType == discord.DAY {
					al := highest_tracker.GetAlertsGuild(tickerDb.ShortGuildID, traderID)

					if msg.Price <= tickerDb.ShortLowest {
						log.Println("Storing Short High for Tracker")
						err = al.WriteResult(highest_tracker.SHORT_TYPE, tickerDb, msg.Price)
						if err != nil {
							log.Println(err)
						}
					}
				}

			case background.POI:
				discord.POI(s, d.Guild, channelID, ticker, caller, tickerDb.ShortPoI, tickerDb.ChannelType, ref)
				err = d.ShortPOIHit(ticker, traderID)
				if err != nil {
					log.Println(err)
				}

			case background.PT1:
				discord.AlertShortPT(s, d.Guild, channelID, caller, ticker, tickerDb.ShortSPt, tickerDb.ChannelType)

			case background.PT2:
				discord.AlertShortPTExit(s, d.Guild, channelID, caller, ticker, tickerDb.ShortSPt, tickerDb.ChannelType)
				err = d.ShortSetNewHigh(ticker, traderID, msg.Price)
				if err != nil {
					log.Println(err)
				}
				if tickerDb.ChannelType == discord.DAY {
					al := highest_tracker.GetAlertsGuild(tickerDb.ShortGuildID, traderID)

					if msg.Price <= tickerDb.ShortLowest {
						log.Println("Storing Short High for Tracker")
						err = al.WriteResult(highest_tracker.SHORT_TYPE, tickerDb, msg.Price)
						if err != nil {
							log.Println(err)
						}
					}
				}

				err = d.RemoveShort(ticker, traderID)
				if err != nil {
					log.Println(err)
				}
			case background.SL:
				discord.StopLoss(s, d.Guild, channelID, ticker, caller, tickerDb.ShortStop, tickerDb.ChannelType, ref)
				err = d.RemoveShort(ticker, traderID)
				if err != nil {
					log.Println(err)
				}

			case background.Price:
				discord.AlertShort(s, d.Guild, channelID, caller, ticker, msg.PctGain, msg.Price, tickerDb.ChannelType, ref)

				if msg.PctGain > 50 {
					twtClient, err := twitter.GetTwitterClient(d.Guild)
					if err != nil {
						log.Println(fmt.Errorf("%v - %w", ticker, err))
					} else {
						go twtClient.SendDelayedTweet("Short", ticker, tickerDb.ShortCallTime, msg.PctGain)
					}
				}
			}

		}
	}
}

func OptionsTracker(s *discordgo.Session, channelID, caller, ticker, traderID, contractType, day, month, year string, price float32, d *db.DB, ref *discordgo.MessageReference, exitChan chan bool) {

	b := background.NewBG(d.Guild, d)
	responseChannel := make(chan background.Response)
	oID := utils.GetCode(ticker, contractType, day, month, year, price)
	prettyStr := utils.NiceStr(ticker, contractType, day, month, year, price)
	optionDb, err := d.GetOption(oID, traderID)
	if err != nil {
		log.Println(err)
		return
	}

	go b.CheckOptionsPriceInBG(responseChannel, d.Guild, caller, ticker, traderID, contractType, day, month, year, price, exitChan)

	for {
		select {
		case msg := <-responseChannel:
			switch msg.Type {

			case background.Exit:
				return

			case background.Error:
				log.Println(msg.Message)
				return

			case background.Expired:
				discord.AlertExpired(s, d.Guild, channelID, caller, prettyStr, optionDb.ChannelType)
				err = d.RemoveOptionByCode(oID, traderID)
				if err != nil {
					log.Println(err)
				}

			case background.New_High:
				err = d.OptionSetNewHigh(oID, traderID, msg.Price)
				if err != nil {
					log.Println(err)
				}

				if optionDb.ChannelType == discord.DAY {
					al := highest_tracker.GetAlertsGuild(optionDb.OptionGuildID, traderID)

					if msg.Price >= optionDb.OptionHighest {
						log.Println("Storing Option High for Tracker")
						err = al.WriteResult(highest_tracker.OPTION_TYPE, optionDb, msg.Price)
						if err != nil {
							log.Println(err)
						}
					}
				}

			case background.POI:
				discord.POI(s, d.Guild, channelID, prettyStr, caller, optionDb.OptionUnderlyingPoI, optionDb.ChannelType, ref)
				err = d.OptionPOIHit(oID, traderID)
				if err != nil {
					log.Println(err)
				}

			case background.SL:
				discord.StopLoss(s, d.Guild, channelID, prettyStr, caller, optionDb.OptionUnderlyingStop, optionDb.ChannelType, ref)
				err = d.RemoveOptionByCode(oID, traderID)
				if err != nil {
					log.Println(err)
				}

			case background.Price:
				discord.AlertOption(s, d.Guild, channelID, caller, prettyStr, msg.PctGain, msg.Price, optionDb.ChannelType, ref)

				if msg.PctGain > 50 {
					twtClient, err := twitter.GetTwitterClient(d.Guild)
					if err != nil {
						log.Println(fmt.Errorf("%v - %w", ticker, err))
					} else {
						go twtClient.SendDelayedTweet("Short", ticker, optionDb.OptionCallTime, msg.PctGain)
					}
				}
			}

		}
	}
}

func KeepTrack(s *discordgo.Session, guild, channelID, traderID, author string) {
	b := background.NewBG(guild, nil)

	responseChannel := make(chan background.Response)

	go b.KeepTrack(responseChannel, traderID)

	for {
		select {
		case msg := <-responseChannel:
			switch msg.Type {

			case background.Error:
				log.Println(msg.Message)
				return
			case background.EoD:
				discord.EOD(s, guild, channelID, traderID, author)
			}
		}
	}
}
