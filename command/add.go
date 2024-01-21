package command

import (
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/discordcalorietracker/database"
	"github.com/discordcalorietracker/discord"
	"github.com/discordcalorietracker/helper"
)

func HandleAddCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	userId := i.Member.User.ID
	userDisplayName := i.Member.User.GlobalName

	user, userErr := database.FetchUserByID(userId)
	if userErr != nil {
		log.Printf("Error fetching user with ID %v and username %v. Error: %v", userId, userDisplayName, userErr)
		s.InteractionRespond(i.Interaction, discord.CreateInteractionResponse("Error fetching user, please try again...", true, nil))
		return
	}

	if (database.User{}) == user {
		log.Printf("User with ID %v and username %v has tried to add without calling /set first.", userId, userDisplayName)
		s.InteractionRespond(i.Interaction, discord.CreateInteractionResponse("Set your daily calories first using the /set command.", true, nil))
		return
	}

	optionMap := helper.ConvertOptionsToMap(i)

	foodItem := optionMap["fooditem"].StringValue()
	calories := optionMap["calories"].IntValue()

	foodLog := database.FoodLog{
		UserID:   userId,
		FoodItem: foodItem,
		Calories: int16(calories),
		Quantity: 1,
	}

	if quantity, ok := optionMap["quantity"]; ok {
		itemQuantity := int16(quantity.IntValue())
		foodLog.Quantity = itemQuantity
	}

	id, addFoodLogErr := database.AddUserFoodLog(&foodLog)
	if addFoodLogErr != nil {
		log.Printf("Error adding food log for user with ID %v and username %v. Error: %v", userId, userDisplayName, addFoodLogErr)
		s.InteractionRespond(i.Interaction, discord.CreateInteractionResponse("There was an error, please try again...", true, nil))
		return
	}

	lastLogged := user.LastLogged.Format(helper.DATEFORMAT)
	currentDate := time.Now().Format(helper.DATEFORMAT)

	if lastLogged != currentDate {
		log.Printf("Updating the daily streak for user %v. They last logged on %v.", userDisplayName, lastLogged)
		n, err := database.UpdateUserStreak(userId)
		if n == 0 || err != nil {
			log.Printf("Error updating daily streak for user %v. Error: %v", userDisplayName, err)
			s.InteractionRespond(i.Interaction, discord.CreateInteractionResponse("There was an error, please try again...", true, nil))
			return
		}
	}

	messageComponents := helper.CreateAddRemoveUpdateButtons(userId, id, foodLog.FoodItem)

	log.Printf("Added food log %v for user %v and retrieved remaining calories.", id, userDisplayName)
	helper.DisplayFoodLogEmbed(s, i, userId, userDisplayName, time.Now(), messageComponents, true)
}
