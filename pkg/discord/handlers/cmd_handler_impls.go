package handlers

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	graph "github.com/m1k8/kronos/pkg/M1K8/Pazuzu/pkg/graph"

	"github.com/m1k8/kronos/pkg/M1K8/harpe/pkg/db"
	"github.com/m1k8/kronos/pkg/M1K8/harpe/pkg/utils"
	stonks "github.com/m1k8/kronos/pkg/M1K8/nabu/pkg/fetcher"
	"github.com/m1k8/kronos/pkg/discord"
)

type ComponenetMapData struct {
	Msg      *discordgo.Message
	TraderID string
	UserID   string
}

var componentMsgMap = make(map[string]ComponenetMapData)

var CommandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	"alert": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if !isAuthed(i) {
			log.Println("no perms for this action " + i.Interaction.Member.User.Username)
			return
		}
		embed := getAlertEmbed(s, i.Interaction.ChannelID, i.Member.User.Mention(), i.ApplicationCommandData().Options...)
		log.Println("Calling alert...")
		if embed == nil {
			discord.SendError(s, i.Interaction.ChannelID, errors.New("unable to create embed"))
			return
		}
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: i.ApplicationCommandData().Options[0].StringValue() + " @everyone",
				AllowedMentions: &discordgo.MessageAllowedMentions{
					Parse: []discordgo.AllowedMentionType{discordgo.AllowedMentionTypeEveryone},
				},
				Embeds: []*discordgo.MessageEmbed{embed},
			},
		})

		msg, _ := s.InteractionResponseEdit(s.State.User.ID, i.Interaction, &discordgo.WebhookEdit{
			Content: "New General Alert! @everyone",
			AllowedMentions: &discordgo.MessageAllowedMentions{
				Parse: []discordgo.AllowedMentionType{discordgo.AllowedMentionTypeEveryone},
			},
			Embeds: []*discordgo.MessageEmbed{embed},
		},
		)

		var ms *discordgo.MessageSend

		if msg == nil {
			ms = &discordgo.MessageSend{
				Content: "@everyone",
				AllowedMentions: &discordgo.MessageAllowedMentions{
					Parse: []discordgo.AllowedMentionType{discordgo.AllowedMentionTypeEveryone},
				},
			}
		} else {
			ms = &discordgo.MessageSend{
				Content: "@everyone",
				AllowedMentions: &discordgo.MessageAllowedMentions{
					Parse: []discordgo.AllowedMentionType{discordgo.AllowedMentionTypeEveryone},
				},
				Reference: msg.Reference(),
			}
		}

		msg, _ = s.ChannelMessageSendComplex(i.Interaction.ChannelID, ms)

		go func() {
			time.Sleep(time.Second * 10)
			s.ChannelMessageDelete(i.Interaction.ChannelID, msg.ID)
		}()
	},
	"tracker": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if !isAuthed(i) {
			log.Println("no perms for this action " + i.Interaction.Member.User.Username)
			return
		}
		traderID, err := discord.GetTraderID(i.Interaction.GuildID, i.Interaction.ChannelID)
		if err != nil {
			log.Println(err)
			return
		}
		author, err := discord.GetAuthorID(i.Interaction.GuildID, i.Interaction.ChannelID)
		if err != nil {
			log.Println(err)
			return
		}
		log.Println("Calling tracker...")
		e := discord.EODEmbed(s, i.Interaction.GuildID, traderID, author)

		if e == nil {
			log.Println("Tracker nil")
			s.InteractionResponseDelete(s.State.User.ID, i.Interaction)
			return
		}

		s.InteractionResponseDelete(s.State.User.ID, i.Interaction)
		discord.EOD(s, i.Interaction.GuildID, i.ChannelID, traderID, author)
	},
	"nuke": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if !isAuthed(i) {
			log.Println("no perms for this action " + i.Interaction.Member.User.Username)
			return
		}
		d := db.NewDB(i.Interaction.GuildID)
		traderID, err := discord.GetTraderID(i.Interaction.GuildID, i.Interaction.ChannelID)
		if err != nil {
			log.Println(err)
			return
		}
		log.Println("Calling nuke...")
		err = d.RmAllTrader(traderID)

		if err != nil {
			discord.SendError(s, i.Interaction.ChannelID, err)
			return
		}
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "***50,000 people used to live here...***",
			},
		})
	},
	"all": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "",
			},
		})
		log.Println("Calling all...")
		respStrs := all(s, i.Interaction.ChannelID, i.Interaction.GuildID)

		if len(respStrs) > 0 {
			s.InteractionResponseEdit(s.State.User.ID, i.Interaction, &discordgo.WebhookEdit{
				Content: "See report below:",
			},
			)
		} else {
			s.InteractionResponseDelete(s.State.User.ID, i.Interaction)
		}

		for _, str := range respStrs {
			s.ChannelMessageSend(i.Interaction.ChannelID, str)
		}
	},
	"refresh": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if !isAuthed(i) {
			log.Println("no perms for this action " + i.Interaction.Member.User.Username)
			return
		}

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "",
			},
		})

		respStr := refresh(s, i.Interaction.ChannelID, i.Interaction.GuildID)
		log.Println("Calling refresh...")

		s.InteractionResponseEdit(s.State.User.ID, i.Interaction, &discordgo.WebhookEdit{
			Content: "***Refreshed the following from the Database...***\n" + respStr,
		},
		)
	},
	"s": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if !isAuthed(i) {
			log.Println("no perms for this action " + i.Interaction.Member.User.Username)
			return
		}
		log.Println("Calling stock...")
		var (
			ticker string
			s_pt   string
			e_pt   string
			desc   string
			poi    string
			expiry string
			stop   string
			entry  string
		)

		argValMap := make(map[string]string)

		for _, v := range i.ApplicationCommandData().Options {
			argValMap[v.Name] = v.StringValue()
		}

		if v, ok := argValMap[TICKER]; !ok {
			discord.SendError(s, i.Interaction.ChannelID, errors.New("ticker not defined"))
			return
		} else {
			ticker = strings.ToUpper(v)
		}
		log.Println("Calling stock for " + ticker)

		if v, ok := argValMap[SCALE_PT]; ok {
			s_pt = v
		}
		if v, ok := argValMap[EXIT_PT]; ok {
			e_pt = v
		}
		if v, ok := argValMap[DESC]; ok {
			desc = v
		}
		if v, ok := argValMap[POI]; ok {
			poi = v
		}
		if v, ok := argValMap[EXPIRY]; ok {
			expiry = v
		}
		if v, ok := argValMap[STOPLOSS]; ok {
			stop = v
		}
		if v, ok := argValMap[ENTRY]; ok {
			entry = v
		}
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "",
			},
		})

		traderID, err := discord.GetTraderID(i.Interaction.GuildID, i.Interaction.ChannelID)
		if err != nil {
			log.Println(err)
			return
		}

		channelType := discord.GetChannelType(i.Interaction.GuildID, i.Interaction.ChannelID)

		if channelType == -1 {
			log.Println(errors.New("invalid channel type"))
			s.InteractionResponseDelete(s.State.User.ID, i.Interaction)
			return
		}

		msg, _ := s.InteractionResponseEdit(s.State.User.ID, i.Interaction, &discordgo.WebhookEdit{Content: "Thinking..."})
		var embed *discordgo.MessageEmbed
		var starting float32
		d := db.NewDB(i.Interaction.GuildID)
		if msg == nil {
			embed, starting, err = createStockHandler(s, d, i.Interaction.ChannelID, i.Member.User.Mention(), ticker, traderID, s_pt, e_pt, expiry, entry, poi, stop, channelType, nil)
		} else {
			embed, starting, err = createStockHandler(s, d, i.Interaction.ChannelID, i.Member.User.Mention(), ticker, traderID, s_pt, e_pt, expiry, entry, poi, stop, channelType, msg.Reference())
		}

		if err != nil {
			log.Println(fmt.Errorf("error creating createStockHandler: %w", err))
			s.InteractionResponseDelete(s.State.User.ID, i.Interaction)
			return
		}

		msg, _ = s.InteractionResponseEdit(s.State.User.ID, i.Interaction, &discordgo.WebhookEdit{
			Content: fmt.Sprintf("Stock %v @ $%.2f \n***%v***", ticker, starting, desc),
			AllowedMentions: &discordgo.MessageAllowedMentions{
				Parse: []discordgo.AllowedMentionType{discordgo.AllowedMentionTypeEveryone},
			},
			Embeds: []*discordgo.MessageEmbed{embed},
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.SelectMenu{
							CustomID:    "opt_stock",
							MinValues:   0,
							MaxValues:   1,
							Placeholder: "Actions",
							Options: []discordgo.SelectMenuOption{
								{
									Label:       "Remove Alert",
									Description: "Removes this alert. Only the alerter can do this - will fail for anyone else.",
									Default:     false,
									Value:       "rm_" + ticker,
									Emoji: discordgo.ComponentEmoji{
										Name: "❌",
									},
								},
								{
									Label:       "Get Email alerts for this (this does nothing atm)",
									Description: "Recieve Email alerts.",
									Default:     false,
									Value:       "em_" + ticker,
									Emoji: discordgo.ComponentEmoji{
										Name: "📧",
									},
								},
							},
						},
					},
				},
			},
		},
		)
		componentMsgMap[fmt.Sprintf("s_%s_%s", ticker, traderID)] = ComponenetMapData{msg, traderID, i.Interaction.Member.User.ID}
		var ms *discordgo.MessageSend

		includeChart := true

		chart, err := graph.Get15MStocksChart(ticker)

		if err != nil {
			includeChart = false
		}

		file, err := os.Open(chart)
		if err != nil {
			includeChart = false
		}

		defer func() {
			file.Close()

			os.Remove(chart)
		}()

		if msg == nil {
			ms = &discordgo.MessageSend{
				Content: "BTO " + ticker + " " + desc + " @everyone",
				AllowedMentions: &discordgo.MessageAllowedMentions{
					Parse: []discordgo.AllowedMentionType{discordgo.AllowedMentionTypeEveryone},
				},
			}

		} else {
			ms = &discordgo.MessageSend{
				Content: "BTO " + ticker + " " + desc + " @everyone",
				AllowedMentions: &discordgo.MessageAllowedMentions{
					Parse: []discordgo.AllowedMentionType{discordgo.AllowedMentionTypeEveryone},
				},
				Reference: msg.Reference(),
			}
		}

		if includeChart {
			ms.Files = []*discordgo.File{
				{
					Name:        chart,
					ContentType: "image/png",
					Reader:      file,
				},
			}
			ms.Embeds = []*discordgo.MessageEmbed{
				{
					Image: &discordgo.MessageEmbedImage{
						URL: "attachment://" + chart,
					},
				},
			}
		}

		msg, _ = s.ChannelMessageSendComplex(i.Interaction.ChannelID, ms)

		go func() {
			time.Sleep(time.Second * 30)
			s.ChannelMessageDelete(i.Interaction.ChannelID, msg.ID)
		}()
	},
	"rms": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if !isAuthed(i) {
			log.Println("no perms for this action " + i.Interaction.Member.User.Username)
			return
		}
		log.Println("Calling rm-stock...")
		var (
			ticker string
			desc   string
		)
		log.Println("Calling rm-stock for " + ticker)

		argValMap := make(map[string]string)

		for _, v := range i.ApplicationCommandData().Options {
			argValMap[v.Name] = v.StringValue()
		}

		if v, ok := argValMap[TICKER]; !ok {
			discord.SendError(s, i.Interaction.ChannelID, errors.New("ticker not defined"))
			return
		} else {
			ticker = strings.ToUpper(v)
		}

		if v, ok := argValMap[DESC]; ok {
			desc = v
		}

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{},
		})

		traderID, err := discord.GetTraderID(i.Interaction.GuildID, i.Interaction.ChannelID)
		if err != nil {
			log.Println(err)
			return
		}

		channelType := discord.GetChannelType(i.Interaction.GuildID, i.Interaction.ChannelID)

		if channelType == -1 {
			log.Println(errors.New("invalid channel type"))
			s.InteractionResponseDelete(s.State.User.ID, i.Interaction)
			return
		}
		var embed *discordgo.MessageEmbed

		embed, err = rmStockHandler(i.Interaction.GuildID, i.Interaction.ChannelID, i.Member.User.Mention(), ticker, traderID, desc, channelType)

		if err != nil {
			log.Println(fmt.Errorf("error creating rmStockHandler: %w", err))
			s.InteractionResponseDelete(s.State.User.ID, i.Interaction)
			return
		}

		fetcher := stonks.NewFetcher()
		closing, _ := fetcher.GetStock(ticker)

		s.InteractionResponseEdit(s.State.User.ID, i.Interaction, &discordgo.WebhookEdit{
			Content: fmt.Sprintf("Stock alert removed - **%v** @ $%.2f @everyone", ticker, closing),
			AllowedMentions: &discordgo.MessageAllowedMentions{
				Parse: []discordgo.AllowedMentionType{discordgo.AllowedMentionTypeEveryone},
			},
			Embeds: []*discordgo.MessageEmbed{embed},
		},
		)
		var msg *discordgo.Message
		if msg, ok := componentMsgMap[fmt.Sprintf("s_%s_%s", ticker, traderID)]; ok {
			str := fmt.Sprintf("This alert has been deleted - closed @ $%.2f", closing)
			s.ChannelMessageEditComplex(&discordgo.MessageEdit{
				Content:    &str,
				Components: []discordgo.MessageComponent{},
				ID:         msg.Msg.ID,
				Channel:    msg.Msg.ChannelID,
			})

			delete(componentMsgMap, fmt.Sprintf("s_%s_%s", ticker, traderID))
		}

		var ms *discordgo.MessageSend

		if msg == nil {
			ms = &discordgo.MessageSend{
				Content: "@everyone",
				AllowedMentions: &discordgo.MessageAllowedMentions{
					Parse: []discordgo.AllowedMentionType{discordgo.AllowedMentionTypeEveryone},
				},
			}
		} else {
			ms = &discordgo.MessageSend{
				Content: "@everyone",
				AllowedMentions: &discordgo.MessageAllowedMentions{
					Parse: []discordgo.AllowedMentionType{discordgo.AllowedMentionTypeEveryone},
				},
				Reference: msg.Reference(),
			}
		}

		msg, _ = s.ChannelMessageSendComplex(i.Interaction.ChannelID, ms)
		go func() {
			time.Sleep(time.Second * 10)
			s.ChannelMessageDelete(i.Interaction.ChannelID, msg.ID)
		}()
	},
	"sh": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if !isAuthed(i) {
			log.Println("no perms for this action " + i.Interaction.Member.User.Username)
			return
		}
		log.Println("Calling short...")
		var (
			ticker string
			spt    string
			ept    string
			desc   string
			poi    string
			expiry string
			stop   string
			entry  string
		)
		argValMap := make(map[string]string)

		for _, v := range i.ApplicationCommandData().Options {
			argValMap[v.Name] = v.StringValue()
		}

		if v, ok := argValMap[TICKER]; !ok {
			discord.SendError(s, i.Interaction.ChannelID, errors.New("ticker not defined"))
			return
		} else {
			ticker = strings.ToUpper(v)
		}
		log.Println("Calling short for " + ticker)

		if v, ok := argValMap[SCALE_PT]; ok {
			spt = v
		}
		if v, ok := argValMap[EXIT_PT]; ok {
			ept = v
		}

		if v, ok := argValMap[DESC]; ok {
			desc = v
		}

		if v, ok := argValMap[STOPLOSS]; ok {
			stop = v
		}

		if v, ok := argValMap[POI]; ok {
			poi = v
		}

		if v, ok := argValMap[EXPIRY]; ok {
			expiry = v
		}

		if v, ok := argValMap[ENTRY]; ok {
			entry = v
		}

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{},
		})

		traderID, err := discord.GetTraderID(i.Interaction.GuildID, i.Interaction.ChannelID)
		if err != nil {
			log.Println(err)
			return
		}

		channelType := discord.GetChannelType(i.Interaction.GuildID, i.Interaction.ChannelID)

		if channelType == -1 {
			log.Println(errors.New("invalid channel type"))
			s.InteractionResponseDelete(s.State.User.ID, i.Interaction)
			return
		}
		msg, _ := s.InteractionResponseEdit(s.State.User.ID, i.Interaction, &discordgo.WebhookEdit{Content: "Thinking..."})
		var embed *discordgo.MessageEmbed
		var starting float32

		d := db.NewDB(i.Interaction.GuildID)
		if msg == nil {
			embed, starting, err = createShortHandler(s, d, i.Interaction.ChannelID, i.Member.User.Mention(), ticker, traderID, spt, ept, expiry, entry, poi, stop, channelType, nil)
		} else {
			embed, starting, err = createShortHandler(s, d, i.Interaction.ChannelID, i.Member.User.Mention(), ticker, traderID, spt, ept, expiry, entry, poi, stop, channelType, msg.Reference())
		}

		if err != nil {
			log.Println(fmt.Errorf("error creating createShortHandler: %w", err))
			s.InteractionResponseDelete(s.State.User.ID, i.Interaction)
			return
		}

		msg, _ = s.InteractionResponseEdit(s.State.User.ID, i.Interaction, &discordgo.WebhookEdit{
			Content: fmt.Sprintf("Short %v @ $%.2f \n***%v***", ticker, starting, desc),
			AllowedMentions: &discordgo.MessageAllowedMentions{
				Parse: []discordgo.AllowedMentionType{discordgo.AllowedMentionTypeEveryone},
			},
			Embeds: []*discordgo.MessageEmbed{embed},
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.SelectMenu{
							CustomID:    "opt_short",
							MinValues:   0,
							MaxValues:   1,
							Placeholder: "Actions",
							Options: []discordgo.SelectMenuOption{
								{
									Label:       "Remove Alert",
									Description: "Removes this alert. Only the alerter can do this - will fail for anyone else.",
									Default:     false,
									Value:       "rm_" + ticker,
									Emoji: discordgo.ComponentEmoji{
										Name: "❌",
									},
								},
								{
									Label:       "Get Email alerts for this (this does nothing atm)",
									Description: "Recieve Email alerts.",
									Default:     false,
									Value:       "em_" + ticker,
									Emoji: discordgo.ComponentEmoji{
										Name: "📧",
									},
								},
							},
						},
					},
				},
			},
		},
		)
		componentMsgMap[fmt.Sprintf("sh_%s_%s", ticker, traderID)] = ComponenetMapData{msg, traderID, i.Interaction.Member.User.ID}
		var ms *discordgo.MessageSend

		if msg == nil {
			ms = &discordgo.MessageSend{
				Content: "Short " + ticker + " @everyone",
				AllowedMentions: &discordgo.MessageAllowedMentions{
					Parse: []discordgo.AllowedMentionType{discordgo.AllowedMentionTypeEveryone},
				},
			}
		} else {
			ms = &discordgo.MessageSend{
				Content: "Short " + ticker + " @everyone",
				AllowedMentions: &discordgo.MessageAllowedMentions{
					Parse: []discordgo.AllowedMentionType{discordgo.AllowedMentionTypeEveryone},
				},
				Reference: msg.Reference(),
			}
		}
		msg, _ = s.ChannelMessageSendComplex(i.Interaction.ChannelID, ms)
		go func() {
			time.Sleep(time.Second * 7)
			s.ChannelMessageDelete(i.Interaction.ChannelID, msg.ID)
		}()
	},
	"rmsh": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if !isAuthed(i) {
			log.Println("no perms for this action " + i.Interaction.Member.User.Username)
			return
		}
		log.Println("Calling rm-short...")
		var (
			ticker string
			desc   string
		)
		log.Println("Calling rm-short for " + ticker)
		argValMap := make(map[string]string)

		for _, v := range i.ApplicationCommandData().Options {
			argValMap[v.Name] = v.StringValue()
		}

		if v, ok := argValMap[TICKER]; !ok {
			discord.SendError(s, i.Interaction.ChannelID, errors.New("ticker not defined"))
			return
		} else {
			ticker = strings.ToUpper(v)
		}

		if v, ok := argValMap[DESC]; ok {
			desc = v
		}

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{},
		})

		traderID, err := discord.GetTraderID(i.Interaction.GuildID, i.Interaction.ChannelID)
		if err != nil {
			log.Println(err)
			return
		}

		channelType := discord.GetChannelType(i.Interaction.GuildID, i.Interaction.ChannelID)

		if channelType == -1 {
			log.Println(errors.New("invalid channel type"))
			s.InteractionResponseDelete(s.State.User.ID, i.Interaction)
			return
		}

		embed, err := rmShortHandler(i.Interaction.GuildID, i.Interaction.ChannelID, i.Member.User.Mention(), ticker, traderID, desc, channelType)

		if err != nil {
			log.Println(fmt.Errorf("error removing short: %w", err))
			s.InteractionResponseDelete(s.State.User.ID, i.Interaction)
			return
		}
		fetcher := stonks.NewFetcher()
		closing, _ := fetcher.GetStock(ticker)

		s.InteractionResponseEdit(s.State.User.ID, i.Interaction, &discordgo.WebhookEdit{
			Content: fmt.Sprintf("Short alert removed - **%v** @ $%.2f @everyone", ticker, closing),
			AllowedMentions: &discordgo.MessageAllowedMentions{
				Parse: []discordgo.AllowedMentionType{discordgo.AllowedMentionTypeEveryone},
			},
			Embeds: []*discordgo.MessageEmbed{embed},
		},
		)
		var msg *discordgo.Message
		if msg, ok := componentMsgMap[fmt.Sprintf("sh_%s_%s", ticker, traderID)]; ok {
			str := fmt.Sprintf("This alert has been deleted - closed @ $%.2f", closing)
			s.ChannelMessageEditComplex(&discordgo.MessageEdit{
				Content:    &str,
				Components: []discordgo.MessageComponent{},
				ID:         msg.Msg.ID,
				Channel:    msg.Msg.ChannelID,
			})

			delete(componentMsgMap, fmt.Sprintf("sh_%s_%s", ticker, traderID))
		}

		var ms *discordgo.MessageSend

		if msg == nil {
			ms = &discordgo.MessageSend{
				Content: "@everyone",
				AllowedMentions: &discordgo.MessageAllowedMentions{
					Parse: []discordgo.AllowedMentionType{discordgo.AllowedMentionTypeEveryone},
				},
			}
		} else {
			ms = &discordgo.MessageSend{
				Content: "@everyone",
				AllowedMentions: &discordgo.MessageAllowedMentions{
					Parse: []discordgo.AllowedMentionType{discordgo.AllowedMentionTypeEveryone},
				},
				Reference: msg.Reference(),
			}
		}

		msg, _ = s.ChannelMessageSendComplex(i.Interaction.ChannelID, ms)
		go func() {
			time.Sleep(time.Second * 10)
			s.ChannelMessageDelete(i.Interaction.ChannelID, msg.ID)
		}()
	},
	"c": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if !isAuthed(i) {
			log.Println("no perms for this action " + i.Interaction.Member.User.Username)
			return
		}
		log.Println("Calling crypto...")
		var (
			ticker string
			ept    string
			spt    string
			desc   string
			poi    string
			stop   string
			expiry string
			entry  string
		)
		argValMap := make(map[string]string)

		for _, v := range i.ApplicationCommandData().Options {
			argValMap[v.Name] = v.StringValue()
		}

		if v, ok := argValMap[COIN]; !ok {
			log.Println(errors.New("coin not defined"))
			return
		} else {
			ticker = strings.ToUpper(v)
		}
		log.Println("Calling crypto for " + ticker)

		if v, ok := argValMap[SCALE_PT]; ok {
			spt = v
		}
		if v, ok := argValMap[EXIT_PT]; ok {
			ept = v
		}

		if v, ok := argValMap[DESC]; ok {
			desc = v
		}

		if v, ok := argValMap[POI]; ok {
			poi = v
		}

		if v, ok := argValMap[STOPLOSS]; ok {
			stop = v
		}

		if v, ok := argValMap[EXPIRY]; ok {
			expiry = v
		}

		if v, ok := argValMap[ENTRY]; ok {
			entry = v
		}

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{},
		})

		traderID, err := discord.GetTraderID(i.Interaction.GuildID, i.Interaction.ChannelID)
		if err != nil {
			log.Println(err)
			return
		}

		channelType := discord.GetChannelType(i.Interaction.GuildID, i.Interaction.ChannelID)
		if channelType == -1 {
			log.Println(errors.New("invalid channel type"))
			s.InteractionResponseDelete(s.State.User.ID, i.Interaction)
			return
		}

		msg, err := s.InteractionResponseEdit(s.State.User.ID, i.Interaction, &discordgo.WebhookEdit{Content: "Thinking..."})

		if err != nil {
			log.Println(err)
			os.Exit(0)
		}
		var embed *discordgo.MessageEmbed
		var starting float32
		d := db.NewDB(i.Interaction.GuildID)

		if msg == nil {
			embed, starting, err = createCryptoHandler(s, d, i.Interaction.ChannelID, i.Member.User.Mention(), ticker, traderID, spt, ept, expiry, poi, entry, stop, channelType, nil)
		} else {
			embed, starting, err = createCryptoHandler(s, d, i.Interaction.ChannelID, i.Member.User.Mention(), ticker, traderID, spt, ept, expiry, poi, entry, stop, channelType, msg.Reference())
		}

		if err != nil {
			log.Println(fmt.Errorf("error creating createCryptoHandler: %w", err))
			s.InteractionResponseDelete(s.State.User.ID, i.Interaction)
			return
		}

		if err != nil {
			log.Println(err)
			starting = -1.0
		}
		msg, _ = s.InteractionResponseEdit(s.State.User.ID, i.Interaction, &discordgo.WebhookEdit{
			Content: fmt.Sprintf("Crypto %v @ $%.2f \n***%v***", ticker, starting, desc),
			AllowedMentions: &discordgo.MessageAllowedMentions{
				Parse: []discordgo.AllowedMentionType{discordgo.AllowedMentionTypeEveryone},
			},
			Embeds: []*discordgo.MessageEmbed{embed},
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.SelectMenu{
							CustomID:    "opt_crypto",
							MinValues:   0,
							MaxValues:   1,
							Placeholder: "Actions",
							Options: []discordgo.SelectMenuOption{
								{
									Label:       "Remove Alert",
									Description: "Removes this alert. Only the alerter can do this - will fail for anyone else.",
									Default:     false,
									Value:       "rm_" + ticker,
									Emoji: discordgo.ComponentEmoji{
										Name: "❌",
									},
								},
								{
									Label:       "Get Email alerts for this (this does nothing atm)",
									Description: "Recieve Email alerts.",
									Default:     false,
									Value:       "em_" + ticker,
									Emoji: discordgo.ComponentEmoji{
										Name: "📧",
									},
								},
							},
						},
					},
				},
			},
		},
		)

		componentMsgMap[fmt.Sprintf("c_%s_%s", ticker, traderID)] = ComponenetMapData{msg, traderID, i.Interaction.Member.User.ID}
		var ms *discordgo.MessageSend

		if msg == nil {
			ms = &discordgo.MessageSend{
				Content: "Crypto  " + ticker + " @everyone",
				AllowedMentions: &discordgo.MessageAllowedMentions{
					Parse: []discordgo.AllowedMentionType{discordgo.AllowedMentionTypeEveryone},
				},
			}
		} else {
			ms = &discordgo.MessageSend{
				Content: "Crypto  " + ticker + " @everyone",
				AllowedMentions: &discordgo.MessageAllowedMentions{
					Parse: []discordgo.AllowedMentionType{discordgo.AllowedMentionTypeEveryone},
				},
				Reference: msg.Reference(),
			}
		}
		msg, _ = s.ChannelMessageSendComplex(i.Interaction.ChannelID, ms)
		go func() {
			time.Sleep(time.Second * 7)
			s.ChannelMessageDelete(i.Interaction.ChannelID, msg.ID)
		}()
	},
	"rmc": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if !isAuthed(i) {
			log.Println("no perms for this action " + i.Interaction.Member.User.Username)
			return
		}
		log.Println("Calling rm-Crypto...")
		var (
			ticker string
			desc   string
		)
		argValMap := make(map[string]string)

		for _, v := range i.ApplicationCommandData().Options {
			argValMap[v.Name] = v.StringValue()
		}

		if v, ok := argValMap[COIN]; !ok {
			log.Println(errors.New("ticker not defined"))
			return
		} else {
			ticker = strings.ToUpper(v)
		}
		log.Println("Calling rm-Crypto for " + ticker)

		if v, ok := argValMap[DESC]; ok {
			desc = v
		}

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{},
		})

		traderID, err := discord.GetTraderID(i.Interaction.GuildID, i.Interaction.ChannelID)
		if err != nil {
			log.Println("TraderID is not defined.")
			return
		}

		channelType := discord.GetChannelType(i.Interaction.GuildID, i.Interaction.ChannelID)

		if channelType == -1 {
			log.Println(errors.New("invalid channel type"))
			s.InteractionResponseDelete(s.State.User.ID, i.Interaction)
			return
		}

		embed, err := rmCryptoHandler(i.Interaction.GuildID, i.Interaction.ChannelID, i.Member.User.Mention(), ticker, traderID, desc, channelType)

		if err != nil {
			log.Println(fmt.Errorf("error removing crypto: %w", err))
			s.InteractionResponseDelete(s.State.User.ID, i.Interaction)
			return
		}
		fetcher := stonks.NewFetcher()
		closing, _ := fetcher.GetCrypto(ticker, false)

		s.InteractionResponseEdit(s.State.User.ID, i.Interaction, &discordgo.WebhookEdit{
			Content: fmt.Sprintf("Crypto alert removed - **%v** @ $%.2f @everyone", ticker, closing),
			AllowedMentions: &discordgo.MessageAllowedMentions{
				Parse: []discordgo.AllowedMentionType{discordgo.AllowedMentionTypeEveryone},
			},
			Embeds: []*discordgo.MessageEmbed{embed},
		},
		)
		var msg *discordgo.Message
		if msg, ok := componentMsgMap[fmt.Sprintf("c_%s_%s", ticker, traderID)]; ok {
			str := fmt.Sprintf("This alert has been deleted - closed @ $%.2f", closing)
			s.ChannelMessageEditComplex(&discordgo.MessageEdit{
				Content:    &str,
				Components: []discordgo.MessageComponent{},
				ID:         msg.Msg.ID,
				Channel:    msg.Msg.ChannelID,
			})

			delete(componentMsgMap, fmt.Sprintf("c_%s_%s", ticker, traderID))
		}

		var ms *discordgo.MessageSend

		if msg == nil {
			ms = &discordgo.MessageSend{
				Content: "@everyone",
				AllowedMentions: &discordgo.MessageAllowedMentions{
					Parse: []discordgo.AllowedMentionType{discordgo.AllowedMentionTypeEveryone},
				},
			}
		} else {
			ms = &discordgo.MessageSend{
				Content: "@everyone",
				AllowedMentions: &discordgo.MessageAllowedMentions{
					Parse: []discordgo.AllowedMentionType{discordgo.AllowedMentionTypeEveryone},
				},
				Reference: msg.Reference(),
			}
		}

		msg, _ = s.ChannelMessageSendComplex(i.Interaction.ChannelID, ms)
		go func() {
			time.Sleep(time.Second * 10)
			s.ChannelMessageDelete(i.Interaction.ChannelID, msg.ID)
		}()
	},
	"o": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if !isAuthed(i) {
			log.Println("no perms for this action " + i.Interaction.Member.User.Username)
			return
		}
		log.Println("Calling option...")
		var (
			ticker string
			expiry string
			strike string
			entry  string
			poi    string
			stop   string
			desc   string
			year   string
			month  string
			day    string
			pt     string
		)
		argValMap := make(map[string]string)

		for _, v := range i.ApplicationCommandData().Options {
			argValMap[v.Name] = v.StringValue()
		}

		if v, ok := argValMap[TICKER]; !ok {
			discord.SendError(s, i.Interaction.ChannelID, errors.New("ticker not defined"))
			return
		} else {
			ticker = strings.ToUpper(v)
		}
		log.Println("Calling option for " + ticker)

		if v, ok := argValMap[EXPIRY]; !ok {
			discord.SendError(s, i.Interaction.ChannelID, errors.New("expiry not defined"))
			return
		} else {
			expiry = v
		}

		if v, ok := argValMap[STRIKE]; !ok {
			discord.SendError(s, i.Interaction.ChannelID, errors.New("strike not defined"))
			return
		} else {
			strike = v
		}

		if v, ok := argValMap[POI]; ok {
			poi = v
		}

		if v, ok := argValMap[STOPLOSS]; ok {
			stop = v
		}

		if v, ok := argValMap[DESC]; ok {
			desc = v
		}

		if v, ok := argValMap[PRICE]; ok {
			pt = v
		}

		if v, ok := argValMap[ENTRY]; ok {
			entry = v
		}

		month, day, year, err := utils.ParseDate(expiry) // americans :(
		if err != nil {
			discord.SendError(s, i.Interaction.ChannelID, err)
			return
		}

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{},
		})

		traderID, err := discord.GetTraderID(i.Interaction.GuildID, i.Interaction.ChannelID)
		if err != nil {
			log.Println("TraderID is not defined.")
			return
		}

		channelType := discord.GetChannelType(i.Interaction.GuildID, i.Interaction.ChannelID)

		if channelType == -1 {
			log.Println(errors.New("invalid channel type"))
			s.InteractionResponseDelete(s.State.User.ID, i.Interaction)
			return
		}
		msg, _ := s.InteractionResponseEdit(s.State.User.ID, i.Interaction, &discordgo.WebhookEdit{Content: "Thinking..."})
		var embed *discordgo.MessageEmbed
		var starting float32
		d := db.NewDB(i.Interaction.GuildID)
		if msg == nil {
			embed, starting, err = createOptionHandler(s, d, i.Interaction.ChannelID, i.Member.User.Mention(), ticker, traderID, strike, year, month, day, entry, poi, stop, pt, channelType, nil)
		} else {
			embed, starting, err = createOptionHandler(s, d, i.Interaction.ChannelID, i.Member.User.Mention(), ticker, traderID, strike, year, month, day, entry, poi, stop, pt, channelType, msg.Reference())
		}

		if err != nil {
			log.Println(fmt.Errorf("error creating createOptionHandler: %w", err).Error())
			s.FollowupMessageCreate(s.State.User.ID, i.Interaction, false, &discordgo.WebhookParams{
				Content: "Unable to create alert - " + err.Error(),
				Flags:   1 << 6,
			})
			s.InteractionResponseDelete(s.State.User.ID, i.Interaction)
			return
		}
		strikePriceFl, _ := strconv.ParseFloat(strike[:len(strike)-1], 32)
		oID := stonks.GetCode(ticker, strike[len(strike)-1:], day, month, year, float32(strikePriceFl))
		msg, err = s.InteractionResponseEdit(s.State.User.ID, i.Interaction, &discordgo.WebhookEdit{
			Content: fmt.Sprintf("%v **%v/%v/%v %v** @ $%.2f \n***%v***", ticker, month, day, year, strike, starting, desc),
			AllowedMentions: &discordgo.MessageAllowedMentions{
				Parse: []discordgo.AllowedMentionType{discordgo.AllowedMentionTypeEveryone},
			},
			Embeds: []*discordgo.MessageEmbed{embed},
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.SelectMenu{
							CustomID:    "opt_options",
							MinValues:   0,
							MaxValues:   1,
							Placeholder: "Actions",
							Options: []discordgo.SelectMenuOption{
								{
									Label:       "Remove Alert",
									Description: "Removes this alert. Only the alerter can do this - will fail for anyone else.",
									Default:     false,
									Value:       "rm_" + oID,
									Emoji: discordgo.ComponentEmoji{
										Name: "❌",
									},
								},
								{
									Label:       "Get Email alerts for this (this does nothing atm)",
									Description: "Recieve Email alerts.",
									Default:     false,
									Value:       "em_" + stonks.GetCode(ticker, strike[len(strike)-1:], day, month, year, float32(strikePriceFl)),
									Emoji: discordgo.ComponentEmoji{
										Name: "📧",
									},
								},
							},
						},
					},
				},
			},
		},
		)

		componentMsgMap[fmt.Sprintf("%s_%s", oID, traderID)] = ComponenetMapData{msg, traderID, i.Interaction.Member.User.ID}

		if err != nil {
			log.Println(err)
		}
		var ms *discordgo.MessageSend
		includeChart := true

		chart, err := graph.Get15MStocksChart(ticker)

		if err != nil {
			includeChart = false
		}

		file, err := os.Open(chart)
		if err != nil {
			includeChart = false
		}
		defer func() {
			file.Close()

			os.Remove(chart)
		}()
		if msg == nil {
			ms = &discordgo.MessageSend{
				Content: fmt.Sprintf("%v **%v/%v %v* \n***%v***  @everyone", ticker, month, day, strike, desc),
				AllowedMentions: &discordgo.MessageAllowedMentions{
					Parse: []discordgo.AllowedMentionType{discordgo.AllowedMentionTypeEveryone},
				},
			}
		} else {
			ms = &discordgo.MessageSend{
				Content: fmt.Sprintf("%v **%v/%v %v** \n***%v*** @everyone", ticker, month, day, strike, desc),
				AllowedMentions: &discordgo.MessageAllowedMentions{
					Parse: []discordgo.AllowedMentionType{discordgo.AllowedMentionTypeEveryone},
				},
				Reference: msg.Reference(),
			}
		}

		if includeChart {
			ms.Files = []*discordgo.File{
				{
					Name:        chart,
					ContentType: "image/png",
					Reader:      file,
				},
			}
			ms.Embeds = []*discordgo.MessageEmbed{
				{
					Image: &discordgo.MessageEmbedImage{
						URL: "attachment://" + chart,
					},
				},
			}
		}
		msg, _ = s.ChannelMessageSendComplex(i.Interaction.ChannelID, ms)
		go func() {
			time.Sleep(time.Second * 30)
			s.ChannelMessageDelete(i.Interaction.ChannelID, msg.ID)
		}()
	},
	"rmo": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if !isAuthed(i) {
			log.Println("no perms for this action " + i.Interaction.Member.User.Username)
			return
		}
		log.Println("Calling rm-option...")
		var (
			ticker string
			expiry string
			strike string
			desc   string
			year   string
			month  string
			day    string
		)
		log.Println("Calling option for " + ticker)
		argValMap := make(map[string]string)

		for _, v := range i.ApplicationCommandData().Options {
			argValMap[v.Name] = v.StringValue()
		}

		if v, ok := argValMap[TICKER]; !ok {
			discord.SendError(s, i.Interaction.ChannelID, errors.New("ticker not defined"))
			return
		} else {
			ticker = strings.ToUpper(v)
		}

		if v, ok := argValMap[EXPIRY]; !ok {
			discord.SendError(s, i.Interaction.ChannelID, errors.New("expiry not defined"))
			return
		} else {
			expiry = v
		}

		if v, ok := argValMap[STRIKE]; !ok {
			discord.SendError(s, i.Interaction.ChannelID, errors.New("strike not defined"))
			return
		} else {
			strike = v
		}
		if v, ok := argValMap[DESC]; ok {
			desc = v
		}

		month, day, year, err := utils.ParseDate(expiry) // americans :(
		if err != nil {
			discord.SendError(s, i.Interaction.ChannelID, err)
			return
		}

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{},
		})

		traderID, err := discord.GetTraderID(i.Interaction.GuildID, i.Interaction.ChannelID)
		if err != nil {
			log.Println("TraderID is not defined.")
			return
		}

		channelType := discord.GetChannelType(i.Interaction.GuildID, i.Interaction.ChannelID)

		if channelType == -1 {
			log.Println(errors.New("invalid channel type"))
			s.InteractionResponseDelete(s.State.User.ID, i.Interaction)
			return
		}
		strikePriceFl, _ := strconv.ParseFloat(strike[:len(strike)-1], 32)
		cType := strike[len(strike)-1:]
		oID := stonks.GetCode(ticker, cType, day, month, year, float32(strikePriceFl))
		prettyStr := utils.NiceStr(ticker, cType, day, month, year, float32(strikePriceFl))

		embed, err := rmOptionHandler(i.Interaction.GuildID, i.Interaction.ChannelID, i.Member.User.Mention(), oID, traderID, prettyStr, desc, channelType)

		if err != nil {
			log.Println(fmt.Errorf("error removing Option: %w", err))
			s.InteractionResponseDelete(s.State.User.ID, i.Interaction)
			return
		}
		fetcher := stonks.NewFetcher()
		closing, _, _ := fetcher.GetOption(ticker, cType, day, month, year, float32(strikePriceFl), 0)

		s.InteractionResponseEdit(s.State.User.ID, i.Interaction, &discordgo.WebhookEdit{
			Content: fmt.Sprintf("Option alert removed - **%v** @ $%.2f @everyone", prettyStr, closing),
			AllowedMentions: &discordgo.MessageAllowedMentions{
				Parse: []discordgo.AllowedMentionType{discordgo.AllowedMentionTypeEveryone},
			},
			Embeds: []*discordgo.MessageEmbed{embed},
		},
		)

		var msg *discordgo.Message
		if msg, ok := componentMsgMap[fmt.Sprintf("%s_%s", oID, traderID)]; ok {
			str := fmt.Sprintf("This alert has been deleted - closed @ $%.2f", closing)
			s.ChannelMessageEditComplex(&discordgo.MessageEdit{
				Content:    &str,
				Components: []discordgo.MessageComponent{},
				ID:         msg.Msg.ID,
				Channel:    msg.Msg.ChannelID,
			})

			delete(componentMsgMap, fmt.Sprintf("%s_%s", oID, traderID))
		}

		var ms *discordgo.MessageSend

		if msg == nil {
			ms = &discordgo.MessageSend{
				Content: "@everyone",
				AllowedMentions: &discordgo.MessageAllowedMentions{
					Parse: []discordgo.AllowedMentionType{discordgo.AllowedMentionTypeEveryone},
				},
			}
		} else {
			ms = &discordgo.MessageSend{
				Content: "@everyone",
				AllowedMentions: &discordgo.MessageAllowedMentions{
					Parse: []discordgo.AllowedMentionType{discordgo.AllowedMentionTypeEveryone},
				},
				Reference: msg.Reference(),
			}
		}

		msg, _ = s.ChannelMessageSendComplex(i.Interaction.ChannelID, ms)
		go func() {
			time.Sleep(time.Second * 10)
			s.ChannelMessageDelete(i.Interaction.ChannelID, msg.ID)
		}()
	},
	"c15": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if !isAuthedLower(i) {
			log.Println("no perms for this action " + i.Interaction.Member.User.Username)
			return
		}

		log.Println("Calling c15...")

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "",
			},
		})

		chart, err := graph.Get15MStocksChart(i.ApplicationCommandData().Options[0].StringValue())

		if err != nil {
			return
		}
		file, err := os.Open(chart)
		defer func() {
			file.Close()

			os.Remove(chart)
		}()

		if err != nil {
			return
		}
		s.InteractionResponseEdit(s.State.User.ID, i.Interaction, &discordgo.WebhookEdit{
			Files: []*discordgo.File{
				{
					Name:        chart,
					ContentType: "image/png",
					Reader:      file,
				},
			},
			Embeds: []*discordgo.MessageEmbed{
				{
					Image: &discordgo.MessageEmbedImage{
						URL: "attachment://" + chart,
					},
				},
			},
		})
	},
	"ch": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if !isAuthedLower(i) {
			log.Println("no perms for this action " + i.Interaction.Member.User.Username)
			return
		}

		log.Println("Calling ch...")

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "",
			},
		})

		chart, err := graph.GetHStocksChart(i.ApplicationCommandData().Options[0].StringValue())

		if err != nil {
			return
		}
		file, err := os.Open(chart)
		defer func() {
			file.Close()

			os.Remove(chart)
		}()

		if err != nil {
			return
		}

		s.InteractionResponseEdit(s.State.User.ID, i.Interaction, &discordgo.WebhookEdit{
			Files: []*discordgo.File{
				{
					Name:        chart,
					ContentType: "image/png",
					Reader:      file,
				},
			},
			Embeds: []*discordgo.MessageEmbed{
				{
					Image: &discordgo.MessageEmbedImage{
						URL: "attachment://" + chart,
					},
				},
			},
		})
	},
	"cd": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if !isAuthedLower(i) {
			log.Println("no perms for this action " + i.Interaction.Member.User.Username)
			return
		}

		log.Println("Calling cd...")

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "",
			},
		})

		chart, err := graph.GetDStocksChart(i.ApplicationCommandData().Options[0].StringValue())

		if err != nil {
			return
		}
		file, err := os.Open(chart)
		defer func() {
			file.Close()

			os.Remove(chart)
		}()

		if err != nil {
			return
		}
		s.InteractionResponseEdit(s.State.User.ID, i.Interaction, &discordgo.WebhookEdit{
			Files: []*discordgo.File{
				{
					Name:        chart,
					ContentType: "image/png",
					Reader:      file,
				},
			},
			Embeds: []*discordgo.MessageEmbed{
				{
					Image: &discordgo.MessageEmbedImage{
						URL: "attachment://" + chart,
					},
				},
			},
		})
	},
	"cc15": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if !isAuthedLower(i) {
			log.Println("no perms for this action " + i.Interaction.Member.User.Username)
			return
		}

		log.Println("Calling cc15...")

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "",
			},
		})

		chart, err := graph.Get15MCryptoChart(i.ApplicationCommandData().Options[0].StringValue())
		if err != nil {
			return
		}

		file, err := os.Open(chart)
		defer func() {
			file.Close()

			os.Remove(chart)
		}()

		if err != nil {
			return
		}

		s.InteractionResponseEdit(s.State.User.ID, i.Interaction, &discordgo.WebhookEdit{
			Files: []*discordgo.File{
				{
					Name:        chart,
					ContentType: "image/png",
					Reader:      file,
				},
			},
			Embeds: []*discordgo.MessageEmbed{
				{
					Image: &discordgo.MessageEmbedImage{
						URL: "attachment://" + chart,
					},
				},
			},
		})
	},
	"cch": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if !isAuthedLower(i) {
			log.Println("no perms for this action " + i.Interaction.Member.User.Username)
			return
		}

		log.Println("Calling cch...")

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "",
			},
		})

		chart, err := graph.GetHCryptoChart(i.ApplicationCommandData().Options[0].StringValue())

		if err != nil {
			return
		}

		file, err := os.Open(chart)
		defer func() {
			file.Close()

			os.Remove(chart)
		}()

		if err != nil {
			return
		}

		s.InteractionResponseEdit(s.State.User.ID, i.Interaction, &discordgo.WebhookEdit{
			Files: []*discordgo.File{
				{
					Name:        chart,
					ContentType: "image/png",
					Reader:      file,
				},
			},
			Embeds: []*discordgo.MessageEmbed{
				{
					Image: &discordgo.MessageEmbedImage{
						URL: "attachment://" + chart,
					},
				},
			},
		})
	},
	"ccd": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if !isAuthedLower(i) {
			log.Println("no perms for this action " + i.Interaction.Member.User.Username)
			return
		}

		log.Println("Calling ccd...")

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "",
			},
		})

		chart, err := graph.GetDCryptoChart(i.ApplicationCommandData().Options[0].StringValue())

		if err != nil {
			return
		}

		file, err := os.Open(chart)
		defer func() {
			file.Close()

			os.Remove(chart)
		}()

		if err != nil {
			return
		}

		s.InteractionResponseEdit(s.State.User.ID, i.Interaction, &discordgo.WebhookEdit{
			Files: []*discordgo.File{
				{
					Name:        chart,
					ContentType: "image/png",
					Reader:      file,
				},
			},
			Embeds: []*discordgo.MessageEmbed{
				{
					Image: &discordgo.MessageEmbedImage{
						URL: "attachment://" + chart,
					},
				},
			},
		})
	},
	"opt_stock": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if !isAuthed(i) {
			log.Println("Not Authoirzed")
			return
		}
		data := i.MessageComponentData().Values
		if len(data) != 1 {
			log.Println("No data")
			return
		}
		method, ticker, err := parseComponentData(data[0])
		if err != nil {
			log.Println("Incorrect Arguments Length")
			return
		}

		traderID, err := discord.GetTraderID(i.Interaction.GuildID, i.Interaction.ChannelID)
		if err != nil {
			log.Println("TraderID is not defined.")
			return
		}

		switch method {
		case "rm":
			if msg, ok := componentMsgMap[fmt.Sprintf("s_%s_%s", ticker, traderID)]; ok {
				if i.Interaction.Member.User.ID == msg.UserID {
					s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{},
					})

					channelType := discord.GetChannelType(i.Interaction.GuildID, i.Interaction.ChannelID)

					embed, err := rmStockHandler(i.Interaction.GuildID, i.Interaction.ChannelID, i.Member.User.Mention(), ticker, traderID, "", channelType)

					if err != nil {
						log.Println(fmt.Errorf("error removing stock: %w", err))
						s.InteractionResponseDelete(s.State.User.ID, i.Interaction)
						return
					}

					fetcher := stonks.NewFetcher()
					closing, _ := fetcher.GetStock(ticker)

					ms := &discordgo.MessageSend{
						Content: fmt.Sprintf("Stock alert removed - **%v** @ $%.2f @everyone", ticker, closing),
						AllowedMentions: &discordgo.MessageAllowedMentions{
							Parse: []discordgo.AllowedMentionType{discordgo.AllowedMentionTypeEveryone},
						},
					}

					s.ChannelMessageSendComplex(i.Interaction.ChannelID, ms)

					s.InteractionResponseEdit(s.State.User.ID, i.Interaction, &discordgo.WebhookEdit{
						AllowedMentions: &discordgo.MessageAllowedMentions{
							Parse: []discordgo.AllowedMentionType{discordgo.AllowedMentionTypeEveryone},
						},
						Embeds: []*discordgo.MessageEmbed{embed},
					},
					)
					str := fmt.Sprintf("This alert has been deleted - closed @ $%.2f", closing)
					s.ChannelMessageEditComplex(&discordgo.MessageEdit{
						Content:    &str,
						Components: []discordgo.MessageComponent{},
						ID:         msg.Msg.ID,
						Channel:    msg.Msg.ChannelID,
					})

					delete(componentMsgMap, fmt.Sprintf("s_%s_%s", ticker, traderID))

				}
			}
		}
	},
	"opt_short": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if !isAuthed(i) {
			log.Println("Not Authoirzed")
			return
		}
		data := i.MessageComponentData().Values
		if len(data) != 1 {
			log.Println("No data")
			return
		}
		method, ticker, err := parseComponentData(data[0])
		if err != nil {
			log.Println("Incorrect Arguments Length")
			return
		}

		traderID, err := discord.GetTraderID(i.Interaction.GuildID, i.Interaction.ChannelID)
		if err != nil {
			log.Println("TraderID is not defined.")
			return
		}

		switch method {
		case "rm":
			if msg, ok := componentMsgMap[fmt.Sprintf("sh_%s_%s", ticker, traderID)]; ok {
				if i.Interaction.Member.User.ID == msg.UserID {

					channelType := discord.GetChannelType(i.Interaction.GuildID, i.Interaction.ChannelID)

					embed, err := rmShortHandler(i.Interaction.GuildID, i.Interaction.ChannelID, i.Member.User.Mention(), ticker, traderID, "", channelType)

					if err != nil {
						log.Println(fmt.Errorf("error removing short: %w", err))
						s.InteractionResponseDelete(s.State.User.ID, i.Interaction)
						return
					}
					fetcher := stonks.NewFetcher()
					closing, _ := fetcher.GetStock(ticker)

					ms := &discordgo.MessageSend{
						Content: fmt.Sprintf("Crypto alert removed - **%v** @ $%.2f @everyone", ticker, closing),
						AllowedMentions: &discordgo.MessageAllowedMentions{
							Parse: []discordgo.AllowedMentionType{discordgo.AllowedMentionTypeEveryone},
						},
					}

					s.ChannelMessageSendComplex(i.Interaction.ChannelID, ms)

					s.InteractionResponseEdit(s.State.User.ID, i.Interaction, &discordgo.WebhookEdit{
						AllowedMentions: &discordgo.MessageAllowedMentions{
							Parse: []discordgo.AllowedMentionType{discordgo.AllowedMentionTypeEveryone},
						},
						Embeds: []*discordgo.MessageEmbed{embed},
					},
					)

					str := fmt.Sprintf("This alert has been deleted - closed @ $%.2f", closing)
					s.ChannelMessageEditComplex(&discordgo.MessageEdit{
						Content:    &str,
						Components: []discordgo.MessageComponent{},
						ID:         msg.Msg.ID,
						Channel:    msg.Msg.ChannelID,
					})

					delete(componentMsgMap, fmt.Sprintf("sh_%s_%s", ticker, traderID))

				}
			}
		}

	},
	"opt_crypto": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if !isAuthed(i) {
			log.Println("Not Authoirzed")
			return
		}
		data := i.MessageComponentData().Values
		if len(data) != 1 {
			log.Println("No data")
			return
		}
		method, ticker, err := parseComponentData(data[0])
		if err != nil {
			log.Println("Incorrect Arguments Length")
			return
		}

		traderID, err := discord.GetTraderID(i.Interaction.GuildID, i.Interaction.ChannelID)
		if err != nil {
			log.Println("TraderID is not defined.")
			return
		}

		switch method {
		case "rm":
			if msg, ok := componentMsgMap[fmt.Sprintf("c_%s_%s", ticker, traderID)]; ok {
				if i.Interaction.Member.User.ID == msg.UserID {

					channelType := discord.GetChannelType(i.Interaction.GuildID, i.Interaction.ChannelID)

					embed, err := rmCryptoHandler(i.Interaction.GuildID, i.Interaction.ChannelID, i.Member.User.Mention(), ticker, traderID, "", channelType)

					if err != nil {
						log.Println(fmt.Errorf("error removing crypto: %w", err))
						s.InteractionResponseDelete(s.State.User.ID, i.Interaction)
						return
					}
					fetcher := stonks.NewFetcher()
					closing, _ := fetcher.GetCrypto(ticker, false)

					ms := &discordgo.MessageSend{
						Content: fmt.Sprintf("Crypto alert removed - **%v** @ $%.2f @everyone", ticker, closing),
						AllowedMentions: &discordgo.MessageAllowedMentions{
							Parse: []discordgo.AllowedMentionType{discordgo.AllowedMentionTypeEveryone},
						},
					}

					s.ChannelMessageSendComplex(i.Interaction.ChannelID, ms)

					s.InteractionResponseEdit(s.State.User.ID, i.Interaction, &discordgo.WebhookEdit{
						AllowedMentions: &discordgo.MessageAllowedMentions{
							Parse: []discordgo.AllowedMentionType{discordgo.AllowedMentionTypeEveryone},
						},
						Embeds: []*discordgo.MessageEmbed{embed},
					},
					)
					str := fmt.Sprintf("This alert has been deleted - closed @ $%.2f", closing)
					s.ChannelMessageEditComplex(&discordgo.MessageEdit{
						Content:    &str,
						Components: []discordgo.MessageComponent{},
						ID:         msg.Msg.ID,
						Channel:    msg.Msg.ChannelID,
					})

					delete(componentMsgMap, fmt.Sprintf("c_%s_%s", ticker, traderID))
				}
			}
		}

	},
	"opt_options": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if !isAuthed(i) {
			log.Println("Not Authoirzed")
			return
		}
		data := i.MessageComponentData().Values
		if len(data) != 1 {
			log.Println("No data")
			return
		}

		method, oID, err := parseComponentData(data[0])
		if err != nil {
			log.Println("Incorrect Arguments Length")
			return
		}

		traderID, err := discord.GetTraderID(i.Interaction.GuildID, i.Interaction.ChannelID)
		if err != nil {
			log.Println("TraderID is not defined.")
			return
		}

		switch method {
		case "rm":
			if msg, ok := componentMsgMap[fmt.Sprintf("%s_%s", oID, traderID)]; ok {
				if i.Interaction.Member.User.ID == msg.UserID {

					ticker, cType, day, month, year, strike, err := parseOptionFromOID(oID)
					if err != nil {
						log.Println("Invalid oID")
						return
					}

					channelType := discord.GetChannelType(i.Interaction.GuildID, i.Interaction.ChannelID)

					strikePriceFl, _ := strconv.ParseFloat(strike, 32)
					prettyStr := utils.NiceStr(ticker, cType, day, month, year, float32(strikePriceFl))

					s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{},
					},
					)

					embed, err := rmOptionHandler(i.Interaction.GuildID, i.Interaction.ChannelID, i.Member.User.Mention(), oID, traderID, prettyStr, "", channelType)

					if err != nil {
						log.Println(fmt.Errorf("error removing option: %w", err))
						s.InteractionResponseDelete(s.State.User.ID, i.Interaction)
						return
					}

					fetcher := stonks.NewFetcher()
					closing, _, err := fetcher.GetOption(ticker, cType, day, month, year, float32(strikePriceFl), 0)

					if err != nil {
						closing = -1
					}

					ms := &discordgo.MessageSend{
						Content: fmt.Sprintf("Option alert removed - **%s** @ $%.2f @everyone", prettyStr, closing),
						AllowedMentions: &discordgo.MessageAllowedMentions{
							Parse: []discordgo.AllowedMentionType{discordgo.AllowedMentionTypeEveryone},
						},
					}

					s.ChannelMessageSendComplex(i.Interaction.ChannelID, ms)

					s.InteractionResponseEdit(s.State.User.ID, i.Interaction, &discordgo.WebhookEdit{
						AllowedMentions: &discordgo.MessageAllowedMentions{
							Parse: []discordgo.AllowedMentionType{discordgo.AllowedMentionTypeEveryone},
						},
						Embeds: []*discordgo.MessageEmbed{embed},
					},
					)
					str := fmt.Sprintf("This alert has been deleted - closed @ $%.2f", closing)
					s.ChannelMessageEditComplex(&discordgo.MessageEdit{
						Content:    &str,
						Components: []discordgo.MessageComponent{},
						ID:         msg.Msg.ID,
						Channel:    msg.Msg.ChannelID,
					})

					delete(componentMsgMap, fmt.Sprintf("%s_%s", oID, traderID))

				}
			}
		}

	},
}

func parseComponentData(data string) (method string, ticker string, err error) {
	dataByte := []byte(data)
	splitData := bytes.Split(dataByte, []byte("_"))
	if len(splitData) != 2 {
		err = errors.New("Incorrect Arguments length")
		return
	}
	method = string(splitData[0])
	ticker = string(splitData[1])
	return
}

func parseOptionFromOID(oID string) (ticker string, cType string, day string, month string, year string, strike string, err error) {
	reg := regexp.MustCompile(`[A-Z]+`)
	ticker = reg.FindString(oID)
	oID = oID[len(ticker):]
	if len(oID) != 15 {
		err = errors.New("Invalid oID")
		return
	}
	year = "20" + oID[:2]
	month = oID[2:4]
	day = oID[4:6]
	cType = oID[6:7]
	strikeInt, err := strconv.Atoi(oID[7:])
	if err != nil {
		err = errors.New("Invalid oID")
		return
	}
	strike = strconv.Itoa(strikeInt / 1000)
	return
}
