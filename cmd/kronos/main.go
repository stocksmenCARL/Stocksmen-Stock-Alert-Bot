package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/m1k8/kronos/pkg/discord"
	"github.com/m1k8/kronos/pkg/discord/handlers"
	"github.com/m1k8/kronos/pkg/worker"
)

var token string
var s *discordgo.Session
var cfg discord.Config

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

	json.Unmarshal(byteValue, &cfg)

	if cfg.DiscordCFG.API == "" {
		panic("")
	}

	token = cfg.DiscordCFG.API
	s, err = discordgo.New("Bot " + token)
	if err != nil {
		panic("")
	}

	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		switch i.Type {
		case discordgo.InteractionMessageComponent:
			if h, ok := handlers.CommandHandlers[i.MessageComponentData().CustomID]; ok {
				h(s, i)
			}

		case discordgo.InteractionApplicationCommand:
			if h, ok := handlers.CommandHandlers[i.ApplicationCommandData().Name]; ok {
				h(s, i)
			}
		}
	})
}

func main() {
	appId := cfg.DiscordCFG.AppID

	s.AddHandler(discord.MessageCreate)
	s.AddHandler(discord.GuildCreate)

	s.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMessages | discordgo.IntentsGuildVoiceStates

	for _, v := range cfg.ServersCfg {
		for _, channel := range v.ChannelConfig {
			if channel.EOD == "" || channel.TraderID == "" || channel.Day == "" {
				continue
			}
			go worker.KeepTrack(s, v.ID, channel.EOD, channel.TraderID, channel.Author)
		}
	}

	for _, v := range handlers.Commands {
		c, err := s.ApplicationCommandCreate(appId, "", v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}

		log.Println("Created " + c.Name)
	}

	defer func() {
		//cleanup
		cmds, _ := s.ApplicationCommands(appId, "")
		for _, cmd := range cmds {
			err := s.ApplicationCommandDelete(appId, "", cmd.ID)
			if err != nil {
				log.Println(fmt.Errorf("error removing %v: %w", cmd.Name, err))
			} else {
				log.Println("Removed " + cmd.Name)
			}
		}
	}()

	fmt.Println("Ready!")

	//opens websocket connection
	err := s.Open()
	if err != nil {
		panic(err)
	}

	defer func() {
		sc := make(chan os.Signal, 1)
		signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
		<-sc
		s.Close()
	}()
}
