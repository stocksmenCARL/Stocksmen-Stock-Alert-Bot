package handlers

import "github.com/bwmarrin/discordgo"

var Commands = []*discordgo.ApplicationCommand{
	{
		Name:        "s",
		Description: "Adds a new stock alert.",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "ticker",
				Description: "The stock ticker you want to alert",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "scale_pt",
				Description: "The PT you want the underlying to hit where its is suggested to scale out.",
				Required:    false,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "exit_pt",
				Description: "The PT you want the underlying to hit.",
				Required:    false,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "desc",
				Description: "Any additional info you want included in the alert",
				Required:    false,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "stop",
				Description: "A stop loss that will expire the alert with the message STOP LOSS HIT",
				Required:    false,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "poi",
				Description: "Point of interest. Will alert to enter when this is hit (when the price reaches +-0.5%)",
				Required:    false,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "entry",
				Description: "The price of the alerters entry, if different from the current price",
				Required:    false,
			},
		},
	},
	{
		Name:        "sh",
		Description: "Adds a new short alert.",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "ticker",
				Description: "The short ticker you want to alert",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "scale_pt",
				Description: "The PT you want the underlying to hit where its is suggested to scale out.",
				Required:    false,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "exit_pt",
				Description: "The PT you want the underlying to hit.",
				Required:    false,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "desc",
				Description: "Any additional info you want included in the alert",
				Required:    false,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "stop",
				Description: "A stop loss that will expire the alert with the message STOP LOSS HIT",
				Required:    false,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "poi",
				Description: "Point of interest. Will alert to enter when this is hit (when the price reaches +-0.5%)",
				Required:    false,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "entry",
				Description: "The price of the alerters entry, if different from the current price",
				Required:    false,
			},
		},
	},
	{
		Name:        "c",
		Description: "Adds a new crypto alert.",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "coin",
				Description: "The coin you want to alert",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "scale_pt",
				Description: "The PT you want the underlying to hit where its is suggested to scale out.",
				Required:    false,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "exit_pt",
				Description: "The PT you want the underlying to hit.",
				Required:    false,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "desc",
				Description: "Any additional info you want included in the alert",
				Required:    false,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "stop",
				Description: "A stop loss that will expire the alert with the message STOP LOSS HIT",
				Required:    false,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "poi",
				Description: "Point of interest. Will alert to enter when this is hit (when the price reaches +-0.5%)",
				Required:    false,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "entry",
				Description: "The price of the alerters entry, if different from the current price",
				Required:    false,
			},
		},
	},
	{
		Name:        "o",
		Description: "Adds a new options alert. The alert will automatically remove itself after the strike date.",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "ticker",
				Description: "The stock you want the option to be based on",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "expiry",
				Description: "The expiry of the contract. If the expiry is this year, then mm/dd. Else, mm/dd/yy.",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "strike",
				Description: "The strike price of the contract + the contract type, e.g. 140C, 55.50P",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "desc",
				Description: "Any additional info you want included in the alert",
				Required:    false,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "entry",
				Description: "The price of the alerters entry, if different from the current price",
				Required:    false,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "stop",
				Description: "A stop loss (based on the underlying) that will expire the alert",
				Required:    false,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "poi",
				Description: "Point of interest. Will alert to enter when this is hit (when the underlying price reaches +-0.5%)",
				Required:    false,
			},
		},
	},
	{
		Name:        "rms",
		Description: "Removes a stock alert.",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "ticker",
				Description: "The stock ticker you want to remove",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "desc",
				Description: "Any additional info you want included in the alert",
				Required:    false,
			},
		},
	},
	{
		Name:        "rmsh",
		Description: "Removes a short alert.",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "ticker",
				Description: "The short ticker you want to remove",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "desc",
				Description: "Any additional info you want included in the alert",
				Required:    false,
			},
		},
	},
	{
		Name:        "rmc",
		Description: "Removes a crypto alert.",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "coin",
				Description: "The coin you want to remove",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "desc",
				Description: "Any additional info you want included in the alert",
				Required:    false,
			},
		},
	},
	{
		Name:        "rmo",
		Description: "Removes an options alert.",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "ticker",
				Description: "The stock you want the option to be based on",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "expiry",
				Description: "The expiry of the contract. If the expiry is this year, then mm/dd. Else, mm/ddd/yy.",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "strike",
				Description: "The strike price of the contract + the contract type, e.g. 140C, 55.50P",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "desc",
				Description: "Any additional info you want included in the alert",
				Required:    false,
			},
		},
	},
	{
		Name:        "nuke",
		Description: "Nuke.",
	},
	{
		Name:        "all",
		Description: "Show all active alerts.",
	},
	{
		Name:        "refresh",
		Description: "Refresh all alerts. Must be called after a restart",
	},
	{
		Name:        "alert",
		Description: "Alert whatever you type",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "msg",
				Description: "The content of the alert.",
				Required:    true,
			},
		},
	},
	{
		Name:        "tracker",
		Description: "Print out the End of Day tracker of day trades for the previous day",
	},
	{
		Name:        "c15",
		Description: "Stocks 15m Chart",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "ticker",
				Description: "Ticker to get the 15m chart of",
				Required:    true,
			},
		},
	},
	{
		Name:        "ch",
		Description: "Stocks Hourly Chart",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "ticker",
				Description: "Ticker to get the Hourly chart of",
				Required:    true,
			},
		},
	},
	{
		Name:        "cd",
		Description: "Stocks Daily Chart",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "ticker",
				Description: "Ticker to get the Daily chart of",
				Required:    true,
			},
		},
	},
	{
		Name:        "cc15",
		Description: "Crypto 15m Chart",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "coin",
				Description: "Coin to get the 15m chart of",
				Required:    true,
			},
		},
	},
	{
		Name:        "cch",
		Description: "Crypto Hourly Chart",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "coin",
				Description: "coin to get the Hourly chart of",
				Required:    true,
			},
		},
	},
	{
		Name:        "ccd",
		Description: "Crypto Daily Chart",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "coin",
				Description: "coin to get the Daily chart of",
				Required:    true,
			},
		},
	},
}
