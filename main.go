package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
)

// Bot parameters
var (
	GuildID        = flag.String("guild", "", "Test guild ID. If not passed - bot registers commands globally")
	BotToken       = flag.String("token", "", "Bot access token")
	RemoveCommands = flag.Bool("rmcmd", true, "Remove all commands after shutting down or not")
)

func main() {
	flag.Parse()

	InitDatabase()

	InitDiscordSession()

	OpenDiscordSession()

	AddCommandsDiscord()

	defer db.Close()
	defer s.Close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Println("Press Ctrl+C to exit")
	<-stop

	if *RemoveCommands {
		RemoveCommandsDiscord()
	}

	log.Println("Gracefully shutting down.")
}
