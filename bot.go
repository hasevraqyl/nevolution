package main

import (
	"database/sql"
	"fmt"
	"log"
	"nevolution/dab"
	"os"
	"os/signal"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/bwmarrin/discordgo"
)

type Init struct {
	Token   string
	GuildID string
}

var info Init
var d dab.Database

func init() {
	f := "setup_secret.toml"
	if _, err := os.Stat(f); err != nil {
		log.Fatal(err)
	}
	if _, err := toml.DecodeFile(f, &info); err != nil {
		log.Fatal(err)
	}
	db, err := sql.Open("sqlite3", "nev.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	d = dab.Wrap(db)
}

var dg *discordgo.Session
var (
	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "rollback",
			Description: "roll back",
		},
		{
			Name:        "add-grade",
			Description: "Command for adding grades",
			Options: []*discordgo.ApplicationCommandOption{

				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "add-grade",
					Description: "Добавить граду",
					Required:    true,
				},
			},
		},
		{
			Name:        "add-biome",
			Description: "Command for adding biomes",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "biome_name",
					Description: "Название биома",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "biome_type",
					Description: "Тип биома",
					Required:    true,
				},
			},
		},
	}
	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"rollback": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			d.Rollback()
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "# успешно откачено",
				},
			})
		},
		"add-grade": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			options := i.ApplicationCommandData().Options
			optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
			for _, opt := range options {
				optionMap[opt.Name] = opt
			}
			var b strings.Builder
			b.WriteString("placeholder:\n")

			if option, ok := optionMap["string-option"]; ok {
				status := d.AddGrade(option.StringValue())
				b.WriteString(status.Text(option.StringValue()))
			}
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: b.String(),
				},
			})
		},
		"add-biome": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			options := i.ApplicationCommandData().Options
			optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
			for _, opt := range options {
				optionMap[opt.Name] = opt
			}
			var b string
			var b2 string
			var textString string
			if option, ok := optionMap["biome_type"]; ok {
				status := d.AddBiomePreliminary(option.StringValue())
				textString = status.Text(option.StringValue())
				b = textString
				if textString == option.StringValue() {
					if option, ok := optionMap["biome_name"]; ok {
						status := d.AddBiome(option.StringValue(), b)
						b2 = (status.Text(option.StringValue()))
					}
				}
			}
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("Касательно типа: %v, касательно биома: %v", b, b2),
				},
			})
		},
	}
)

func init() {
	var err error
	dg, err = discordgo.New("Bot " + info.Token)
	if err != nil {
		log.Printf("error creating session %v", err)
		return
	}
}

func init() {
	dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})
}

func main() {
	dg.Identify.Intents = discordgo.IntentsGuildMessages
	err := dg.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer dg.Close()
	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
	for i, v := range commands {
		cmd, err := dg.ApplicationCommandCreate(dg.State.User.ID, info.GuildID, v)
		if err != nil {
			log.Fatal(err)
		}
		registeredCommands[i] = cmd
	}
	if err != nil {
		log.Fatal(err)
		return
	}
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	fmt.Println("removing commands...")
	for _, v := range registeredCommands {
		err := dg.ApplicationCommandDelete(dg.State.User.ID, info.GuildID, v.ID)
		if err != nil {
			log.Panicf("cannot delete '%v' command: %v", v.Name, err)
		}
	}
	fmt.Println("shutdown.")
}
