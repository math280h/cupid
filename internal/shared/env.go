package shared

import (
	"flag"
	"os"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

var (
	BotToken = flag.String("token", "", "Bot access token") //nolint:gochecknoglobals,lll // This is a flag shared across the application
	GuildID  = flag.String("guild", "", "Guild ID")         //nolint:gochecknoglobals,lll // This is a flag shared across the application
)

func Init() {
	flag.Parse()

	envErr := godotenv.Load()
	if envErr != nil {
		log.Info().Msg("No .env file found, trying to using environment variables")
	}

	if *BotToken == "" {
		*BotToken = os.Getenv("DISCORD_BOT_TOKEN")
	}

	if *GuildID == "" {
		*GuildID = os.Getenv("DISCORD_GUILD_ID")
	}
}
