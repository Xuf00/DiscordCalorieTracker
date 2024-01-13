package main

import (
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
)

var s *discordgo.Session

var (
	minCalorieIntake = 1.0
	maxItemCalories  = 5000.0

	minAverageDays = 2.0
	minQuantity    = 1.0

	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "set",
			Description: "Set your daily calorie intake",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "calories",
					Description: "The amount of calories you need to intake daily",
					Required:    true,
					MinValue:    &minCalorieIntake,
					MaxValue:    maxItemCalories,
				},
			},
		},
		{
			Name:        "avg",
			Description: "Gives you an average of your calories consumed over X days",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "days",
					Description: "The amount of days to calculate the average for",
					MinValue:    &minAverageDays,
					MaxValue:    7,
					Required:    true,
				},
			},
		},
		{
			Name:        "rem",
			Description: "Gives your remaining calories for the day",
		},
		{
			Name:        "update",
			Description: "Update a log entry",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "logid",
					Description: "The ID of the log entry",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "fooditem",
					Description: "The name of the food product",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "calories",
					Description: "Calories consumed by eating this product",
					Required:    true,
					MinValue:    &minCalorieIntake,
					MaxValue:    maxItemCalories,
				},
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "quantity",
					Description: "The quantity consumed",
					Required:    false,
					MinValue:    &minQuantity,
				},
			},
		},
		{
			Name:        "list",
			Description: "List log entries for the current day or if the specified day if provided.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "date",
					Description: "The date to list",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "user",
					Description: "Which users list to view",
					Required:    false,
				},
			},
		},
		{
			Name:        "del",
			Description: "Delete a log entry",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "logid",
					Description: "The ID of the log entry",
					Required:    true,
				},
			},
		},
		{
			Name:        "add",
			Description: "Add an entry to your daily calories",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "fooditem",
					Description: "The name of the food product",
					Required:    true,
					MaxLength:   50,
				},
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "calories",
					Description: "The amount of calories consumed by eating this product",
					Required:    true,
					MinValue:    &minCalorieIntake,
					MaxValue:    maxItemCalories,
				},
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "quantity",
					Description: "The quantity consumed",
					Required:    false,
					MinValue:    &minQuantity,
				},
			},
		},
		{
			Name:        "conv",
			Description: "Figure out actual calories consumed when only given per X units",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionNumber,
					Name:        "units",
					Description: "The per X units (e.g grams or ml) given on the nutrition label",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionNumber,
					Name:        "calories",
					Description: "The amount of calories for the unit specified",
					Required:    true,
					MinValue:    &minCalorieIntake,
					MaxValue:    maxItemCalories,
				},
				{
					Type:        discordgo.ApplicationCommandOptionNumber,
					Name:        "weight",
					Description: "The actual weight in units of the packet",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "fooditem",
					Description: "The name of the food product",
					Required:    false,
				},
			},
		},
	}

	componentHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"flquantity": HandleModifyFoodQuantity,
		"fllist":     HandleUpdateList,
	}

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"set":    HandleSetCommand,
		"add":    HandleAddCommand,
		"update": HandleUpdateCommand,
		"del":    HandleDeleteCommand,
		"conv":   HandleConvCommand,
		"list":   HandleListCommand,
		"rem":    HandleRemCommand,
		"avg":    HandleAverageCommand,
	}

	registeredCommands = make([]*discordgo.ApplicationCommand, len(commands))
)

func InitDiscordSession() {
	var err error
	s, err = discordgo.New("Bot " + *BotToken)
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}
	s.AddHandler(onReady)
	s.AddHandler(handleCommands)
}

func OpenDiscordSession() {
	err := s.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}
}

func AddCommandsDiscord() {
	log.Println("Adding commands...")
	for i, v := range commands {
		cmd, err := s.ApplicationCommandCreate(s.State.User.ID, *GuildID, v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
		registeredCommands[i] = cmd
	}
}

func RemoveCommandsDiscord() {
	log.Println("Removing commands...")
	for _, v := range registeredCommands {
		err := s.ApplicationCommandDelete(s.State.User.ID, *GuildID, v.ID)
		if err != nil {
			log.Panicf("Cannot delete '%v' command: %v", v.Name, err)
		}
	}
}

func onReady(s *discordgo.Session, r *discordgo.Ready) {
	log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
}

func handleCommands(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		log.Printf("Handling slash command interaction %v", i.ApplicationCommandData().Name)
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}

	case discordgo.InteractionMessageComponent:
		log.Printf("Handling component interaction %v", i.MessageComponentData().CustomID)
		idPrefix := strings.Split(i.MessageComponentData().CustomID, "_")[0]
		if h, ok := componentHandlers[idPrefix]; ok {
			h(s, i)
		}
	}

}

func CreateInteractionResponse(content string, ephemeral bool, messageComponents []discordgo.MessageComponent) *discordgo.InteractionResponse {
	interactionResponse := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
			/* Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.Button{
							Emoji: discordgo.ComponentEmoji{
								Name: ":arrow_up:",
							},
							Label: "Increase Quantity",
							Style: discordgo.SecondaryButton,
							CustomID: "",
						},
						discordgo.Button{
							Emoji: discordgo.ComponentEmoji{
								Name: "🔧",
							},
							Label: "Discord developers",
							Style: discordgo.LinkButton,
							URL:   "https://discord.gg/discord-developers",
						},
						discordgo.Button{
							Emoji: discordgo.ComponentEmoji{
								Name: "🦫",
							},
							Label: "Discord Gophers",
							Style: discordgo.LinkButton,
							URL:   "https://discord.gg/7RuRrVHyXF",
						},
					},
				},
			}, */
		},
	}

	if ephemeral {
		// Message only shows for the user that triggered it
		interactionResponse.Data.Flags = discordgo.MessageFlagsEphemeral
	}

	if len(messageComponents) > 0 {
		interactionResponse.Data.Components = messageComponents
	}

	return interactionResponse
}
