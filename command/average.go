package command

import (
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/discordcalorietracker/database"
	"github.com/discordcalorietracker/discord"
	"github.com/discordcalorietracker/helper"
)

func HandleAverageCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	userId := i.Member.User.ID
	userDisplayName := i.Member.User.GlobalName

	log.Printf("Checking user %v has enough data to get an average.", userDisplayName)
	count, countErr := database.FetchFoodLogDaysCount(userId)
	if countErr != nil {
		log.Printf("Error checking if the user %v has enough data to get an average.", userDisplayName)
		s.InteractionRespond(i.Interaction, discord.CreateInteractionResponse("Error checking your average calories, please try again...", true, nil))
		return
	}

	// Convert the slice into a map
	optionMap := helper.ConvertOptionsToMap(i)

	days := optionMap["days"].IntValue()

	if count != days {
		log.Printf("User %v doesn't have enough data to get an average. They have %d days of data but requested an average for %d.", userDisplayName, count, days)
		s.InteractionRespond(i.Interaction, discord.CreateInteractionResponse(fmt.Sprintf("You only have enough data to request an average over %d days.", count), true, nil))
		return
	}

	startDate := time.Now().AddDate(0, 0, -int(days)).Format("2006-01-02")
	log.Printf("Fetching average calories for user %v. The start date is: %v.", userDisplayName, startDate)
	averageCalories, averageCalErr := database.FetchAverageConsumedCalories(userId, startDate)
	if averageCalErr != nil {
		log.Printf("Error fetching average calories for user %v.", userDisplayName)
		s.InteractionRespond(i.Interaction, discord.CreateInteractionResponse("Error fetching your average calories, please try again...", true, nil))
		return
	}

	log.Printf("Retrieved average calories for user %v.", userDisplayName)
	s.InteractionRespond(i.Interaction, discord.CreateInteractionResponse(fmt.Sprintf("You have consumed an average of %d calories over %d days.", averageCalories, days), true, nil))
}
