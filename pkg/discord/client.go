package discord

import (
	"regexp"

	"github.com/bwmarrin/discordgo"
)

func GuildCreate(s *discordgo.Session, event *discordgo.GuildCreate) {

	if event.Guild.Unavailable {
		return
	}

	for _, channel := range event.Guild.Channels {
		if channel.ID == event.Guild.ID {
			_, _ = s.ChannelMessageSend(channel.ID, "StocksBot is ready!")
			return
		}
	}
}

func Ready(s *discordgo.Session, event *discordgo.Ready) {
	err := s.UpdateStatusComplex(discordgo.UpdateStatusData{
		Activities: []*discordgo.Activity{
			{
				Name: "Stocksmen to the Moon",
				Type: discordgo.ActivityTypeGame,
			},
		},
		Status: "online",
	})

	if err != nil {
		panic(err)
	}
}

func MessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}
}

func getTickersFromString(s string) []string {
	r, _ := regexp.Compile(`\$([aA-zZ])+`)

	matches := r.FindAll([]byte(s), -1)

	tickerSlice := make([]string, len(matches))

	for _, v := range matches {
		tickerSlice = append(tickerSlice, string(v[1:]))
	}
	return tickerSlice
}
