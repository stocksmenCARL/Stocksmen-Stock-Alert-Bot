package handlers

import (
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/m1k8/kronos/pkg/M1K8/harpe/pkg/db"
	"github.com/m1k8/kronos/pkg/M1K8/harpe/pkg/utils"
	"github.com/m1k8/kronos/pkg/M1K8/nabu/pkg/fetcher"
	"github.com/m1k8/kronos/pkg/discord"
	"github.com/m1k8/kronos/pkg/worker"
)

func getStockAndCheckOwner(d *db.DB, ticker, traderID, caller string) (err error) {

	stock, err := d.GetStock(ticker, traderID)

	if err != nil {
		return err
	}

	trimmedstockCaller := stock.Caller[2 : len(stock.Caller)-1]
	if trimmedstockCaller != caller {
		fmt.Println(caller + " || " + stock.Caller)
		return errors.New("callers dont match")
	}

	return

}

func createStockHandler(sess *discordgo.Session, d *db.DB, channelID, author, ticker, traderID, spt, ept, expiry, poi, entry, stop string, channelType int, ref *discordgo.MessageReference) (*discordgo.MessageEmbed, float32, error) {
	fetch := fetcher.NewFetcher()
	initPrice, err := fetch.GetStock(ticker)

	if err != nil {
		discord.SendError(sess, channelID, fmt.Errorf("couldnt get price for %v: %w", ticker, err))
		return nil, 0, err
	}

	expiryI, err := strconv.ParseInt(expiry, 10, 64)
	if err != nil {
		expiryI = 0
	}
	sptFl, pErr := strconv.ParseFloat(spt, 32)

	if pErr != nil {
		sptFl = 0
	}

	eptFl, pErr := strconv.ParseFloat(ept, 32)

	if pErr != nil {
		eptFl = 0
	}
	entryFl, pErr := strconv.ParseFloat(entry, 32)

	if pErr != nil {
		entryFl = 0
	} else {
		initPrice = float32(entryFl)
	}

	stopFl, pErr := strconv.ParseFloat(stop, 32)

	if pErr != nil {
		stopFl = 0
	}

	poiFl, pErr := strconv.ParseFloat(poi, 32)

	if pErr != nil {
		poiFl = 0
	}
	var starting float32

	if float32(entryFl) == 0 {
		starting, err = fetch.GetStock(ticker)
		if err != nil {
			return nil, 0, err
		}
	} else {
		starting = float32(entryFl)
	}

	exitChan, exists, err := d.CreateStock(ticker, traderID, author, channelType, float32(sptFl), float32(eptFl), float32(poiFl), float32(stopFl), expiryI, starting)

	if err != nil {
		return nil, 0, err
	}

	if !exists {
		go worker.StockTracker(sess, channelID, ticker, traderID, author, expiry, d, ref, exitChan)
	} else {
		stock, err := d.GetStock(ticker, traderID)
		if err != nil {
			return nil, 0, err
		}
		rmStockHandler(d.Guild, channelID, author, ticker, traderID, "", channelType)
		starting = (starting + stock.StockStarting) / 2
		err = d.RemoveStock(ticker, traderID)
		if err != nil {
			return nil, 0, err
		}
		exitChan, _, err = d.CreateStock(ticker, traderID, author, channelType, float32(sptFl), float32(eptFl), float32(poiFl), float32(stopFl), expiryI, starting)
		if err != nil {
			return nil, 0, err
		}
		go worker.StockTracker(sess, channelID, ticker, traderID, author, expiry, d, ref, exitChan)
	}

	if err != nil {
		return nil, 0, err
	}

	embed := discord.GetStockEmbed(utils.BTO, author, ticker, expiry, initPrice, float32(stopFl), float32(poiFl), float32(sptFl), float32(eptFl))

	return embed, initPrice, nil
}

func rmStockHandler(guild, channelID, author, ticker, traderID, desc string, channelType int) (*discordgo.MessageEmbed, error) {
	d := db.NewDB(guild)

	err := d.RemoveStock(ticker, traderID)

	if err != nil {
		return nil, err
	}

	if err != nil {
		log.Println(fmt.Errorf("Stock remove error %w", err))
		return nil, err
	} else {
		respStr := "Ticker ***" + ticker + "***  has been removed"

		var embed *discordgo.MessageEmbed
		if desc != "" {
			embed = discord.GetMessageEmbed(utils.STC, author, respStr, desc)
		} else {
			embed = discord.GetMessageEmbed(utils.STC, author, respStr)
		}
		log.Println("Removed " + ticker)
		return embed, nil
	}
}
