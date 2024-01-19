package component

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/discordcalorietracker/database"
	"github.com/discordcalorietracker/discord"
	"github.com/discordcalorietracker/helper"
)

func HandleModifyFoodQuantity(s *discordgo.Session, i *discordgo.InteractionCreate) {
	userDisplayName := i.Member.User.GlobalName
	parts := strings.Split(i.MessageComponentData().CustomID, "_")
	direction := parts[1]
	userId := parts[2]

	parsedId, parseErr := strconv.ParseInt(parts[3], 10, 16)
	if parseErr != nil {
		log.Printf("Failed to parse log ID. Error: %v", parseErr)
		return
	}

	logId := int64(parsedId)
	foodName := parts[4]

	n, updateErr := database.UpdateFoodLogQuantity(userId, logId, direction)
	if updateErr != nil {
		log.Printf("Error updating food log with ID %v for user %v: %v", logId, userDisplayName, updateErr)
		s.InteractionRespond(i.Interaction, discord.CreateInteractionResponse("There was an error, please try again...", true, nil))
		return
	}

	if n == 0 {
		log.Printf("Failed to update food log with ID %v for user %v.", logId, userDisplayName)
		s.InteractionRespond(i.Interaction, discord.CreateInteractionResponse(fmt.Sprintf("Failed to update the quantity for food log with ID %v.", logId), true, nil))
		return
	}

	messageComponents := helper.CreateAddRemoveUpdateButtons(userId, logId, foodName)

	log.Printf("Updated the quantity for food log %v for user %v and retrieved remaining calories.", logId, userDisplayName)
	helper.DisplayFoodLogEmbed(s, i, userId, userDisplayName, time.Now(), messageComponents, true)
}
