package command

import (
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/discordcalorietracker/database"
	"github.com/discordcalorietracker/discord"
)

func HandleRemCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	userId := i.Member.User.ID
	userDisplayName := i.Member.User.GlobalName

	log.Printf("Fetching remaining calories for user %v.", userDisplayName)
	remaining, remainingErr := database.FetchRemainingCalories(userId, time.Now())
	if remainingErr != nil {
		log.Printf("Error fetching remaining calories for user %v.", userDisplayName)
		s.InteractionRespond(i.Interaction, discord.CreateInteractionResponse("Error fetching remaining calories, please try again...", true, nil))
		return
	}

	if remaining == 10000 {
		log.Printf("Couldn't get the remaining calories for user %v.", userDisplayName)
		s.InteractionRespond(i.Interaction, discord.CreateInteractionResponse("Couldn't work out your remaining calories, make sure you've used /set to set your daily calories.", true, nil))
		return
	}

	s.InteractionRespond(i.Interaction, discord.CreateInteractionResponse(fmt.Sprintf("You have %d calories remaining today.", remaining), true, nil))
}
