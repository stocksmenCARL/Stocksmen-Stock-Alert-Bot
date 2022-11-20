package discord

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/m1k8/kronos/pkg/M1K8/harpe/pkg/utils"
	"github.com/m1k8/kronos/pkg/highest_tracker"
)

// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
func SendError(s *discordgo.Session, channelID string, err error) {
	errStr := fmt.Errorf("**Error** : *%w*", err)
	s.ChannelMessageSend(channelID, errStr.Error())

	log.Println(errStr)
}

func SendReplyWithEmbed(sess *discordgo.Session, guildID, channelID string, msg *discordgo.MessageEmbed, ref *discordgo.MessageReference) error {
	var err error
	if s, ok := servers[guildID]; ok {
		for _, channel := range s.ChannelConfig {
			switch channelID {
			case channel.Swing:
				_, err = sendEmbedReply(sess, guildID, channelID, msg, ref)
			case channel.Day:
				_, err = sendEmbedReply(sess, guildID, channelID, msg, ref)
			case channel.Watchlist:
				_, err = sendEmbedReply(sess, guildID, channelID, msg, ref)
			case channel.Alerts:
				_, err = sendEmbedReply(sess, guildID, channelID, msg, ref)
			}
		}
	} else {
		return errors.New("server not configured")
	}
	return err
}

func SendAlertWithEmbed(sess *discordgo.Session, guildID, channelID string, msg *discordgo.MessageEmbed) error {
	var err error
	if s, ok := servers[guildID]; ok {
		for _, channel := range s.ChannelConfig {
			switch channelID {
			case channel.Swing:
				_, err = sendEmbed(sess, guildID, channelID, msg)
			case channel.Day:
				_, err = sendEmbed(sess, guildID, channelID, msg)
			case channel.Watchlist:
				_, err = sendEmbed(sess, guildID, channelID, msg)
			case channel.Alerts:
				_, err = sendEmbed(sess, guildID, channelID, msg)
			}
		}
	} else {
		return errors.New("server not configured")
	}
	return err
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func GetMessageEmbed(colour int, alerter string, msg ...string) *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Author:      &discordgo.MessageEmbedAuthor{},
		Color:       colour,
		Description: msg[0],
		Timestamp:   time.Now().Format(time.RFC3339),
		Footer: &discordgo.MessageEmbedFooter{
			Text:    "Built for Stocksmen Inc. | Managed by Stocksmen Carl | NOT FINANCIAL ADVICE",
			IconURL: "https://i.imgur.com/Uvknicu.png",
		},
	}
	if len(msg) > 1 {
		for i := 1; i < len(msg); i++ {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:   "Info",
				Value:  msg[i],
				Inline: false,
			})
		}
	}
	return embed
}

func GetOptionsEmbed(colour int, alerter, prettyStr, cType, ticker, strike, expiry string, cost, stop, poi, pt float32, under float64) *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Color:       colour,
		Title:       fmt.Sprintf("📜Options alert for: *%v*", prettyStr),
		Description: "Called by **" + alerter + "**",
		Timestamp:   time.Now().Format(time.RFC3339),
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Built for Stocksmen Inc. | Managed by Stocksmen Carl | NOT FINANCIAL ADVICE",
		},
	}

	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name:   "📜**Contract**",
		Value:  prettyStr,
		Inline: false,
	})

	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name:   "💰**Entry Price**",
		Value:  fmt.Sprintf("$%.2f", cost),
		Inline: false,
	})

	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name:   fmt.Sprintf("💵** %v Price**", ticker),
		Value:  fmt.Sprintf("$%.2f", under),
		Inline: false,
	})

	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name:   "📈**Underlying Chart**",
		Value:  utils.GetTDUrl(ticker),
		Inline: false,
	})

	if stop > 0 {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "❌**Stop Loss** *(Underlying)*",
			Value:  fmt.Sprintf("$%.2f", stop),
			Inline: false,
		})
	}

	if poi > 0 {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "👀**Point Of Interest** *(Underlying)*",
			Value:  fmt.Sprintf("$%.2f", poi),
			Inline: false,
		})
	}

	if pt > 0 {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "🔮**Price Target**",
			Value:  fmt.Sprintf("$%.2f", pt),
			Inline: false,
		})
	}

	return embed
}

func GetStockEmbed(colour int, alerter, ticker, expiry string, cost, stop, poi, spt, ept float32) *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Color:       colour,
		Title:       fmt.Sprintf("📈Stock alert for: *%v*", ticker),
		Description: "Called by **" + alerter + "**",
		Timestamp:   time.Now().Format(time.RFC3339),
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Built for Stocksmen Inc. | Managed by Stocksmen Carl | NOT FINANCIAL ADVICE",
		},
	}

	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name:   "💹**Stock**",
		Value:  ticker,
		Inline: false,
	})

	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name:   "💰**Entry Price**",
		Value:  fmt.Sprintf("$%.2f", cost),
		Inline: false,
	})

	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name:   "📈**Chart**",
		Value:  utils.GetTDUrl(ticker),
		Inline: false,
	})

	if stop > 0 {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "❌**Stop Loss**",
			Value:  fmt.Sprintf("$%.2f", stop),
			Inline: false,
		})
	}

	if poi > 0 {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "👀**Point Of Interest**",
			Value:  fmt.Sprintf("$%.2f", poi),
			Inline: false,
		})
	}

	if spt > 0 {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "🔮**Scale Price Target**",
			Value:  fmt.Sprintf("$%.2f", spt),
			Inline: false,
		})
	}

	if spt > 0 {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "➡️**Exit Price Target**",
			Value:  fmt.Sprintf("$%.2f", ept),
			Inline: false,
		})
	}
	return embed
}

func GetShortEmbed(colour int, alerter, ticker, expiry string, cost, stop, poi, spt, ept float32) *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Color:       colour,
		Title:       fmt.Sprintf("📉Short alert for: *%v*", ticker),
		Description: "Called by **" + alerter + "**",
		Timestamp:   time.Now().Format(time.RFC3339),
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Built for Stocksmen Inc. | Managed by Stocksmen Carl | NOT FINANCIAL ADVICE",
		},
	}

	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name:   "💹**Stock**",
		Value:  ticker,
		Inline: false,
	})

	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name:   "💰**Entry Price**",
		Value:  fmt.Sprintf("$%.2f", cost),
		Inline: false,
	})

	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name:   "📈**Chart**",
		Value:  utils.GetTDUrl(ticker),
		Inline: false,
	})
	if stop > 0 {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "❌**Stop Loss**",
			Value:  fmt.Sprintf("$%.2f", stop),
			Inline: false,
		})
	}

	if poi > 0 {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "👀**Point Of Interest**",
			Value:  fmt.Sprintf("$%.2f", poi),
			Inline: false,
		})
	}

	if spt > 0 {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "🔮**Scale Price Target**",
			Value:  fmt.Sprintf("$%.2f", spt),
			Inline: false,
		})
	}

	if spt > 0 {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "➡️**Exit Price Target**",
			Value:  fmt.Sprintf("$%.2f", ept),
			Inline: false,
		})
	}
	return embed
}

func GetCryptoEmbed(colour int, alerter, ticker, expiry string, cost, stop, poi, spt, ept float32) *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Color:       colour,
		Title:       fmt.Sprintf("🪙Crypto alert for: *%v*", ticker),
		Description: "Called by **" + alerter + "**",
		Timestamp:   time.Now().Format(time.RFC3339),
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Built for Stocksmen Inc. | Managed by Stocksmen Carl | NOT FINANCIAL ADVICE",
		},
	}

	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name:   "🪙**Coin**",
		Value:  ticker,
		Inline: false,
	})

	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name:   "💰**Entry Price**",
		Value:  fmt.Sprintf("$%.2f", cost),
		Inline: false,
	})

	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name:   "📈**Chart**",
		Value:  utils.GetTDCrypto(ticker),
		Inline: false,
	})

	if stop > 0 {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "❌**Stop Loss**",
			Value:  fmt.Sprintf("$%.2f", stop),
			Inline: false,
		})
	}

	if poi > 0 {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "👀**Point Of Interest**",
			Value:  fmt.Sprintf("$%.2f", poi),
			Inline: false,
		})
	}

	if spt > 0 {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "🔮**Scale Price Target**",
			Value:  fmt.Sprintf("$%.2f", spt),
			Inline: false,
		})
	}

	if spt > 0 {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "➡️**Exit Price Target**",
			Value:  fmt.Sprintf("$%.2f", ept),
			Inline: false,
		})
	}
	return embed
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func AlertExpired(sess *discordgo.Session, guildID, channelID, author, prettyStr string, channelType int) {
	msg := "Alert for ***" + prettyStr + "***" + " has expired."
	sendAlertNoPing(sess, guildID, channelID, msg, author, utils.Neutral, channelType)
}

func AlertStock(sess *discordgo.Session, guildID, channelID, author, ticker string, diff, value float32, channelType int, ref *discordgo.MessageReference) {
	m := "Stock Alert from " + author + " for **" +
		ticker + "** has hit **" + fmt.Sprintf("%.2f", diff) + "%**, price: **$**" +
		fmt.Sprintf("%.2f", value) + " !"

	sendReply(sess, guildID, channelID, m, author, utils.Hit, channelType, ref)
	//go func() {
	//	m := msg //capture the variable
	//	time.Sleep(time.Second * 30)
	//	sess.ChannelMessageDelete(m.ChannelID, m.ID)
	//}()
}

func AlertStockPT(sess *discordgo.Session, guildID, channelID, author, ticker string, pt float32, channelType int) {
	msg := "**" + ticker + "** has hit its **scale-out** PT of **$" + fmt.Sprintf("%.2f", pt) + "**"

	sendAlert(sess, guildID, channelID, msg, author, utils.STC, channelType)
}

func AlertStockPTExit(sess *discordgo.Session, guildID, channelID, author, ticker string, pt float32, channelType int) {
	msg := "**" + ticker + "** has hit its **exit** PT of **" + fmt.Sprintf("%.2f", pt) + "**"

	sendAlert(sess, guildID, channelID, msg, author, utils.STC, channelType)
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func AlertShort(sess *discordgo.Session, guildID, channelID, author, ticker string, diff, value float32, channelType int, ref *discordgo.MessageReference) {
	m := "Short Alert " + author + " for **" +
		ticker + "** has hit **" + fmt.Sprintf("%.2f", diff) + "%**, price: **$**" +
		fmt.Sprintf("%.2f", value) + " !"

	sendReply(sess, guildID, channelID, m, author, utils.Hit, channelType, ref)
	//go func() {
	//	m := msg //capture the variable
	//	time.Sleep(time.Second * 30)
	//	sess.ChannelMessageDelete(m.ChannelID, m.ID)
	//}()
}

func AlertShortPT(sess *discordgo.Session, guildID, channelID, author, ticker string, pt float32, channelType int) {
	msg := "** Short " + ticker + "** has hit its **scale-out** PT of **$" + fmt.Sprintf("%.2f", pt) + "**"

	sendAlert(sess, guildID, channelID, msg, author, utils.STC, channelType)
}

func AlertShortPTExit(sess *discordgo.Session, guildID, channelID, author, ticker string, pt float32, channelType int) {
	msg := "** Short " + ticker + "** has hit its **exit** PT of **" + fmt.Sprintf("%.2f", pt) + "**"

	sendAlert(sess, guildID, channelID, msg, author, utils.STC, channelType)
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func AlertCrypto(sess *discordgo.Session, guildID, channelID, author, ticker string, diff, value float32, channelType int, ref *discordgo.MessageReference) {
	m := "Crypto Alert " + author + " for **" +
		ticker + "** has hit **" + fmt.Sprintf("%.2f", diff) + "%**, price: **$**" +
		fmt.Sprintf("%.2f", value) + " !"

	sendReply(sess, guildID, channelID, m, author, utils.Hit, channelType, ref)
	//go func() {
	//	m := msg //capture the variable
	//	time.Sleep(time.Second * 30)
	//	sess.ChannelMessageDelete(m.ChannelID, m.ID)
	//}()
}

func AlertCryptoPT(sess *discordgo.Session, guildID, channelID, author, ticker string, pt float32, channelType int) {
	msg := "**" + ticker + "** has hit its **scale-out** PT of **$" + fmt.Sprintf("%.8f", pt) + "**"

	sendAlert(sess, guildID, channelID, msg, author, utils.STC, channelType)
}

func AlertCryptoPTExit(sess *discordgo.Session, guildID, channelID, author, ticker string, pt float32, channelType int) {
	msg := "**" + ticker + "** has hit its **exit** PT of **" + fmt.Sprintf("%.8f", pt) + "**"

	sendAlert(sess, guildID, channelID, msg, author, utils.STC, channelType)
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func AlertOption(sess *discordgo.Session, guildID, channelID, author, prettyStr string, diff, value float32, channelType int, ref *discordgo.MessageReference) {
	m := "Options Alert from " + author + " for **" +
		prettyStr + "** has hit **" + fmt.Sprintf("%.2f", diff) + "%**, price: **$**" +
		fmt.Sprintf("%.2f", value) + " !"

	sendReply(sess, guildID, channelID, m, author, utils.Hit, channelType, ref)
	//go func() {
	//	m := msg //capture the variable
	//	time.Sleep(time.Second * 30)
	//	sess.ChannelMessageDelete(m.ChannelID, m.ID)
	//}()
}

func AlertOptionPT(sess *discordgo.Session, guildID, channelID, author, ticker string, pt float32, channelType int) {
	msg := "**Option " + ticker + "** has hit its PT of **" + fmt.Sprintf("%.8f", pt)

	sendAlert(sess, guildID, channelID, msg, author, utils.STC, channelType)
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func StopLoss(sess *discordgo.Session, guildID, channelID, ticker, author string, stop float32, channelType int, ref *discordgo.MessageReference) {
	msg := fmt.Sprintf("**%v** has hit its stop loss of **$%f**. Alert removed.", ticker, stop)

	sendReply(sess, guildID, channelID, msg, author, utils.STC, channelType, ref)
}

func POI(sess *discordgo.Session, guildID, channelID, ticker, author string, poi float32, channelType int, ref *discordgo.MessageReference) {
	msg := fmt.Sprintf("**%v** is within 1%% of its point of interest: **$%.2f**", ticker, poi)

	sendReply(sess, guildID, channelID, msg, author, utils.Neutral, channelType, ref)
}

func EODEmbed(s *discordgo.Session, guildID, traderID string, author string) *discordgo.MessageEmbed {
	a := highest_tracker.GetAlertsGuild(guildID, traderID)

	stocks, shorts, crypto, options, total := a.GetAll()

	winners := a.GetWinners()
	cryptStr := ""
	shtStr := ""
	stStr := ""
	optStrOverflow := make([]string, 0, 0)
	tmpOptStrBuffer := ""

	if total == 0 {
		log.Println("TOTAL == 0")
		return nil
	}

	for k, v := range crypto {
		cryptStr += fmt.Sprintf("__%v__\n  \t• Starting: *$%.4f*\n\t• Highest: *$%.4f*\n\t• Pct Change: *%.2f%%*\n\n", k, v.Starting, v.Peak, v.PctDiff)
	}

	for k, v := range stocks {
		stStr += fmt.Sprintf("__%v__\n  \t• Starting: *$%.4f*\n\t• Highest: *$%.4f*\n\t• Pct Change: *%.2f%%*\n\n", k, v.Starting, v.Peak, v.PctDiff)
	}

	for k, v := range shorts {
		shtStr += fmt.Sprintf("__%v__\n  \t• Starting: *$%.4f*\n\t• Highest: *$%.4f*\n\t• Pct Change: *%.2f%%*\n\n", k, v.Starting, v.Peak, v.PctDiff)
	}

	for k, v := range options {
		tmpOptStrBuffer += fmt.Sprintf("__%v__\n  \t• Starting: *$%.4f*\n\t• Highest: *$%.4f*\n\t• Pct Change: *%.2f%%*\n\n", k, v.Starting, v.Peak, v.PctDiff)

		if len(tmpOptStrBuffer) >= 800 {
			optStrOverflow = append(optStrOverflow, tmpOptStrBuffer)
			tmpOptStrBuffer = fmt.Sprintf("__%v__\n  \t• Starting: *$%.4f*\n\t• Highest: *$%.4f*\n\t• Pct Change: *%.2f%%*\n\n", k, v.Starting, v.Peak, v.PctDiff)
		}
	}

	e := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{},
		Color:  0x0000ff,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "🥇Winners (>5% gain on alert close)",
				Value: fmt.Sprintf("**%.2f%%** *(Winners: %d, Total alerts: %d*)", (float64(winners) / float64(total) * 100), winners, total),
			},
		},
		Timestamp: time.Now().Format(time.RFC3339),
		Title:     "End of Day Report:",
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Built for Stocksmen Inc. | Managed by Stocksmen Carl | NOT FINANCIAL ADVICE",
		},
	}

	if len(crypto) > 0 {
		e.Fields = append(e.Fields, &discordgo.MessageEmbedField{
			Name:   "🪙Crypto",
			Value:  cryptStr,
			Inline: false,
		})
	}

	if len(stocks) > 0 {
		e.Fields = append(e.Fields, &discordgo.MessageEmbedField{
			Name:   "💹Stocks",
			Value:  stStr,
			Inline: false,
		})
	}

	if len(shorts) > 0 {
		e.Fields = append(e.Fields, &discordgo.MessageEmbedField{
			Name:   "🐻Shorts",
			Value:  shtStr,
			Inline: false,
		})
	}

	if len(options) > 0 {
		e.Fields = append(e.Fields, &discordgo.MessageEmbedField{
			Name:   "📜Options Pt 1",
			Value:  tmpOptStrBuffer,
			Inline: false,
		})
		for i, v := range optStrOverflow {
			e.Fields = append(e.Fields, &discordgo.MessageEmbedField{
				Name:   fmt.Sprintf("📜Options Pt %d", i+2),
				Value:  v,
				Inline: false,
			})
		}
	}
	e.Fields = append(e.Fields, &discordgo.MessageEmbedField{
		Name:   "Alerter",
		Value:  author,
		Inline: true,
	})

	return e
}

func EOD(s *discordgo.Session, guildID, channelID, traderID string, author string) {
	e := EODEmbed(s, guildID, traderID, author)
	if e == nil {
		return
	}
	ms := &discordgo.MessageSend{
		Embed:   e,
		Content: "@everyone",
	}
	var err error
	if _, ok := servers[guildID]; ok {
		log.Println("Sending tracker to " + channelID)
		_, err = s.ChannelMessageSendComplex(channelID, ms)
		if err != nil {
			log.Println(err)
		}
		return
	} else {
		err = errors.New("this guild is not whitelisted to use this bot. Contact @M1K_8 on Twitter to explain yourself")
		log.Println(err)
		return
	}
}

// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
func sendAlert(sess *discordgo.Session, guildID, channelID, msg, author string, colour, channelType int) error {
	var err error

	if s, ok := servers[guildID]; ok {
		for _, channel := range s.ChannelConfig {
			switch channelID {
			case channel.Swing:
				_, err = sendEmbed(sess, guildID, channelID, GetMessageEmbed(colour, author, msg))
			case channel.Day:
				_, err = sendEmbed(sess, guildID, channelID, GetMessageEmbed(colour, author, msg))
			case channel.Watchlist:
				_, err = sendEmbed(sess, guildID, channelID, GetMessageEmbed(colour, author, msg))
			}
		}
	} else {
		return errors.New("server not configured")
	}
	return err
}

func sendAlertNoPing(sess *discordgo.Session, guildID, channelID, msg, author string, colour, channelType int) error {
	var err error

	if s, ok := servers[guildID]; ok {
		for _, channel := range s.ChannelConfig {
			switch channelID {
			case channel.Swing:
				err = sendEmbedNoPing(sess, guildID, channelID, GetMessageEmbed(colour, author, msg))
			case channel.Day:
				err = sendEmbedNoPing(sess, guildID, channelID, GetMessageEmbed(colour, author, msg))
			case channel.Watchlist:
				err = sendEmbedNoPing(sess, guildID, channelID, GetMessageEmbed(colour, author, msg))
			case channel.Alerts:
				err = sendEmbedNoPing(sess, guildID, channelID, GetMessageEmbed(colour, author, msg))
			}
		}
	} else {
		return errors.New("server not configured")
	}
	return err
}

func sendReply(sess *discordgo.Session, guildID, channelID, m, author string, colour, channelType int, ref *discordgo.MessageReference) (msg *discordgo.Message, err error) {
	if s, ok := servers[guildID]; ok {
		for _, channel := range s.ChannelConfig {
			switch channelID {
			case channel.Swing:
				_, err = sendEmbedReply(sess, guildID, channelID, GetMessageEmbed(colour, author, m), ref)
			case channel.Day:
				_, err = sendEmbedReply(sess, guildID, channelID, GetMessageEmbed(colour, author, m), ref)
			case channel.Watchlist:
				_, err = sendEmbedReply(sess, guildID, channelID, GetMessageEmbed(colour, author, m), ref)
			case channel.Alerts:
				_, err = sendEmbedReply(sess, guildID, channelID, GetMessageEmbed(colour, author, m), ref)
			}
		}
	} else {
		err = errors.New("server not configured")
		return
	}
	return
}

func sendEmbed(s *discordgo.Session, thisGuild, channelToSendTo string, e *discordgo.MessageEmbed) (msg *discordgo.Message, err error) {
	ms := &discordgo.MessageSend{
		Embed:   e,
		Content: "@everyone",
	}

	if _, ok := servers[thisGuild]; ok {
		msg, err = s.ChannelMessageSendComplex(channelToSendTo, ms)
		if err != nil {
			log.Println(err.Error())
		}
		return
	} else {
		err = errors.New("this guild is not whitelisted to use this bot. Contact @M1K_8 on Twitter to explain yourself")
		log.Println(err)
		return
	}
}

func sendEmbedReply(s *discordgo.Session, thisGuild, channelToSendTo string, e *discordgo.MessageEmbed, ref *discordgo.MessageReference) (msg *discordgo.Message, err error) {

	if ref == nil {
		msg, err = sendEmbed(s, thisGuild, channelToSendTo, e)
		return
	}
	ms := &discordgo.MessageSend{
		Embed:     e,
		Content:   "@everyone",
		Reference: ref,
	}

	if _, ok := servers[thisGuild]; ok {
		msg, err = s.ChannelMessageSendComplex(channelToSendTo, ms)
	} else {
		err = errors.New("this guild is not whitelisted to use this bot. Contact @M1K_8 on Twitter to explain yourself")
		log.Println(err)
		return
	}
	return
}

func sendEmbedNoPing(s *discordgo.Session, thisGuild, channelToSendTo string, e *discordgo.MessageEmbed) (err error) {
	ms := &discordgo.MessageSend{
		Embed: e,
	}

	if _, ok := servers[thisGuild]; ok {
		_, err = s.ChannelMessageSendComplex(channelToSendTo, ms)
		return
	} else {
		err = errors.New("this guild is not whitelisted to use this bot. Contact @M1K_8 on Twitter to explain yourself")
		log.Println(err)
		return
	}
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
