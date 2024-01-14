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

func HandleUpdateCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	userId := i.Member.User.ID
	userDisplayName := i.Member.User.GlobalName

	optionMap := helper.ConvertOptionsToMap(i)

	logId := optionMap["logid"].IntValue()
	foodItem := optionMap["fooditem"].StringValue()
	calories := optionMap["calories"].IntValue()

	foodLog := database.FoodLog{
		ID:       logId,
		UserID:   userId,
		FoodItem: foodItem,
		Calories: int16(calories),
		Quantity: 1,
	}

	if quantity, ok := optionMap["quantity"]; ok {
		itemQuantity := int16(quantity.IntValue())
		foodLog.Quantity = itemQuantity
	}

	n, updateErr := database.UpdateUserFoodLog(&foodLog)
	if updateErr != nil {
		log.Printf("Error updating food log with ID %v for user %v: %v", logId, userDisplayName, updateErr)
		s.InteractionRespond(i.Interaction, discord.CreateInteractionResponse("There was an error, please try again...", true, nil))
		return
	}

	if n == 0 {
		log.Printf("Could not find a food log with ID %v for user %v.", logId, userDisplayName)
		s.InteractionRespond(i.Interaction, discord.CreateInteractionResponse(fmt.Sprintf("Could not find a food log with ID %v.", logId), true, nil))
		return
	}

	log.Printf("Updated food log %v for user %v and retrieved remaining calories.", logId, userDisplayName)
	helper.DisplayFoodLogEmbed(s, i, userId, userDisplayName, time.Now(), nil, true)
}
