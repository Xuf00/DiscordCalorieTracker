package helper

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/discordcalorietracker/database"
	"github.com/discordcalorietracker/discord"
)

const DATEFORMAT = "02/01/2006"

func ConvertOptionsToMap(i *discordgo.InteractionCreate) map[string]*discordgo.ApplicationCommandInteractionDataOption {
	// Access options in the order provided by the user.
	options := i.ApplicationCommandData().Options
	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
	for _, opt := range options {
		optionMap[opt.Name] = opt
	}
	return optionMap
}

func DisplayFoodLogEmbed(s *discordgo.Session, i *discordgo.InteractionCreate, userId string, userDisplayName string, date time.Time, messageComponents []discordgo.MessageComponent, ephemeral bool) {
	user, userErr := database.FetchUserByID(userId)
	if userErr != nil {
		log.Printf("Error fetching user with ID %v and username %v. Error: %v", userId, userDisplayName, userErr)
		s.InteractionRespond(i.Interaction, discord.CreateInteractionResponse("Error fetching user, please try again...", true, nil))
		return
	}

	log.Printf("Fetching food logs for user %v on date %v.", userDisplayName, date.Format(DATEFORMAT))
	foodLogs, foodLogErr := database.FetchDailyFoodLogs(userId, date)
	if foodLogErr != nil {
		log.Printf("Error fetching food logs for user %v: %v", userDisplayName, foodLogErr)
		s.InteractionRespond(i.Interaction, discord.CreateInteractionResponse("Error fetching food logs, please try again...", true, nil))
		return
	}

	if len(foodLogs) == 0 {
		log.Printf("User %v has no logs on %v.", userDisplayName, date.Format(DATEFORMAT))
		s.InteractionRespond(i.Interaction, discord.CreateInteractionResponse(fmt.Sprintf("No logs found for %v on %v.", userDisplayName, date.Format(DATEFORMAT)), true, nil))
		return
	}

	daysSinceSunday := int(date.Weekday())
	previousSunday := date.AddDate(0, 0, -daysSinceSunday)

	log.Printf("Previous Sunday was %v and searched for date is %v. Had to subtract %v days.", previousSunday.Format(DATEFORMAT), date.Format(DATEFORMAT), daysSinceSunday)

	consumed, consumedErr := database.FetchConsumedCaloriesForDate(userId, date)
	remaining, remainingErr := database.FetchRemainingCalories(userId, date)
	remainingWeek, remainingWeekErr := database.FetchWeeksRemainingCalories(userId, previousSunday, date)
	if consumedErr != nil || remainingErr != nil || remainingWeekErr != nil {
		log.Printf("Error fetching consumed or remaining calories for user %v.", userDisplayName)
		s.InteractionRespond(i.Interaction, discord.CreateInteractionResponse("Error fetching consumed or remaining calories, please try again...", true, nil))
		return
	}

	if !ephemeral {
		updateBtn := discordgo.Button{
			Emoji: discordgo.ComponentEmoji{
				Name: "♻️",
			},
			Label:    "Update",
			Style:    discordgo.SecondaryButton,
			CustomID: fmt.Sprintf("fllist_%s_%s_%s", userId, userDisplayName, date.Format(DATEFORMAT)),
		}

		messageComponents = append(messageComponents, updateBtn)
	}

	embed := createFoodLogEmbed(userDisplayName, int64(user.DayStreak), date, foodLogs, int64(user.DailyCalories), consumed, remaining, remainingWeek)
	interactionResponse := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				embed,
			},
		},
	}

	if ephemeral {
		// Message only shows for the user that triggered it
		interactionResponse.Data.Flags = discordgo.MessageFlagsEphemeral
	}

	if len(messageComponents) > 0 {
		interactionResponse.Data.Components = []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: messageComponents,
			},
		}
	}

	s.InteractionRespond(i.Interaction, interactionResponse)
}

func createFoodLogEmbed(username string, streak int64, date time.Time, foodLogs []database.FoodLog, daily int64, consumed int64, remaining int64, remainingWeek int64) *discordgo.MessageEmbed {
	var foodItemNames strings.Builder
	var calories strings.Builder
	var times strings.Builder

	for _, foodLog := range foodLogs {
		totalCalories := foodLog.Calories
		if foodLog.Quantity > 1 {
			totalCalories = foodLog.Calories * foodLog.Quantity
			foodItemNames.WriteString(fmt.Sprintf("(%d) x%d %s\n", foodLog.ID, foodLog.Quantity, foodLog.FoodItem))
		} else {
			foodItemNames.WriteString(fmt.Sprintf("(%d) %s\n", foodLog.ID, foodLog.FoodItem))
		}
		calories.WriteString(fmt.Sprintf("%d\n", totalCalories))
		times.WriteString(fmt.Sprintf("%s\n", foodLog.DateTime.Format("15:04")))
	}

	now := time.Now()

	weeklyGoalStr := "Under"
	if remainingWeek < 0 {
		weeklyGoalStr = "Over"
	}

	stats := fmt.Sprintf("**Total Consumed**: %d\n**Remaining On Day**: %d\n**Calories %s Weekly Goal**: %d\n", consumed, remaining, weeklyGoalStr, remainingWeek)

	embed := &discordgo.MessageEmbed{
		Title:  fmt.Sprintf("Food Log - %s (%s)", username, date.Format(DATEFORMAT)),
		Author: &discordgo.MessageEmbedAuthor{},
		Color:  0x89CFF0,
		Fields: []*discordgo.MessageEmbedField{
			{
				Value: fmt.Sprintf("**Daily Calories**: %d\n", daily),
			},
			{
				Value: "\u200b",
			},
			{
				Name:   "Time",
				Value:  times.String(),
				Inline: true,
			},
			{
				Name:   "Name",
				Value:  foodItemNames.String(),
				Inline: true,
			},
			{
				Name:   "Calories",
				Value:  calories.String(),
				Inline: true,
			},
			{
				Value: "\u200b",
			},
			{
				Value: stats,
			},
		},
		Timestamp: now.Format(time.RFC3339),
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("%v day streak", streak),
		},
	}

	return embed
}

func CreateAddRemoveUpdateButtons(userId string, logId int64, foodName string) []discordgo.MessageComponent {
	return []discordgo.MessageComponent{
		discordgo.Button{
			Emoji: discordgo.ComponentEmoji{
				Name: "⬆️",
			},
			Label:    fmt.Sprintf("Add %s", foodName),
			Style:    discordgo.SecondaryButton,
			CustomID: fmt.Sprintf("flquantity_inc_%s_%d_%s", userId, logId, foodName),
		},
		discordgo.Button{
			Emoji: discordgo.ComponentEmoji{
				Name: "⬇️",
			},
			Label:    fmt.Sprintf("Remove %s", foodName),
			Style:    discordgo.SecondaryButton,
			CustomID: fmt.Sprintf("flquantity_dec_%s_%d_%s", userId, logId, foodName),
		},
		discordgo.Button{
			Emoji: discordgo.ComponentEmoji{
				Name: "🚮",
			},
			Label:    fmt.Sprintf("Delete %s", foodName),
			Style:    discordgo.DangerButton,
			CustomID: fmt.Sprintf("fldel_%s_%d_%s", userId, logId, foodName),
		},
	}
}
