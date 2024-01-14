package command

import (
	"fmt"
	"log"
	"math"

	"github.com/bwmarrin/discordgo"
	"github.com/discordcalorietracker/discord"
	"github.com/discordcalorietracker/helper"
)

func HandleConvCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	userDisplayName := i.Member.User.GlobalName

	// Convert the slice into a map
	optionMap := helper.ConvertOptionsToMap(i)

	units := optionMap["units"].FloatValue()
	calories := optionMap["calories"].FloatValue()
	weight := optionMap["weight"].FloatValue()

	perUnit := calories / units
	totalCalories := perUnit * weight

	if foodItem, ok := optionMap["fooditem"]; ok {
		log.Printf("User %v provided the optional food item name when converting.", userDisplayName)
		s.InteractionRespond(i.Interaction, discord.CreateInteractionResponse(fmt.Sprintf("%v\n%.2f calories per unit \nTotal amount of calories is %.0f", foodItem.StringValue(), perUnit, math.Ceil(totalCalories)), false, nil))
	} else {
		log.Printf("User %v did not provide the optional food item name when converting.", userDisplayName)
		s.InteractionRespond(i.Interaction, discord.CreateInteractionResponse(fmt.Sprintf("%.2f calories per unit \nTotal amount of calories is %.0f", perUnit, math.Ceil(totalCalories)), false, nil))
	}
}
