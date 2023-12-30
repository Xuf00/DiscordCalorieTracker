package main

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

var s *discordgo.Session

var (
	minCalorieIntake = 1.0
	maxItemCalories  = 5000.0

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
		/*{
			Name:        "average",
			Description: "Gives you an average of your calories consumed over the provided days",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionNumber,
					Name:        "days",
					Description: "The amount of days to calculate the average for",
					MaxValue:    7,
					Required:    true,
				},
			},
		}, */
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
			},
		},
		{
			Name:        "list",
			Description: "List the days log entries",
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
			},
		},
		{
			Name:        "conv",
			Description: "Figure out actual calories consumed when only given per X grams",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionNumber,
					Name:        "grams",
					Description: "The per X grams given on the nutrition label",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionNumber,
					Name:        "calories",
					Description: "The amount of calories for the grams specified",
					Required:    true,
					MinValue:    &minCalorieIntake,
					MaxValue:    maxItemCalories,
				},
				{
					Type:        discordgo.ApplicationCommandOptionNumber,
					Name:        "weight",
					Description: "The actual weight in grams of the packet",
					Required:    true,
				},
			},
		},
	}

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"set":    HandleSetCommand,
		"add":    HandleAddCommand,
		"update": HandleUpdateCommand,
		"del":    HandleDeleteCommand,
		"conv":   HandleConvCommand,
		"list":   HandleListCommand,
		"rem":    HandleRemCommand,
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
	if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
		h(s, i)
	}
}

func CreateInteractionResponse(content string) *discordgo.InteractionResponse {
	interactionResponse := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
		},
	}
	return interactionResponse
}
