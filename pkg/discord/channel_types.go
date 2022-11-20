package discord

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
)

const (
	SWING     = 0
	DAY       = 1
	WATCHLIST = 2
	ALERTS    = 3
)

type Config struct {
	ServersCfg []ServerConfig `json:"servers"`
	DiscordCFG DiscordConfig  `json:"discord"`
}

type DiscordConfig struct {
	API   string `json:"DISCORD_API"`
	AppID string `json:"APP_ID"`
}

type ServerConfig struct {
	ID               string          `json:"id"`
	ChannelConfig    []ChannelConfig `json:"channels"`
	Roles            []string        `json:"allowed_roles"`
	Whitelisted_UIDs []string        `json:"whitelisted_ids"`
	AlertRole        string          `json:"alert_role"`
}
type ChannelConfig struct {
	TraderID  string `json:"trader_id"`
	Day       string `json:"day_trades"`
	Swing     string `json:"swings"`
	Watchlist string `json:"watchlist"`
	Alerts    string `json:"alerts"`
	EOD       string `json:"eod"`
	Author    string `json:"author"`
}

var (
	servers = make(map[string]ServerConfig)
)

func init() {
	cfgFile, err := os.Open("config.json")

	if err != nil {
		panic("Unable to open config.json!")
	}

	defer cfgFile.Close()

	byteValue, err := ioutil.ReadAll(cfgFile)

	if err != nil {
		panic("Error reading config.json!")
	}
	var cfg Config
	json.Unmarshal(byteValue, &cfg)

	if len(cfg.ServersCfg) == 0 {
		panic("Servers not confugred!")
	} else {
		for _, v := range cfg.ServersCfg {
			servers[v.ID] = v
		}
	}
}

func GetChannelType(guildID, channelID string) int {

	if s, ok := servers[guildID]; ok {
		for _, channel := range s.ChannelConfig {
			switch channelID {
			case channel.Swing:
				return SWING
			case channel.Day:
				return DAY
			case channel.Watchlist:
				return WATCHLIST
			}
		}
	}
	return -1
}

func GetTraderID(guildID, channelID string) (string, error) {
	var traderID string
	var err error

	if s, ok := servers[guildID]; ok {
		for _, channel := range s.ChannelConfig {
			switch channelID {
			case channel.Swing:
				traderID = channel.TraderID
			case channel.Day:
				traderID = channel.TraderID
			case channel.Watchlist:
				traderID = channel.TraderID
			case channel.Alerts:
				traderID = channel.TraderID
			case channel.EOD:
				traderID = channel.TraderID
			}
		}
	} else {
		err = errors.New("Server not configured")
		return "", err
	}
	return traderID, err

}

func GetAuthorID(guildID, channelID string) (string, error) {
	var author string
	var err error

	if s, ok := servers[guildID]; ok {
		for _, channel := range s.ChannelConfig {
			switch channelID {
			case channel.Swing:
				author = channel.Author
			case channel.Day:
				author = channel.Author
			case channel.Watchlist:
				author = channel.Author
			case channel.Alerts:
				author = channel.Author
			case channel.EOD:
				author = channel.Author
			}
		}
	} else {
		err = errors.New("Server not configured")
		return "", err
	}
	return author, err

}
