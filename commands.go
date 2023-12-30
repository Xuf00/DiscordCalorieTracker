package main

import (
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

func convertOptionsToMap(options []*discordgo.ApplicationCommandInteractionDataOption) map[string]*discordgo.ApplicationCommandInteractionDataOption {
	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
	for _, opt := range options {
		optionMap[opt.Name] = opt
	}
	return optionMap
}

func HandleSetCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Access options in the order provided by the user.
	options := i.ApplicationCommandData().Options

	// Convert the slice into a map
	optionMap := convertOptionsToMap(options)

	calories := optionMap["calories"].IntValue()
	userId := i.Member.User.ID

	user := User{
		id:             userId,
		daily_calories: int16(calories),
	}

	var response string

	_, err := SetUserCalories(&user)

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
}

func HandleAddCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	var response string
	userId := i.Member.User.ID
	userDisplayName := i.Member.User.GlobalName

	user, err := FetchUserByID(userId)
	if err != nil {
		response = fmt.Sprintf("Encountered an error: %v", err)
	} else if (User{}) == user {
		response = "Set your daily calories first using the /set command."
	} else {
		// Access options in the order provided by the user.
		options := i.ApplicationCommandData().Options

		optionMap := convertOptionsToMap(options)

		foodItem := optionMap["fooditem"].StringValue()
		calories := optionMap["calories"].IntValue()

		foodLog := FoodLog{
			user_id:   userId,
			food_item: foodItem,
			calories:  int16(calories),
		}

		id, err := AddUserFoodLog(&foodLog)

		if err != nil {
			response = "There was an error, please try again..."
		} else {
			response = fmt.Sprintf("Successfully added %v to your daily log for %d calories and with ID %d.", foodItem, calories, id)
		}

		remaining, err := FetchRemainingCalories(userId)
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
}

func HandleUpdateCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	userId := i.Member.User.ID
	userDisplayName := i.Member.User.GlobalName

	// Access options in the order provided by the user.
	options := i.ApplicationCommandData().Options

	optionMap := convertOptionsToMap(options)

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

	n, err := UpdateUserFoodLog(&foodLog)
	if err != nil {
		response = "There was an error, please try again..."
	} else if n == 0 {
		response = fmt.Sprintf("Could not find a food log with ID %v.", logId)
	} else {
		response = fmt.Sprintf("Successfully updated food log with ID %v.", logId)

		remaining, err := FetchRemainingCalories(userId)
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
}

func HandleDeleteCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	userId := i.Member.User.ID

	// Access options in the order provided by the user.
	options := i.ApplicationCommandData().Options

	// Convert the slice into a map
	optionMap := convertOptionsToMap(options)

	logId := optionMap["logid"].IntValue()

	n, err := DeleteUserFoodLog(userId, logId)

	var response string

	if err != nil {
		response = "There was an error, please try again..."
	} else if n == 0 {
		response = fmt.Sprintf("Could not find a food log with ID %v.", logId)
	} else {
		response = fmt.Sprintf("Successfully deleted food log with the ID %v.", logId)

		remaining, err := FetchRemainingCalories(userId)
		if err != nil {
			response = "There was an error, please try again..."
		}

		response = response + fmt.Sprintf("\nYou have %d calories remaining today.", remaining)
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: response,
		},
	})
}

func HandleConvCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Access options in the order provided by the user.
	options := i.ApplicationCommandData().Options

	// Convert the slice into a map
	optionMap := convertOptionsToMap(options)

	calories := optionMap["calories"].FloatValue()
	grams := optionMap["grams"].FloatValue()
	weight := optionMap["weight"].FloatValue()

	perGram := calories / grams
	totalCalories := perGram * weight

	s.InteractionRespond(i.Interaction, CreateInteractionResponse(fmt.Sprintf("%.2f calories per gram \nTotal amount of calories is %.0f", perGram, math.Ceil(totalCalories))))
}

func HandleListCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	userId := i.Member.User.ID
	userDisplayName := i.Member.User.GlobalName

	log.Printf("Fetching daily food logs for user %v.", userDisplayName)
	foodLogs, foodLogErr := FetchDailyFoodLogs(userId)
	if foodLogErr != nil {
		log.Printf("Error fetching food logs for user %v: %v", userDisplayName, foodLogErr)
		s.InteractionRespond(i.Interaction, CreateInteractionResponse("Error fetching food logs, please try again..."))
		return
	}

	if len(foodLogs) == 0 {
		log.Printf("User %v currently has no logs.", userDisplayName)
		s.InteractionRespond(i.Interaction, CreateInteractionResponse("No logs found for you today."))
		return
	}

	consumed, consumedErr := FetchConsumedCalories(userId)
	remaining, remainingErr := FetchRemainingCalories(userId)
	if consumedErr != nil || remainingErr != nil {
		log.Printf("Error fetching consumed or remaining calories for user %v.", userDisplayName)
		s.InteractionRespond(i.Interaction, CreateInteractionResponse("Error fetching consumed or remaining calories, please try again..."))
		return
	}

	embed := createFoodLogEmbed(userDisplayName, foodLogs, consumed, remaining)
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				embed,
			},
		},
	})
}

func HandleRemCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	userId := i.Member.User.ID
	userDisplayName := i.Member.User.GlobalName

	log.Printf("Fetching remaining calories for user %v.", userDisplayName)
	remaining, remainingErr := FetchRemainingCalories(userId)
	if remainingErr != nil {
		log.Printf("Error fetching remaining calories for user %v.", userDisplayName)
		s.InteractionRespond(i.Interaction, CreateInteractionResponse("Error fetching remaining calories, please try again..."))
		return
	}

	if remaining == 10000 {
		log.Printf("Couldn't get the remaining calories for user %v.", userDisplayName)
		s.InteractionRespond(i.Interaction, CreateInteractionResponse("Couldn't work out your remaining calories, make sure you've used /set to set your daily calories."))
		return
	}

	s.InteractionRespond(i.Interaction, CreateInteractionResponse(fmt.Sprintf("You have %d calories remaining today.", remaining)))
}

func createFoodLogEmbed(username string, foodLogs []FoodLog, consumed int64, remaining int64) *discordgo.MessageEmbed {
	var foodItemNames strings.Builder
	var calories strings.Builder
	var times strings.Builder

	for _, foodLog := range foodLogs {
		foodItemNames.WriteString(fmt.Sprintf("(%d) %s\n", foodLog.id, foodLog.food_item))
		calories.WriteString(fmt.Sprintf("%d\n", foodLog.calories))
		times.WriteString(fmt.Sprintf("%s\n", foodLog.date_time.Format("15:04")))
	}

	now := time.Now()

	embed := &discordgo.MessageEmbed{
		Title:  fmt.Sprintf("Food Log - %s (%s)", username, now.Format("02/01/2006")),
		Author: &discordgo.MessageEmbedAuthor{},
		Color:  0x89CFF0,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Time",
				Value:  times.String(),
				Inline: true,
			},
			{
				Name:   "Name",
				Value:  foodItemNames.String(),
				Inline: true,
			},
			{
				Name:   "Calories",
				Value:  calories.String(),
				Inline: true,
			},
			{
				Name:   "Calories Consumed",
				Value:  strconv.FormatInt(consumed, 10),
				Inline: true,
			},
			{
				Name:   "Calories Remaining",
				Value:  strconv.FormatInt(remaining, 10),
				Inline: true,
			},
		},
		Timestamp: now.Format(time.RFC3339),
	}

	return embed
}
