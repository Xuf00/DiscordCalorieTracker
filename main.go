package main

import (
	"flag"
	"log"
	"os"
	"os/signal"

	"github.com/discordcalorietracker/command"
	"github.com/discordcalorietracker/component"
	"github.com/discordcalorietracker/database"
	"github.com/discordcalorietracker/discord"
)

// Bot parameters
var (
	GuildID        = flag.String("guild", "", "Test guild ID. If not passed - bot registers commands globally")
	BotToken       = flag.String("token", "", "Bot access token")
	RemoveCommands = flag.Bool("rmcmd", true, "Remove all commands after shutting down or not")
)

func main() {
	flag.Parse()

	database.InitDatabase()

	discord.InitDiscordSession(*BotToken)
	discord.OpenDiscordSession()
	discord.InitDiscordCommands(command.CommandDefinitions, command.CommandHandlers)
	discord.InitDiscordComponentHandlers(component.ComponentHandlers)
	discord.AddCommandsDiscord(*GuildID)

	defer database.DB.Close()
	defer discord.S.Close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Println("Press Ctrl+C to exit")
	<-stop

	if *RemoveCommands {
		discord.RemoveCommandsDiscord(*GuildID)
	}

	log.Println("Gracefully shutting down.")
}
