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
var allMutations map[string]struct{}

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
	d = dab.Wrap(db)
	allMutations = make(map[string]struct{})
	allMutations["kill"] = struct{}{}
	allMutations["eat"] = struct{}{}
}

// opwdckijmiipweojwejoi
var dg *discordgo.Session

var (
	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "rollback",
			Description: "откатывает",
		},
		{
			Name:        "turn",
			Description: "делает ход",
		},
		{
			Name:        "meteor",
			Description: "бум бабах",
		},
		{
			Name:        "add-grade",
			Description: "Command for adding grades",
			Options: []*discordgo.ApplicationCommandOption{

				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "grade_name",
					Description: "Имя грады",
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
		{
			Name:        "grade-to-biome",
			Description: "Command for adding grades to biomes",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "grade_name",
					Description: "Название грады",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "biome_name",
					Description: "Название биома",
					Required:    true,
				},
			},
		},
		{
			Name:        "new-mutation",
			Description: "Command for adding mutations to grades",
			Options: []*discordgo.ApplicationCommandOption{
				{Type: discordgo.ApplicationCommandOptionString,
					Name:        "grade_name",
					Description: "Название грады",
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
			if option, ok := optionMap["grade_name"]; ok {
				status := d.AddGrade(option.StringValue())
				if status == 1 {
					b.WriteString("Grade added:\n")
				}
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
		"grade-to-biome": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			options := i.ApplicationCommandData().Options
			optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
			for _, opt := range options {
				optionMap[opt.Name] = opt
			}
			var b1 string
			var b2 string
			var textString string
			if option, ok := optionMap["biome_name"]; ok {
				status := d.AddGradeToBiomePreliminary(option.StringValue())
				textString = status.Text(option.StringValue())
				b1 = textString
				if textString == option.StringValue() {
					if option, ok := optionMap["grade_name"]; ok {
						status := d.AddGradeToBiome(option.StringValue(), b1)
						b2 = (status.Text(option.StringValue()))
					}
				}
			}
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("Касательно типа: %v, касательно биома: %v", b1, b2),
				},
			})
		},
		"meteor": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			d.Meteor()
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "# бомбануло",
				},
			})
		},
		"turn": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			d.Turn()
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "# ход сделан",
				},
			})
		},
		"new-mutation": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			options := i.ApplicationCommandData().Options
			optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
			for _, opt := range options {
				optionMap[opt.Name] = opt
			}
			var absentMutations []string
			if option, ok := optionMap["grade_name"]; ok {
				b, status := d.GetGradeMutations(option.StringValue())
				if status == 1 {
					for key := range allMutations {
						if _, ok := b[key]; !ok {
							absentMutations = append(absentMutations, key)
						}
					}
					var cmp []discordgo.MessageComponent
					for _, v := range absentMutations {
						cmp = append(cmp, discordgo.Button{
							Emoji: discordgo.ComponentEmoji{
								Name: "😭",
							},
							Label:    v,
							Style:    discordgo.PrimaryButton,
							CustomID: option.StringValue() + "|" + v,
						})
					}

					err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: fmt.Sprintf("блядь поггль это пездец, два часа над этим корчусь. энивей, следующие мутации доступны для грады %v, выберите одну", option.StringValue()),
							Flags:   discordgo.MessageFlagsEphemeral,
							Components: []discordgo.MessageComponent{
								discordgo.ActionsRow{
									Components: cmp,
								},
							},
						},
					})
					if err != nil {
						fmt.Println("so this is where the problem lies...")
						log.Fatal(err)
					}
					buttonHandlers := make(map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate))
					for _, v := range absentMutations {
						buttonHandlers[option.StringValue()+"|"+v] = func(s *discordgo.Session, i *discordgo.InteractionCreate) {
							d.StartMutation(option.StringValue(), v)
							err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
								Type: discordgo.InteractionResponseChannelMessageWithSource,
								Data: &discordgo.InteractionResponseData{
									Content: fmt.Sprintf("Начато исследование мутации %v в граде %v", v, option.StringValue()),
								},
							})
							if err != nil {
								fmt.Println("so this is where the problem lies... version 2")
								log.Fatal(err)
							}
						}
					}
					dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
						if h, ok := buttonHandlers[i.ApplicationCommandData().Name]; ok {
							h(s, i)
						}
					})
				} else {
					s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: "произошла какая-то хрень",
						},
					})
				}
			}
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
	defer d.CloseDB()
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
