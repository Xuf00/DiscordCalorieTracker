package component

import (
	"log"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/discordcalorietracker/discord"
	"github.com/discordcalorietracker/helper"
)

func HandleUpdateList(s *discordgo.Session, i *discordgo.InteractionCreate) {
	parts := strings.Split(i.MessageComponentData().CustomID, "_")
	userId := parts[1]
	userDisplayName := parts[2]
	providedDate := parts[3]

	date, dateParseErr := time.Parse("02/01/2006", providedDate)
	if dateParseErr != nil {
		log.Printf("Error parsing for user with username %v. Error: %v", userDisplayName, dateParseErr)
		s.InteractionRespond(i.Interaction, discord.CreateInteractionResponse("Encountered an error.", true, nil))
		return
	}

	helper.DisplayFoodLogEmbed(s, i, userId, userDisplayName, date, nil, true)
}
