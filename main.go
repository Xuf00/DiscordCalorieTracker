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
		"set": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			// Access options in the order provided by the user.
			options := i.ApplicationCommandData().Options

			// Convert the slice into a map
			optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
			for _, opt := range options {
				optionMap[opt.Name] = opt
			}

			calories := optionMap["calories"].IntValue()
			userId := i.Member.User.ID

			user := User{
				id:             userId,
				daily_calories: int16(calories),
			}

			var response string

			_, err := setUserCalories(&user)

			if err != nil {
				response = "There was an error, please try again..."
			} else {
				response = fmt.Sprintf("Your daily calorie intake has successfully been set to %d.", calories)
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: response,
				},
			})
		},
		"add": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			var response string
			userId := i.Member.User.ID
			userDisplayName := i.Member.User.Username

			user, err := userById(userId)
			if err != nil {
				response = fmt.Sprintf("Encountered an error: %v", err)
			} else if (User{}) == user {
				response = "Set your daily calories first using the /set command."
			} else {
				// Access options in the order provided by the user.
				options := i.ApplicationCommandData().Options

				// Convert the slice into a map
				optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
				for _, opt := range options {
					optionMap[opt.Name] = opt
				}

				foodItem := optionMap["fooditem"].StringValue()
				calories := optionMap["calories"].IntValue()

				foodLog := FoodLog{
					user_id:   userId,
					food_item: foodItem,
					calories:  int16(calories),
				}

				_, err := addUserFoodLog(&foodLog)

				if err != nil {
					response = "There was an error, please try again..."
				} else {
					response = fmt.Sprintf("Successfully added %v to your daily log for %d calories.", foodItem, calories)
				}

				remaining, err := fetchRemainingCalories(userId)
				if err != nil {
					log.Printf("Encountered error when retrieving remaining calories for user %v.", userDisplayName)
				} else {
					response = response + fmt.Sprintf("\nYou have %d calories remaining today.", remaining)
				}
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: response,
				},
			})
		},
		"update": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			userId := i.Member.User.ID

			// Access options in the order provided by the user.
			options := i.ApplicationCommandData().Options

			// Convert the slice into a map
			optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
			for _, opt := range options {
				optionMap[opt.Name] = opt
			}

			logId := optionMap["logid"].IntValue()
			foodItem := optionMap["fooditem"].StringValue()
			calories := optionMap["calories"].IntValue()

			foodLog := FoodLog{
				id:        logId,
				user_id:   userId,
				food_item: foodItem,
				calories:  int16(calories),
			}

			var response string

			n, err := updateUserFoodLog(&foodLog)
			if err != nil {
				response = "There was an error, please try again..."
			} else if n == 0 {
				response = fmt.Sprintf("Could not find a food log with ID %v.", logId)
			} else {
				response = fmt.Sprintf("Successfully updated food log with ID %v.", logId)
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: response,
				},
			})
		},
		"del": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			userId := i.Member.User.ID

			// Access options in the order provided by the user.
			options := i.ApplicationCommandData().Options

			// Convert the slice into a map
			optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
			for _, opt := range options {
				optionMap[opt.Name] = opt
			}

			logId := optionMap["logid"].IntValue()

			n, err := deleteUserFoodLog(userId, logId)

			var response string

			if err != nil {
				response = "There was an error, please try again..."
			} else if n == 0 {
				response = fmt.Sprintf("Could not find a food log with ID %v.", logId)
			} else {
				response = "Successfully deleted the item from your daily log."
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: response,
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

			calories := optionMap["calories"].FloatValue()
			grams := optionMap["grams"].FloatValue()
			weight := optionMap["weight"].FloatValue()

			perGram := calories / grams
			totalCalories := perGram * weight

			response := fmt.Sprintf("%.2f calories per gram \nTotal amount of calories is %.0f", perGram, math.Ceil(totalCalories))

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: response,
				},
			})
		},
		"list": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			userId := i.Member.User.ID
			userDisplayName := i.Member.User.Username

			var response string

			log.Printf("Fetching daily food logs for user %v.", userDisplayName)
			foodLogs, err := fetchDailyFoodLogs(userId)
			if err != nil {
				log.Printf("Encountered error when listing food logs for user %v.", userDisplayName)
				response = "There was an error, please try again..."
			} else if len(foodLogs) == 0 {
				log.Printf("User %v currently has no logs.", userDisplayName)
				response = "Set your daily calories first using the /set command."
			} else {
				for _, foodLog := range foodLogs {
					response = response + fmt.Sprintf("%v %v %v %v\n", foodLog.id, foodLog.food_item, foodLog.calories, foodLog.date_time.Format("02/01/2006 15:04"))
				}
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: response,
				},
			})
		},
		"rem": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			userId := i.Member.User.ID
			userDisplayName := i.Member.User.Username

			var response string

			log.Printf("Fetching remaining calories for user %v.", userDisplayName)
			remaining, err := fetchRemainingCalories(userId)
			if err != nil {
				log.Printf("Encountered error when retrieving remaining calories for user %v.", userDisplayName)
			} else if remaining == 10000 {
				log.Printf("Couldn't get the remaining calories for user %v.", userDisplayName)
				response = "Couldn't work out your remaining calories, make sure you've used /set to set your daily calories."
			} else {
				response = fmt.Sprintf("You have %d calories remaining today.", remaining)
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

func main() {
	flag.Parse()

	initDb()

	initDiscordSession()

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

func initDiscordSession() {
	var err error
	s, err = discordgo.New("Bot " + *BotToken)
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}
	s.AddHandler(onReady)
	s.AddHandler(handleCommands)
}

func onReady(s *discordgo.Session, r *discordgo.Ready) {
	log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
}

func handleCommands(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
		h(s, i)
	}
}
