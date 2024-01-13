package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
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

	n, updateErr := UpdateFoodLogQuantity(userId, logId, direction)
	if updateErr != nil {
		log.Printf("Error updating food log with ID %v for user %v: %v", logId, userDisplayName, updateErr)
		s.InteractionRespond(i.Interaction, CreateInteractionResponse("There was an error, please try again...", true, nil))
		return
	}

	if n == 0 {
		log.Printf("Failed to update food log with ID %v for user %v.", logId, userDisplayName)
		s.InteractionRespond(i.Interaction, CreateInteractionResponse(fmt.Sprintf("Failed to update food log with ID %v.", logId), true, nil))
		return
	}

	log.Printf("Updated food log %v for user %v and retrieved remaining calories.", logId, userDisplayName)

	s.InteractionRespond(i.Interaction, CreateInteractionResponse(fmt.Sprintf("Successfully updated the quantity for food log with ID %v.", logId), true, nil))
}

func HandleUpdateList(s *discordgo.Session, i *discordgo.InteractionCreate) {
	parts := strings.Split(i.MessageComponentData().CustomID, "_")
	userId := parts[1]
	userDisplayName := parts[2]
	providedDate := parts[3]

	date, dateParseErr := time.Parse("02/01/2006", providedDate)
	if dateParseErr != nil {
		log.Printf("Error parsing for user with username %v. Error: %v", userDisplayName, dateParseErr)
		s.InteractionRespond(i.Interaction, CreateInteractionResponse("Encountered an error.", true, nil))
		return
	}

	user, userErr := FetchUserByID(userId)
	if userErr != nil {
		log.Printf("Error fetching user with ID %v and username %v. Error: %v", userId, userDisplayName, userErr)
		s.InteractionRespond(i.Interaction, CreateInteractionResponse("Error fetching user, please try again...", true, nil))
		return
	}

	log.Printf("Fetching food logs for user %v on date %v.", userDisplayName, date.Format("02/01/2006"))
	foodLogs, foodLogErr := FetchDailyFoodLogs(userId, date)
	if foodLogErr != nil {
		log.Printf("Error fetching food logs for user %v: %v", userDisplayName, foodLogErr)
		s.InteractionRespond(i.Interaction, CreateInteractionResponse("Error fetching food logs, please try again...", true, nil))
		return
	}

	if len(foodLogs) == 0 {
		log.Printf("User %v has no logs on %v.", userDisplayName, date.Format("02/01/2006"))
		s.InteractionRespond(i.Interaction, CreateInteractionResponse(fmt.Sprintf("No logs found for %v on %v.", userDisplayName, date.Format("02/01/2006")), true, nil))
		return
	}

	consumed, consumedErr := FetchConsumedCaloriesForDate(userId, date)
	remaining, remainingErr := FetchRemainingCalories(userId, date)
	if consumedErr != nil || remainingErr != nil {
		log.Printf("Error fetching consumed or remaining calories for user %v.", userDisplayName)
		s.InteractionRespond(i.Interaction, CreateInteractionResponse("Error fetching consumed or remaining calories, please try again...", true, nil))
		return
	}

	embed := createFoodLogEmbed(userDisplayName, date, foodLogs, int64(user.daily_calories), consumed, remaining)
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				embed,
			},
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.Button{
							Emoji: discordgo.ComponentEmoji{
								Name: "♻️",
							},
							Label:    "Update",
							Style:    discordgo.SecondaryButton,
							CustomID: fmt.Sprintf("fllist_%s_%s_%s", userId, userDisplayName, date.Format("02/01/2006")),
						},
					},
				},
			},
		},
	})
}
