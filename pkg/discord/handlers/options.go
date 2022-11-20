package handlers

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/m1k8/kronos/pkg/M1K8/harpe/pkg/db"
	"github.com/m1k8/kronos/pkg/M1K8/harpe/pkg/utils"
	"github.com/m1k8/kronos/pkg/M1K8/nabu/pkg/fetcher"
	"github.com/m1k8/kronos/pkg/discord"
	"github.com/m1k8/kronos/pkg/worker"
)

func destructOID(oID string) (ticker, cType, day, month, year string, price float32) {
	offset := 0
	ticker = oID[:1]

	for i := 1; rune(oID[i]) >= 65; i++ {
		ticker += oID[i : i+1]
		offset += 1
	}

	year = "20" + oID[1+offset:3+offset]
	month = oID[3+offset : 5+offset]
	day = oID[5+offset : 7+offset]
	cType = oID[7+offset : 8+offset]
	priceBy1k, _ := strconv.ParseFloat(oID[8+offset:], 32)
	price = float32(priceBy1k) / 1000
	return
}

func getOptionAndCheckOwner(d *db.DB, oID, traderID, caller string) (ticker, cType, day, month, year string, price float32, err error) {

	opt, err := d.GetOption(oID, traderID)

	if err != nil {
		return "", "", "", "", "", 0, err
	}

	trimmedoptCaller := opt.Caller[2 : len(opt.Caller)-1]
	if trimmedoptCaller != caller {
		fmt.Println(caller + " || " + opt.Caller)
		return "", "", "", "", "", 0, errors.New("callers dont match")
	}
	ticker, cType, day, month, year, price = destructOID(oID)
	return
}

func createOptionHandler(sess *discordgo.Session, d *db.DB, channelID, author, ticker, traderID, strike, year, month, day, entry, poi, stop, pt string, channelType int, ref *discordgo.MessageReference) (*discordgo.MessageEmbed, float32, error) {
	fetch := fetcher.NewFetcher()

	var (
		err             error
		price           string
		cType           string
		actual_starting float32
	)

	price = strike[:len(strike)-1]
	cType = strings.ToUpper(strike[len(strike)-1:])

	priceFl, err := strconv.ParseFloat(price, 32)

	if err != nil {
		return nil, 0, err
	}
	starting, err := fetch.GetStock(ticker)
	if err != nil {
		return nil, 0, err
	}

	if entry == "" {
		actual_starting, _, err = fetch.GetOption(ticker, cType, day, month, year, float32(priceFl), 0)
		if err != nil {
			return nil, 0, err
		}
	} else {
		log.Println("MANUAL ENTRY FOR " + ticker)
		pEntry, err := strconv.ParseFloat(entry, 32)
		entryFl := float32(pEntry)
		if err != nil {
			return nil, 0, err
		} else {
			actual_starting = float32(entryFl)
		}
	}

	if actual_starting == 0 || starting == -1 { // || starting.Results.LastQuote.Ask == 0 || starting.Results.LastQuote.Bid == 0 {
		return nil, 0, errors.New(utils.NiceStr(ticker, cType, day, month, year, float32(priceFl)) + " doesnt appear to be valid. Has it already expired?")
	}

	ptFl, err := strconv.ParseFloat(pt, 32)

	if err != nil {
		ptFl = 0
	}

	poiFl, err := strconv.ParseFloat(poi, 32)

	if err != nil {
		poiFl = 0
	}

	stopFl, err := strconv.ParseFloat(stop, 32)

	if err != nil {
		stopFl = 0
	}
	oID := utils.GetCode(ticker, cType, day, month, year, float32(priceFl))
	prettyStr := utils.NiceStr(ticker, cType, day, month, year, float32(priceFl))

	exitChan, _, exists, err := d.CreateOption(oID, traderID, author, channelType, ticker, cType, day, month, year, float32(priceFl), actual_starting, float32(ptFl), float32(poiFl), float32(stopFl), starting)

	if err != nil {
		return nil, 0, fmt.Errorf("couldnt create Option alert %v: %w", utils.NiceStr(ticker, cType, day, month, year, float32(priceFl)), err)
	}

	if !exists {
		go worker.OptionsTracker(sess, channelID, author, ticker, traderID, cType, day, month, year, float32(priceFl), d, ref, exitChan)
	} else {
		option, err := d.GetOption(oID, traderID)
		if err != nil {
			return nil, 0, err
		}
		rmOptionHandler(option.OptionGuildID, channelID, author, oID, traderID, prettyStr, "", channelType)
		actual_starting = (actual_starting + option.OptionStarting) / 2
		err = d.RemoveOption(oID, traderID, cType, day, month, year, float32(priceFl))
		if err != nil {
			return nil, 0, err
		}
		exitChan, _, _, err = d.CreateOption(oID, traderID, author, channelType, ticker, cType, day, month, year, float32(priceFl), actual_starting, float32(ptFl), float32(poiFl), float32(stopFl), actual_starting)
		if err != nil {
			return nil, 0, err
		}
		go worker.OptionsTracker(sess, channelID, author, ticker, traderID, cType, day, month, year, float32(priceFl), d, ref, exitChan)
	}

	embed := discord.GetOptionsEmbed(utils.BTO, author, prettyStr, cType, ticker, string(strike[:len(strike)-1]), fmt.Sprintf("%s-%s-%s", year, month, day),
		actual_starting, float32(stopFl), float32(poiFl), float32(ptFl), float64(starting))

	log.Println(author + " options alert: " + utils.NiceStr(ticker, cType, day, month, year, float32(priceFl)))
	return embed, actual_starting, nil
}

func rmOptionHandler(guild, channelID, author, oID, traderID, prettyStr, desc string, channelType int) (*discordgo.MessageEmbed, error) {
	d := db.NewDB(guild)

	err := d.RemoveOptionByCode(oID, traderID)

	if err != nil {
		return nil, err
	}

	if err != nil {
		log.Println(fmt.Errorf("Option remove error %w", err))
		return nil, err
	} else {
		respStr := "Ticker ***" + prettyStr + "***  has been removed"

		var embed *discordgo.MessageEmbed
		if desc != "" {
			embed = discord.GetMessageEmbed(utils.STC, author, respStr, desc)
		} else {
			embed = discord.GetMessageEmbed(utils.STC, author, respStr)
		}
		log.Println("Removed " + prettyStr)
		return embed, nil
	}
}
