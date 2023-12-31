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
	userId := i.Member.User.ID
	userDisplayName := i.Member.User.GlobalName

	// Access options in the order provided by the user.
	options := i.ApplicationCommandData().Options

	// Convert the slice into a map
	optionMap := convertOptionsToMap(options)

	calories := optionMap["calories"].IntValue()

	user := User{
		id:             userId,
		daily_calories: int16(calories),
	}

	_, setCaloriesErr := SetUserCalories(&user)
	if setCaloriesErr != nil {
		log.Printf("Error setting calories for user with ID %v and username %v. Error: %v", userId, userDisplayName, setCaloriesErr)
		s.InteractionRespond(i.Interaction, CreateEphemeralInteractionResponse("There was an error, please try again..."))
		return
	}

	log.Printf("Successfully set daily calorie intake to %d for user with ID %v and username %v.", calories, userId, userDisplayName)
	s.InteractionRespond(i.Interaction, CreateEphemeralInteractionResponse(fmt.Sprintf("Your daily calorie intake has successfully been set to %d.", calories)))
}

func HandleAddCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	userId := i.Member.User.ID
	userDisplayName := i.Member.User.GlobalName

	user, userErr := FetchUserByID(userId)
	if userErr != nil {
		log.Printf("Error fetching user with ID %v and username %v. Error: %v", userId, userDisplayName, userErr)
		s.InteractionRespond(i.Interaction, CreateEphemeralInteractionResponse("Error fetching user, please try again..."))
		return
	}

	if (User{}) == user {
		log.Printf("User with ID %v and username %v has tried to add without calling /set first.", userId, userDisplayName)
		s.InteractionRespond(i.Interaction, CreateEphemeralInteractionResponse("Set your daily calories first using the /set command."))
		return
	}

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

	id, addFoodLogErr := AddUserFoodLog(&foodLog)
	if addFoodLogErr != nil {
		log.Printf("Error adding food log for user with ID %v and username %v. Error: %v", userId, userDisplayName, addFoodLogErr)
		s.InteractionRespond(i.Interaction, CreateEphemeralInteractionResponse("There was an error, please try again..."))
		return
	}

	remaining, remainingErr := FetchRemainingCalories(userId)
	if remainingErr != nil {
		log.Printf("Updated the food log but couldn't fetch the remaining calories when updating for user %v.", userDisplayName)
		s.InteractionRespond(i.Interaction, CreateEphemeralInteractionResponse(fmt.Sprintf("Successfully added to your food log with the ID %v.\nCould not retrieve remaining calories at this time.", id)))
		return
	}

	log.Printf("Aded food log %v for user %v and retrieved remaining calories.", id, userDisplayName)
	s.InteractionRespond(i.Interaction, CreateEphemeralInteractionResponse(fmt.Sprintf("Successfully added food log with the ID %v.\nYou have %d calories remaining today.", id, remaining)))
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

	n, updateErr := UpdateUserFoodLog(&foodLog)
	if updateErr != nil {
		log.Printf("Error updating food log with ID %v for user %v: %v", logId, userDisplayName, updateErr)
		s.InteractionRespond(i.Interaction, CreateEphemeralInteractionResponse("There was an error, please try again..."))
		return
	}

	if n == 0 {
		log.Printf("Could not find a food log with ID %v for user %v.", logId, userDisplayName)
		s.InteractionRespond(i.Interaction, CreateEphemeralInteractionResponse(fmt.Sprintf("Could not find a food log with ID %v.", logId)))
		return
	}

	remaining, remainingErr := FetchRemainingCalories(userId)
	if remainingErr != nil {
		log.Printf("Updated the food log but couldn't fetch the remaining calories when updating for user %v.", userDisplayName)
		s.InteractionRespond(i.Interaction, CreateEphemeralInteractionResponse(fmt.Sprintf("Successfully updated food log with the ID %v.\nCould not retrieve remaining calories at this time.", logId)))
		return
	}

	log.Printf("Updated food log %v for user %v and retrieved remaining calories.", logId, userDisplayName)
	s.InteractionRespond(i.Interaction, CreateEphemeralInteractionResponse(fmt.Sprintf("Successfully updated food log with the ID %v.\nYou have %d calories remaining today.", logId, remaining)))
}

func HandleDeleteCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	userId := i.Member.User.ID
	userDisplayName := i.Member.User.GlobalName

	// Access options in the order provided by the user.
	options := i.ApplicationCommandData().Options

	// Convert the slice into a map
	optionMap := convertOptionsToMap(options)

	logId := optionMap["logid"].IntValue()

	n, deleteErr := DeleteUserFoodLog(userId, logId)
	if deleteErr != nil {
		log.Printf("Error deleting food log for user %v: %v", userDisplayName, deleteErr)
		s.InteractionRespond(i.Interaction, CreateEphemeralInteractionResponse("There was an error, please try again..."))
		return
	}

	if n == 0 {
		log.Printf("Could not find a food log with ID %v for user %v.", logId, userDisplayName)
		s.InteractionRespond(i.Interaction, CreateEphemeralInteractionResponse(fmt.Sprintf("Could not find a food log with ID %v.", logId)))
		return
	}

	remaining, remainingErr := FetchRemainingCalories(userId)
	if remainingErr != nil {
		log.Printf("Deleted the food log but couldn't fetch the remaining calories when deleting for user %v.", userDisplayName)
		s.InteractionRespond(i.Interaction, CreateEphemeralInteractionResponse(fmt.Sprintf("Successfully deleted food log with the ID %v.\nCould not retrieve remaining calories at this time.", logId)))
		return
	}

	log.Printf("Deleted food log %v for user %v and retrieved remaining calories.", logId, userDisplayName)
	s.InteractionRespond(i.Interaction, CreateEphemeralInteractionResponse(fmt.Sprintf("Successfully deleted food log with the ID %v.\nYou have %d calories remaining today.", logId, remaining)))
}

func HandleConvCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	userDisplayName := i.Member.User.GlobalName

	// Access options in the order provided by the user.
	options := i.ApplicationCommandData().Options

	// Convert the slice into a map
	optionMap := convertOptionsToMap(options)

	calories := optionMap["calories"].FloatValue()
	grams := optionMap["grams"].FloatValue()
	weight := optionMap["weight"].FloatValue()

	perGram := calories / grams
	totalCalories := perGram * weight

	foodItem, foodItemExists := optionMap["fooditem"]

	if foodItemExists {
		log.Printf("User %v provided the optional food item name when converting.", userDisplayName)
		s.InteractionRespond(i.Interaction, CreateInteractionResponse(fmt.Sprintf("%v\n%.2f calories per gram \nTotal amount of calories is %.0f", foodItem.StringValue(), perGram, math.Ceil(totalCalories))))
	} else {
		log.Printf("User %v did not provide the optional food item name when converting.", userDisplayName)
		s.InteractionRespond(i.Interaction, CreateInteractionResponse(fmt.Sprintf("%.2f calories per gram \nTotal amount of calories is %.0f", perGram, math.Ceil(totalCalories))))
	}
}

func HandleListCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	userId := i.Member.User.ID
	userDisplayName := i.Member.User.GlobalName

	// Access options in the order provided by the user.
	options := i.ApplicationCommandData().Options

	// Convert the slice into a map
	optionMap := convertOptionsToMap(options)

	var dateStr string = time.Now().Format("2006-01-02")
	dateCmd, dateItemExists := optionMap["date"]
	if dateItemExists {
		date, dateParseErr := time.Parse("2006-01-02", dateCmd.StringValue())
		if dateParseErr != nil {
			log.Printf("Error parsing date %v for user with username %v. Error: %v", date, userDisplayName, dateParseErr)
			s.InteractionRespond(i.Interaction, CreateEphemeralInteractionResponse(fmt.Sprintf("Error parsing date, please try again with format %v.", dateStr)))
			return
		}
		dateStr = date.Format("2006-01-02")
	}

	user, userErr := FetchUserByID(userId)
	if userErr != nil {
		log.Printf("Error fetching user with ID %v and username %v. Error: %v", userId, userDisplayName, userErr)
		s.InteractionRespond(i.Interaction, CreateEphemeralInteractionResponse("Error fetching user, please try again..."))
		return
	}

	log.Printf("Fetching daily food logs for user %v.", userDisplayName)
	foodLogs, foodLogErr := FetchDailyFoodLogs(userId, dateStr)
	if foodLogErr != nil {
		log.Printf("Error fetching food logs for user %v: %v", userDisplayName, foodLogErr)
		s.InteractionRespond(i.Interaction, CreateEphemeralInteractionResponse("Error fetching food logs, please try again..."))
		return
	}

	if len(foodLogs) == 0 {
		log.Printf("User %v currently has no logs.", userDisplayName)
		s.InteractionRespond(i.Interaction, CreateEphemeralInteractionResponse("No logs found for you today."))
		return
	}

	consumed, consumedErr := FetchConsumedCalories(userId)
	remaining, remainingErr := FetchRemainingCalories(userId)
	if consumedErr != nil || remainingErr != nil {
		log.Printf("Error fetching consumed or remaining calories for user %v.", userDisplayName)
		s.InteractionRespond(i.Interaction, CreateEphemeralInteractionResponse("Error fetching consumed or remaining calories, please try again..."))
		return
	}

	embed := createFoodLogEmbed(userDisplayName, dateStr, foodLogs, int64(user.daily_calories), consumed, remaining)
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
		s.InteractionRespond(i.Interaction, CreateEphemeralInteractionResponse("Error fetching remaining calories, please try again..."))
		return
	}

	if remaining == 10000 {
		log.Printf("Couldn't get the remaining calories for user %v.", userDisplayName)
		s.InteractionRespond(i.Interaction, CreateEphemeralInteractionResponse("Couldn't work out your remaining calories, make sure you've used /set to set your daily calories."))
		return
	}

	s.InteractionRespond(i.Interaction, CreateEphemeralInteractionResponse(fmt.Sprintf("You have %d calories remaining today.", remaining)))
}

func HandleAverageCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	userId := i.Member.User.ID
	userDisplayName := i.Member.User.GlobalName

	log.Printf("Checking user %v has enough data to get an average.", userDisplayName)
	count, countErr := FetchFoodLogDaysCount(userId)
	if countErr != nil {
		log.Printf("Error checking if the user %v has enough data to get an average.", userDisplayName)
		s.InteractionRespond(i.Interaction, CreateEphemeralInteractionResponse("Error checking your average calories, please try again..."))
		return
	}

	// Access options in the order provided by the user.
	options := i.ApplicationCommandData().Options

	// Convert the slice into a map
	optionMap := convertOptionsToMap(options)

	days := optionMap["days"].IntValue()

	if count != days {
		log.Printf("User %v doesn't have enough data to get an average. They have %d days of data but requested an average for %d.", userDisplayName, count, days)
		s.InteractionRespond(i.Interaction, CreateEphemeralInteractionResponse(fmt.Sprintf("You only have enough data to request an average over %d days.", count)))
		return
	}

	startDate := time.Now().AddDate(0, 0, -int(days)).Format("2006-01-02")
	log.Printf("Fetching average calories for user %v. The start date is: %v.", userDisplayName, startDate)
	averageCalories, averageCalErr := FetchAverageConsumedCalories(userId, startDate)
	if averageCalErr != nil {
		log.Printf("Error fetching average calories for user %v.", userDisplayName)
		s.InteractionRespond(i.Interaction, CreateEphemeralInteractionResponse("Error fetching your average calories, please try again..."))
		return
	}

	log.Printf("Retrieved average calories for user %v.", userDisplayName)
	s.InteractionRespond(i.Interaction, CreateEphemeralInteractionResponse(fmt.Sprintf("You have consumed an average of %d calories over %d days.", averageCalories, days)))
}

func createFoodLogEmbed(username string, date string, foodLogs []FoodLog, daily int64, consumed int64, remaining int64) *discordgo.MessageEmbed {
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
		Title:  fmt.Sprintf("Food Log - %s (%s)", username, date),
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
				Name:   "Daily",
				Value:  strconv.FormatInt(daily, 10),
				Inline: true,
			},
			{
				Name:   "Consumed",
				Value:  strconv.FormatInt(consumed, 10),
				Inline: true,
			},
			{
				Name:   "Remaining",
				Value:  strconv.FormatInt(remaining, 10),
				Inline: true,
			},
		},
		Timestamp: now.Format(time.RFC3339),
	}

	return embed
}
