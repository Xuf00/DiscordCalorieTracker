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

func HandleDeleteLog(s *discordgo.Session, i *discordgo.InteractionCreate) {
	userDisplayName := i.Member.User.GlobalName
	parts := strings.Split(i.MessageComponentData().CustomID, "_")
	userId := parts[1]

	logId, parseErr := strconv.ParseInt(parts[2], 10, 16)
	if parseErr != nil {
		log.Printf("Failed to parse log ID. Error: %v", parseErr)
		return
	}

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

	log.Printf("Deleted food log %v for user %v.", logId, userDisplayName)
	helper.DisplayFoodLogEmbed(s, i, userId, userDisplayName, time.Now(), nil, true)
}
