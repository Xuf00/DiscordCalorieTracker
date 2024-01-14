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

func HandleDeleteCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	userId := i.Member.User.ID
	userDisplayName := i.Member.User.GlobalName

	// Convert the slice into a map
	optionMap := helper.ConvertOptionsToMap(i)

	logId := optionMap["logid"].IntValue()

	n, deleteErr := database.DeleteUserFoodLog(userId, logId)
	if deleteErr != nil {
		log.Printf("Error deleting food log for user %v: %v", userDisplayName, deleteErr)
		s.InteractionRespond(i.Interaction, discord.CreateInteractionResponse("There was an error, please try again...", true, nil))
		return
	}

	if n == 0 {
		log.Printf("Could not find a food log with ID %v for user %v.", logId, userDisplayName)
		s.InteractionRespond(i.Interaction, discord.CreateInteractionResponse(fmt.Sprintf("Could not find a food log with ID %v.", logId), true, nil))
		return
	}

	log.Printf("Deleted food log %v for user %v and retrieved remaining calories.", logId, userDisplayName)
	helper.DisplayFoodLogEmbed(s, i, userId, userDisplayName, time.Now(), nil, true)
}
