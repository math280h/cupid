package shared

import (
	"context"
	"math280h/cupid/db"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

var DBClient *db.PrismaClient //nolint:gochecknoglobals // This is the database client

func InitDB() {
	DBClient = db.NewClient()
	if err := DBClient.Prisma.Connect(); err != nil {
		panic(err)
	}
}

func FindOrCreateUser(discordUser *discordgo.User) (*db.UserModel, error) {
	userObj, err := DBClient.User.UpsertOne(
		db.User.DiscordID.Equals(discordUser.ID),
	).Create(
		db.User.DiscordID.Set(discordUser.ID),
		db.User.DiscordUsername.Set(discordUser.Username),
	).Update().Exec(context.Background())
	if err != nil {
		log.Error().Err(err).Msg("Failed to create user")
		return nil, err
	}

	return userObj, nil
}
