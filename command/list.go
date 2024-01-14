package command

import (
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/discordcalorietracker/discord"
	"github.com/discordcalorietracker/helper"
)

func HandleListCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	userId := i.Member.User.ID
	userDisplayName := i.Member.User.GlobalName

	// Convert the slice into a map
	optionMap := helper.ConvertOptionsToMap(i)

	userParam, userProvided := optionMap["user"]
	if userProvided {
		user := userParam.UserValue(s)
		isBot := user.Bot
		userId = user.ID
		userDisplayName = user.GlobalName

		if isBot {
			log.Printf("User %v has requested to see a bots list which isn't valid.", i.Member.User.GlobalName)
			s.InteractionRespond(i.Interaction, discord.CreateInteractionResponse("That is a bot, please select a user.", true, nil))
			return
		}
		log.Printf("User %v has requested to see the list of user %v.", i.Member.User.GlobalName, userDisplayName)
	}

	startDate := time.Now()
	dateCmd, dateItemExists := optionMap["date"]
	if dateItemExists {
		date, dateParseErr := time.Parse("02/01/2006", dateCmd.StringValue())
		if dateParseErr != nil {
			log.Printf("Error parsing for user with username %v. Error: %v", userDisplayName, dateParseErr)
			s.InteractionRespond(i.Interaction, discord.CreateInteractionResponse(fmt.Sprintf("Error parsing date, please try again with format like %v.", startDate.Format("02/01/2006")), true, nil))
			return
		}
		startDate = date
	}

	helper.DisplayFoodLogEmbed(s, i, userId, userDisplayName, startDate, nil, false)
}
