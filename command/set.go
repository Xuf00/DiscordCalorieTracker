package command

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/discordcalorietracker/database"
	"github.com/discordcalorietracker/discord"
	"github.com/discordcalorietracker/helper"
)

func HandleSetCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	userId := i.Member.User.ID
	userDisplayName := i.Member.User.GlobalName

	// Convert the slice into a map
	optionMap := helper.ConvertOptionsToMap(i)

	calories := optionMap["calories"].IntValue()

	user := database.User{
		ID:            userId,
		DailyCalories: int16(calories),
	}

	_, setCaloriesErr := database.SetUserCalories(&user)
	if setCaloriesErr != nil {
		log.Printf("Error setting calories for user with ID %v and username %v. Error: %v", userId, userDisplayName, setCaloriesErr)
		s.InteractionRespond(i.Interaction, discord.CreateInteractionResponse("There was an error, please try again...", true, nil))
		return
	}

	log.Printf("Successfully set daily calorie intake to %d for user with ID %v and username %v.", calories, userId, userDisplayName)
	s.InteractionRespond(i.Interaction, discord.CreateInteractionResponse(fmt.Sprintf("Your daily calorie intake has successfully been set to %d.", calories), true, nil))
}
