package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	discord "github.com/bwmarrin/discordgo"

	"github.com/superoo7/statbot2/command"
	"github.com/superoo7/statbot2/command/sh"
	"github.com/superoo7/statbot2/command/steem"
	"github.com/superoo7/statbot2/config"
	d "github.com/superoo7/statbot2/discord"
)

func main() {
	env := os.Getenv("ENV")
	bot := d.Discord

	// Register the messageCreate func as a callback for MessageCreate events.
	bot.AddHandlerOnce(botReady)
	bot.AddHandler(func(s *discord.Session, m *discord.MessageCreate) {
		messageCreate(s, m, d.DiscordEmbedMessageChannel, d.DiscordMessageChannel)
	})
	err := bot.Open()

	go d.ProcessEmbedMessage(d.DiscordEmbedMessageChannel)
	go d.ProcessMessage(d.DiscordMessageChannel)

	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	if env == "development" {
		fmt.Println("Press CTRL-C to exit.")
		sc := make(chan os.Signal, 1)
		signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
		<-sc

		// Cleanly close down the Discord session.
		bot.Close()
	} else {
		defer bot.Close()
		<-make(chan struct{})
	}
}

func botReady(s *discord.Session, r *discord.Ready) {
	go d.UpdateSession(s)
	fmt.Println("Bot is running.")
	s.UpdateStatus(0, "Statbot V2 %help to get started")
}

func messageCreate(s *discord.Session, m *discord.MessageCreate, emc chan<- d.DiscordEmbedMessage, mc chan<- d.DiscordMessage) {
	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	// filter whitelist channel
	exit := true
	for _, cid := range config.Whitelist {
		if m.ChannelID == cid {
			exit = false
			break
		}
	}
	if exit {
		return
	}

	// Update session struct
	go d.UpdateSession(s)

	trigger := string(m.Content[0])
	args := strings.Fields(m.Content[1:])

	if trigger == "$" {
		if len(args) < 1 {
			return
		}
		coin := args[0]
		command.PriceCommand(coin, m, emc)
	} else if trigger == "%" {
		if len(args) < 1 {
			return
		}
		switch args[0] {
		case "chart":
			if len(args) >= 2 {
				coin := args[1]
				command.ChartCommand(coin, m, emc)
			} else {
				em := d.GenErrorMessage("Invalid command, try `%price <coin>`")
				emc <- d.DiscordEmbedMessage{CID: m.ChannelID, Message: em}
			}
			break
		case "p", "price":
			if len(args) >= 2 {
				coin := args[1]
				command.PriceCommand(coin, m, emc)
			} else {
				em := d.GenErrorMessage("Invalid command, try `%price <coin>`")
				emc <- d.DiscordEmbedMessage{CID: m.ChannelID, Message: em}
			}
			break
		case "ping":
			command.PingCommand(m, emc)
			break
		case "discord":
			msg := d.GenSimpleEmbed(d.Blue, "Join our discord channel")
			emc <- d.DiscordEmbedMessage{CID: m.ChannelID, Message: msg}
			mc <- d.DiscordMessage{CID: m.ChannelID, Message: "https://discord.gg/J99vTUS"}
			break
		case "bug", "bugs":
			msg := d.GenSimpleEmbed(d.Blue, "Report bugs at https://github.com/superoo7/statbot2/issues")
			emc <- d.DiscordEmbedMessage{CID: m.ChannelID, Message: msg}
			break
		case "donate", "donation":
			emc <- d.DiscordEmbedMessage{
				CID: m.ChannelID,
				Message: d.GenMultipleEmbed(
					d.Blue,
					"Donation to make this project sustainable",
					[]*discord.MessageEmbedField{
						&discord.MessageEmbedField{
							Name:   "I accept following cryptocurrencies, feel free to donate.",
							Value:  "can drop me an email or discord message after you had done.",
							Inline: false,
						},
						&discord.MessageEmbedField{
							Name:   "Steem",
							Value:  "[@superoo7](https://steemitwallet.com/@superoo7)",
							Inline: false,
						},
						&discord.MessageEmbedField{
							Name:   "Ethereum/Hunt-Token/ERC20 tokens",
							Value:  "[superoo7.eth](https://etherscan.io/address/superoo7.eth) (ENS Address) or [0xfCAD3475520fb54Fc95305A6549A79170DA8B7C0](https://etherscan.io/address/0xfCAD3475520fb54Fc95305A6549A79170DA8B7C0)",
							Inline: false,
						},
						&discord.MessageEmbedField{
							Name:   "Bitcoin",
							Value:  "[3QvrngBgkwT7x1ybUcLNSCPZrooiErVgfZ](https://www.blockchain.com/btc/address/3QvrngBgkwT7x1ybUcLNSCPZrooiErVgfZ)",
							Inline: false,
						},
					},
				),
			}
			break
		case "h", "help":
			emc <- d.DiscordEmbedMessage{
				CID: m.ChannelID,
				Message: d.GenMultipleEmbed(
					d.Blue,
					fmt.Sprintf("Help Message (%s)", config.Version),
					[]*discord.MessageEmbedField{
						&discord.MessageEmbedField{
							Name:   "`$<coin>` , `%price <coin>`, `%p <coin>`",
							Value:  "for checking cryptocurrency price",
							Inline: false,
						},
						&discord.MessageEmbedField{
							Name:   "`%convert <amount> <crypto/fiat> <crypto/fiat>`",
							Value:  "Convert crypto->crypto, crypto->fiat, fiat->crypto. (Does not support fiat->fiat",
							Inline: false,
						},
						&discord.MessageEmbedField{
							Name:   "`%chart <coin>`",
							Value:  "To get cryptocurrency chart of a certain coin",
							Inline: false,
						},
						&discord.MessageEmbedField{
							Name:   "`%daily`",
							Value:  "Get daily top 12 cryptocurrencies prices",
							Inline: false,
						},
						&discord.MessageEmbedField{
							Name:   "`%delegate <from> <to> <amount>` or `%delegate <to> <amount>`",
							Value:  "Delegate to a person with steemconnect",
							Inline: false,
						},
						// &discord.MessageEmbedField{
						// 	Name:   "`%sf` or `%steemfest`",
						// 	Value:  "Count down to SteemFest!",
						// 	Inline: false,
						// },
						&discord.MessageEmbedField{
							Name:   "`%hunt <steemhunt link>` or `%sh <steemhunt link`",
							Value:  "Get hunt post details",
							Inline: false,
						},
						&discord.MessageEmbedField{
							Name:   "`%donate`",
							Value:  "To get details to donate to me",
							Inline: false,
						},
						&discord.MessageEmbedField{
							Name:   "`%discord`",
							Value:  "to join our discord!",
							Inline: false,
						},
						&discord.MessageEmbedField{
							Name:   "`%help` , `%h`",
							Value:  "for help",
							Inline: false,
						},
					},
				),
			}
			break
		case "s", "steem":
			steem.SteemCommand(m.ChannelID, args, emc, mc)
			break
		case "s/sbd":
			command.ConvertCommand(m, emc, []string{"1", "steem", "sbd"})
			break
		case "sbd/s":
			command.ConvertCommand(m, emc, []string{"1", "sbd", "steem"})
			break
		case "delegate":
			if len(args) > 1 {
				steem.DelegateCommand(m.ChannelID, emc, args[1:])
			} else {
				emc <- d.DiscordEmbedMessage{
					CID:     m.ChannelID,
					Message: d.GenErrorMessage("Invalid command, try `%delegate <from> <to> <no of sp>`"),
				}
			}
			break
		// case "sf", "steemfest":
		// 	steem.SteemFestCommand(m, emc)
		// 	break
		case "daily":
			command.DailyCommand(m, emc)
			break
		case "convert":
			command.ConvertCommand(m, emc, args[1:])
			break
		case "sh", "hunt", "steemhunt":
			if len(args) >= 2 {
				sh.HuntCommand(m, emc, args[1])
			}
			break
		default:
			emc <- d.DiscordEmbedMessage{
				CID:     m.ChannelID,
				Message: d.GenErrorMessage("Invalid command, Try `%help` to get started"),
			}
			break
		}
	}
}
