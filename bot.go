package main

import (
	"database/sql"
	"fmt"
	"log"
	"nevolution/dab"
	"os"
	"os/signal"
	"strconv"
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

var dg *discordgo.Session

func biome(s *discordgo.Session, i *discordgo.InteractionCreate) {
	rows := strings.Split(i.MessageComponentData().CustomID, "|")
	d.AddBiome(rows[1], rows[2])
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("Успешно добавлен биом %v типа %v", rows[1], rows[2]),
		},
	})
	if err != nil {
		log.Fatal(err)
	}
}
func mutation(s *discordgo.Session, i *discordgo.InteractionCreate) {
	rows := strings.Split(i.MessageComponentData().CustomID, "|")
	bid, err := strconv.Atoi(rows[4])
	if err != nil {
		log.Fatal(err)
	}
	gid, err := strconv.Atoi(rows[2])
	if err != nil {
		log.Fatal(err)
	}
	d.StartMutation(bid, gid, rows[3])
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("Начато исследование мутации %v в граде %v", rows[2], rows[5]),
		},
	})
	if err != nil {
		log.Fatal(err)
	}
}
func mutButton(gid int, mut string, bid int, grade string) (button discordgo.Button) {
	return discordgo.Button{
		Emoji: discordgo.ComponentEmoji{
			Name: "🧬",
		},
		Label:    mut,
		Style:    discordgo.PrimaryButton,
		CustomID: "newm@|" + fmt.Sprint(gid) + "|" + mut + "|" + fmt.Sprint(bid) + "|" + grade,
	}

}

var (
	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "rename-grade",
			Description: "переименовывает граду",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "grade_name",
					Description: "текущее Название грады",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "new_name",
					Description: "новое название грады",
					Required:    true,
				},
			},
		},
		{
			Name:        "rollback",
			Description: "откатывает",
		},
		{
			Name:        "startup",
			Description: "запускает новую игру",
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
				{Type: discordgo.ApplicationCommandOptionString,
					Name:        "source_biome",
					Description: "Биом-источник",
					Required:    true,
				},
			},
		},
		{
			Name:        "grade-info",
			Description: "Gives grade info",
			Options: []*discordgo.ApplicationCommandOption{
				{Type: discordgo.ApplicationCommandOptionString,
					Name:        "grade_name",
					Description: "Название грады",
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
				{Type: discordgo.ApplicationCommandOptionString,
					Name:        "biome_name",
					Description: "Название биома",
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
		"startup": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			d.StartNew()
			d.DebugRemoveLater()
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "# успешно запущено",
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
					b.WriteString("# спешно добавлена града ")
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
			if biome, ok := optionMap["biome_name"]; ok {
				_, status := d.BiomeID(biome.StringValue())
				if status == 2 {
					err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: fmt.Sprintf("Выберите тип биома %v.", biome.StringValue()),
							Flags:   discordgo.MessageFlagsEphemeral,
							Components: []discordgo.MessageComponent{
								discordgo.ActionsRow{
									Components: []discordgo.MessageComponent{
										discordgo.Button{
											Emoji: discordgo.ComponentEmoji{
												Name: "♨️",
											},
											Label:    "Гейзеры",
											Style:    discordgo.PrimaryButton,
											CustomID: "newb@|" + biome.StringValue() + "|" + "geysers",
										},
										discordgo.Button{
											Emoji: discordgo.ComponentEmoji{
												Name: "🚬",
											},
											Label:    "Курильщики",
											Style:    discordgo.PrimaryButton,
											CustomID: "newb@|" + biome.StringValue() + "|" + "smokers",
										},
										discordgo.Button{
											Emoji: discordgo.ComponentEmoji{
												Name: "🌊",
											},
											Label:    "Пелагиаль",
											Style:    discordgo.PrimaryButton,
											CustomID: "newb@|" + biome.StringValue() + "|" + "pelagial",
										},
										discordgo.Button{
											Emoji: discordgo.ComponentEmoji{
												Name: "💧",
											},
											Label:    "Пресные воды",
											Style:    discordgo.PrimaryButton,
											CustomID: "newb@|" + biome.StringValue() + "|" + "freshwater",
										},
									},
								},
								discordgo.ActionsRow{
									Components: []discordgo.MessageComponent{
										discordgo.Button{
											Emoji: discordgo.ComponentEmoji{
												Name: "🪨",
											},
											Label:    "Эндолиты",
											Style:    discordgo.PrimaryButton,
											CustomID: "newb@|" + biome.StringValue() + "|" + "endolytes",
										},
										discordgo.Button{
											Emoji: discordgo.ComponentEmoji{
												Name: "☁️",
											},
											Label:    "Атмосфера",
											Style:    discordgo.PrimaryButton,
											CustomID: "newb@|" + biome.StringValue() + "|" + "atmosphere",
										},
										discordgo.Button{
											Emoji: discordgo.ComponentEmoji{
												Name: "🌀",
											},
											Label:    "Литораль",
											Style:    discordgo.PrimaryButton,
											CustomID: "newb@|" + biome.StringValue() + "|" + "littoral",
										},
									},
								},
							},
						},
					})
					if err != nil {
						log.Fatal(err)
					}
				}

			}
		},
		"grade-to-biome": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			options := i.ApplicationCommandData().Options
			optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
			for _, opt := range options {
				optionMap[opt.Name] = opt
			}
			var r strings.Builder
			if biome, ok := optionMap["biome_name"]; ok {
				_, status := d.BiomeID(biome.StringValue())
				if status == 2 {
					r.WriteString(fmt.Sprintf("Биома %v нет", biome.StringValue()))
				} else if status == 1 {
					if grade, ok := optionMap["grade_name"]; ok {
						gid, status := d.GradeID(grade.StringValue())
						if status == 2 {
							r.WriteString(fmt.Sprintf("Грады %v нет", grade.StringValue()))
						} else if status == 1 {
							if source, ok := optionMap["source_biome"]; ok {
								sid, status := d.BiomeID(source.StringValue())
								if status == 2 {
									r.WriteString(fmt.Sprintf("Биома %v нет", source.StringValue()))
								} else if status == 1 {
									amount, status := d.GradeAmount(gid, sid)
									if status == 2 {
										r.WriteString(fmt.Sprintf("Грады %v нет в биоме %v", grade.StringValue(), source.StringValue()))
									} else if status == 1 {
										status = d.AddGradeToBiome(biome.StringValue(), grade.StringValue(), amount/3)
										if status == 4 {
											r.WriteString(fmt.Sprintf("Града %v уже есть в биоме %v", grade.StringValue(), biome.StringValue()))
										} else if status == 1 {
											r.WriteString(fmt.Sprintf("Града %v колонизирует биом %v из биома %v в количестве %v особей", grade.StringValue(), biome.StringValue(), source.StringValue(), amount/3))
										}
									}
								}
							}
						}
					}
				}
			}
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: r.String(),
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
		"grade-info": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			options := i.ApplicationCommandData().Options
			optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
			for _, opt := range options {
				optionMap[opt.Name] = opt
			}
			var r strings.Builder
			if name, ok := optionMap["grade_name"]; ok {
				str, status := d.GetGradeInfo(name.StringValue())
				if status == 2 {
					r.WriteString(fmt.Sprintf("Грады %v не существует", name.StringValue()))
				} else if status == 1 {
					r.WriteString(str)
				}
			}
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: r.String(),
				},
			})
		},
		"rename-grade": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			options := i.ApplicationCommandData().Options
			optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
			for _, opt := range options {
				optionMap[opt.Name] = opt
			}
			if gr_old, ok := optionMap["grade_name"]; ok {
				golid, status := d.GradeID(gr_old.StringValue())
				if status == 1 {
					if gr_new, ok := optionMap["new_name"]; ok {
						status := d.RenameGrade(golid, gr_new.StringValue())
						if status == 1 {
							s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
								Type: discordgo.InteractionResponseChannelMessageWithSource,
								Data: &discordgo.InteractionResponseData{
									Content: fmt.Sprintf("Града %v успешно переименована в %v", gr_old.StringValue(), gr_new.StringValue()),
								},
							})
						} else {
							s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
								Type: discordgo.InteractionResponseChannelMessageWithSource,
								Data: &discordgo.InteractionResponseData{
									Content: fmt.Sprintf("Града %v уже существует", gr_new.StringValue()),
								},
							})
						}
					}
				} else {
					s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: fmt.Sprintf("Грады %v не существует", gr_old.StringValue()),
						},
					})
				}
			}
		},
		"new-mutation": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			options := i.ApplicationCommandData().Options
			optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
			for _, opt := range options {
				optionMap[opt.Name] = opt
			}
			var absentMutations []string
			if biome, ok := optionMap["biome_name"]; ok {
				bid, status := d.BiomeID(biome.StringValue())
				if status == 1 {
					if option, ok := optionMap["grade_name"]; ok {
						b, gid, status := d.GetGradeMutations(option.StringValue())
						if status == 1 {
							for key := range allMutations {
								if _, ok := b[key]; !ok {
									absentMutations = append(absentMutations, key)
								}
							}
							var cmp []discordgo.MessageComponent
							if len(absentMutations) > 5 {
								var overarchingArray [][]string
								c := (len(absentMutations) / 5) + 1
								for i := 0; i < len(absentMutations)-c; {
									overarchingArray = append(overarchingArray, absentMutations[i:i+c])
									i = i + c
								}
								for _, v := range overarchingArray {
									var comp []discordgo.MessageComponent
									for _, v2 := range v {
										comp = append(comp, mutButton(gid, v2, bid, option.StringValue()))
									}
									cmp = append(cmp, discordgo.ActionsRow{Components: comp})
								}
							} else {
								for _, v := range absentMutations {
									cmp = append(cmp, mutButton(gid, v, bid, option.StringValue()))
								}
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
								log.Fatal(err)
							}
						} else {
							s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
								Type: discordgo.InteractionResponseChannelMessageWithSource,
								Data: &discordgo.InteractionResponseData{
									Content: "произошла какая-то хрень",
								},
							})
						}
					}
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

func main() {
	dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
				h(s, i)
			}
		case discordgo.InteractionMessageComponent:
			if strings.Contains(i.MessageComponentData().CustomID, "|") && strings.Contains(i.MessageComponentData().CustomID, "newb@") {
				biome(s, i)
			} else if strings.Contains(i.MessageComponentData().CustomID, "|") && strings.Contains(i.MessageComponentData().CustomID, "newm@") {
				mutation(s, i)
			}
		}
	})
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
	fmt.Println("\n Bot running.")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	fmt.Println("\n removing commands...")
	for _, v := range registeredCommands {
		err := dg.ApplicationCommandDelete(dg.State.User.ID, info.GuildID, v.ID)
		if err != nil {
			log.Panicf("\n cannot delete '%v' command: %v", v.Name, err)
		}
	}
	fmt.Println("\n shutdown.")
}
