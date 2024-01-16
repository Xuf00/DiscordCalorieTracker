package component

import (
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/discordcalorietracker/helper"
)

func HandleUpdateList(s *discordgo.Session, i *discordgo.InteractionCreate) {
	parts := strings.Split(i.MessageComponentData().CustomID, "_")
	userId := parts[1]
	userDisplayName := parts[2]
	helper.DisplayFoodLogEmbed(s, i, userId, userDisplayName, time.Now(), nil, false)
}
