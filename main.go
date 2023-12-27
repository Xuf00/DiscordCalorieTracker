package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
)

// Bot parameters
var (
	GuildID        = flag.String("guild", "", "Test guild ID. If not passed - bot registers commands globally")
	BotToken       = flag.String("token", "", "Bot access token")
	RemoveCommands = flag.Bool("rmcmd", true, "Remove all commands after shutting down or not")
)

var s *discordgo.Session

func init() { flag.Parse() }

func init() {
	var err error
	s, err = discordgo.New("Bot " + *BotToken)
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}
}

var (
	consumedMinValue = 1.0

	commands = []*discordgo.ApplicationCommand{
		/* {
			Name:        "set",
			Description: "Set your daily calorie intake",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionNumber,
					Name:        "calories",
					Description: "The amount of calories you need to intake daily",
					Required:    true,
				},
			},
		},
		{
			Name:        "update",
			Description: "Update a log entry",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionNumber,
					Name:        "logid",
					Description: "The ID of the log entry",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "label",
					Description: "The name of the food product",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionNumber,
					Name:        "calories",
					Description: "Calories consumed by eating this product",
					Required:    true,
				},
			},
		},
		{
			Name:        "delete",
			Description: "Delete a log entry",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionNumber,
					Name:        "label",
					Description: "The ID of the log entry",
					Required:    true,
				},
			},
		},
		{
			Name:        "list",
			Description: "List the days log entries",
		},
		{
			Name:        "remaining",
			Description: "Gives your remaining calories for the day",
		},
		{
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
		},
		{
			Name:        "add",
			Description: "Add an entry to your daily calories",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "label",
					Description: "The name of the food product",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionNumber,
					Name:        "calories",
					Description: "The amount of calories consumed by eating this product",
					Required:    true,
				},
			},
		}, */
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
				},
				{
					Type:        discordgo.ApplicationCommandOptionNumber,
					Name:        "weight",
					Description: "The actual weight in grams of the packet",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionNumber,
					Name:        "consumed",
					Description: "The amount consumed",
					Required:    false,
					MinValue:    &consumedMinValue,
				},
			},
		},
	}

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"set": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "You requested to set your daily calories.",
				},
			})
		},
		"update": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "You requested to update your daily calories.",
				},
			})
		},
		"conv": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			// Access options in the order provided by the user.
			options := i.ApplicationCommandData().Options

			// Convert the slice into a map
			optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
			for _, opt := range options {
				optionMap[opt.Name] = opt
			}

			getValue := func(key string) float64 {
				if opt, ok := optionMap[key]; ok {
					return opt.FloatValue()
				}
				return 0
			}

			calories := getValue("calories")
			grams := getValue("grams")
			weight := getValue("weight")
			consumed := getValue("consumed")

			perGram := calories / grams
			totalCalories := perGram * weight

			response := fmt.Sprintf("%.2f calories per gram \nTotal amount of calories is %.0f", perGram, math.Ceil(totalCalories))

			if consumed != 0 {
				consumedCalories := perGram * consumed
				response += fmt.Sprintf("\nConsumed %.0f which is %.0f calories", consumed, math.Ceil(consumedCalories))
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: response,
				},
			})
		},
	}
)

func init() {
	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})
}

func main() {
	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})
	err := s.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}

	log.Println("Adding commands...")
	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
	for i, v := range commands {
		cmd, err := s.ApplicationCommandCreate(s.State.User.ID, *GuildID, v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
		registeredCommands[i] = cmd
	}

	defer s.Close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Println("Press Ctrl+C to exit")
	<-stop

	if *RemoveCommands {
		log.Println("Removing commands...")
		for _, v := range registeredCommands {
			err := s.ApplicationCommandDelete(s.State.User.ID, *GuildID, v.ID)
			if err != nil {
				log.Panicf("Cannot delete '%v' command: %v", v.Name, err)
			}
		}
	}

	log.Println("Gracefully shutting down.")
}
