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

	messageComponents := []discordgo.MessageComponent{
		discordgo.Button{
			Emoji: discordgo.ComponentEmoji{
				Name: "⬆️",
			},
			Label:    fmt.Sprintf("Add a %s", foodLog.FoodItem),
			Style:    discordgo.SecondaryButton,
			CustomID: fmt.Sprintf("flquantity_inc_%s_%d_%s", userId, id, foodLog.FoodItem),
		},
	}

	if foodLog.Quantity > 1 {
		decreaseQuantityBtn := discordgo.Button{
			Emoji: discordgo.ComponentEmoji{
				Name: "⬇️",
			},
			Label:    fmt.Sprintf("Remove a %s", foodLog.FoodItem),
			Style:    discordgo.SecondaryButton,
			CustomID: fmt.Sprintf("flquantity_dec_%s_%d_%s", userId, id, foodLog.FoodItem),
		}

		messageComponents = append(messageComponents, decreaseQuantityBtn)
	}

	deleteBtn := discordgo.Button{
		Emoji: discordgo.ComponentEmoji{
			Name: "🚮",
		},
		Label:    fmt.Sprintf("Delete %s", foodLog.FoodItem),
		Style:    discordgo.DangerButton,
		CustomID: fmt.Sprintf("fldel_%s_%d_%s", userId, id, foodLog.FoodItem),
	}

	messageComponents = append(messageComponents, deleteBtn)

	log.Printf("Added food log %v for user %v and retrieved remaining calories.", id, userDisplayName)
	helper.DisplayFoodLogEmbed(s, i, userId, userDisplayName, time.Now(), messageComponents, true)
}
