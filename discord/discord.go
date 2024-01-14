package discord

import (
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
)

var S *discordgo.Session
var commandDefinitions []*discordgo.ApplicationCommand
var registeredCommands []*discordgo.ApplicationCommand
var commandHandlers map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate)
var componentHandlers map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate)

func InitDiscordSession(botToken string) {
	var err error
	S, err = discordgo.New("Bot " + botToken)
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}
	S.AddHandler(onReady)
	S.AddHandler(handleCommands)
}

func InitDiscordCommands(cmdDefinitions []*discordgo.ApplicationCommand, cmdHandlers map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate)) {
	commandDefinitions = cmdDefinitions
	commandHandlers = cmdHandlers
	registeredCommands = make([]*discordgo.ApplicationCommand, len(cmdDefinitions))
}

func InitDiscordComponentHandlers(cmpHandlers map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate)) {
	componentHandlers = cmpHandlers
}

func OpenDiscordSession() {
	err := S.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}
}

func AddCommandsDiscord(guildID string) {
	log.Println("Adding commands...")
	for i, v := range commandDefinitions {
		cmd, err := S.ApplicationCommandCreate(S.State.User.ID, guildID, v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
		log.Printf("Registered command %v", cmd.Name)
		registeredCommands[i] = cmd
	}
}

func RemoveCommandsDiscord(guildID string) {
	log.Println("Removing commands...")
	for _, v := range registeredCommands {
		err := S.ApplicationCommandDelete(S.State.User.ID, guildID, v.ID)
		if err != nil {
			log.Panicf("Cannot delete '%v' command: %v", v.Name, err)
		}
		log.Printf("Removed command %v", v.Name)
	}
}

func onReady(s *discordgo.Session, r *discordgo.Ready) {
	log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
}

func handleCommands(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		log.Printf("Handling slash command interaction %v", i.ApplicationCommandData().Name)
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}

	case discordgo.InteractionMessageComponent:
		log.Printf("Handling component interaction %v", i.MessageComponentData().CustomID)
		idPrefix := strings.Split(i.MessageComponentData().CustomID, "_")[0]
		if h, ok := componentHandlers[idPrefix]; ok {
			h(s, i)
		}
	}

}

func CreateInteractionResponse(content string, ephemeral bool, messageComponents []discordgo.MessageComponent) *discordgo.InteractionResponse {
	interactionResponse := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
		},
	}

	if ephemeral {
		// Message only shows for the user that triggered it
		interactionResponse.Data.Flags = discordgo.MessageFlagsEphemeral
	}

	if len(messageComponents) > 0 {
		interactionResponse.Data.Components = messageComponents
	}

	return interactionResponse
}
