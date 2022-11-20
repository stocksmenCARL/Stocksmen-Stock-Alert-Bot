package handlers

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/m1k8/kronos/pkg/M1K8/harpe/pkg/config"
)

var servers = make(map[string]config.ServerConfig)

func init() {
	cfgFile, err := os.Open("config.json")

	if err != nil {
		panic("")
	}

	defer cfgFile.Close()

	byteValue, err := ioutil.ReadAll(cfgFile)

	if err != nil {
		panic("")
	}
	var cfg config.Config

	json.Unmarshal(byteValue, &cfg)
	for _, v := range cfg.ServersCfg {
		servers[v.ID] = v
	}
}

func isAuthed(i *discordgo.InteractionCreate) bool {
	rolesMap := map[string]bool{}
	for _, v := range i.Interaction.Member.Roles {
		rolesMap[strings.ToLower(v)] = true
	}

	if v, ok := servers[i.Interaction.GuildID]; ok {
		for _, id := range v.Whitelisted_UIDs {
			if i.Interaction.Member.User.ID == id || id == "*" {
				return true
			}
		}

		for _, role := range v.Roles {
			if _, ok := rolesMap[role]; ok || role == "*" {
				return true
			}
		}
	} else {
		return false
	}
	return false
}

func isAuthedLower(i *discordgo.InteractionCreate) bool {
	rolesMap := map[string]bool{}
	for _, v := range i.Interaction.Member.Roles {
		rolesMap[strings.ToLower(v)] = true
	}

	if v, ok := servers[i.Interaction.GuildID]; ok {
		for _, id := range v.Whitelisted_UIDs {
			if i.Interaction.Member.User.ID == id || id == "*" {
				return true
			}
		}

		for _, role := range v.Roles {
			if _, ok := rolesMap[role]; ok || role == "*" || role == v.AlertRole {
				return true
			}
		}
	} else {
		return false
	}
	return false
}
