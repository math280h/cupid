package board

import (
	"context"
	"math280h/cupid/db"
	"math280h/cupid/internal/shared"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

func getTopUsers() []struct {
	ID                   int
	DiscordID            string `json:"discordId"`
	DiscordUsername      string
	FlowersReceivedCount string `json:"flowers_received_count"`
} {
	query := `
		SELECT 
			u.id, 
			u.discordId, 
			u.discordUsername, 
			COUNT(f.id) AS flowers_received_count
		FROM "User" u
		LEFT JOIN "Flower" f ON u.id = f.receiver_id
		GROUP BY u.id, u.discordId, u.discordUsername
		ORDER BY flowers_received_count DESC
		LIMIT 10
	`

	var results []struct {
		ID                   int
		DiscordID            string `json:"discordId"`
		DiscordUsername      string
		FlowersReceivedCount string `json:"flowers_received_count"`
	}

	err := shared.DBClient.Prisma.QueryRaw(query).Exec(context.Background(), &results)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get top users")
		return nil
	}

	return results
}

func CreateBoard(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Get the channel ID from the interaction
	channel := i.ApplicationCommandData().Options[0].ChannelValue(s)

	// Get the top users
	topUsers := getTopUsers()
	var fields []*discordgo.MessageEmbedField

	// Create the fields for the embed
	for i, user := range topUsers {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:  strconv.Itoa(i+1) + ". " + user.DiscordUsername,
			Value: ":rose: " + user.FlowersReceivedCount + " roses received",
		})
	}

	embed := &discordgo.MessageEmbed{
		Title:  ":rose: Rose Leaderboard :rose:",
		Color:  0xFF0000,
		Fields: fields,
	}

	// Send the embed to the channel
	embedObj, err := s.ChannelMessageSendEmbed(channel.ID, embed)
	if err != nil {
		log.Error().Err(err).Msg("Failed to send leaderboard")
	}

	_, dbEmbedErr := shared.DBClient.Embed.FindMany().Delete().Exec(context.Background())
	if dbEmbedErr != nil {
		log.Error().Err(err).Msg("Failed to find embed record")
	}

	_, err = shared.DBClient.Embed.CreateOne(
		db.Embed.ChannelID.Set(channel.ID),
		db.Embed.MessageID.Set(embedObj.ID),
	).Exec(context.Background())
	if err != nil {
		log.Error().Err(err).Msg("Failed to create embed record")
	}

	shared.SimpleEphemeralInteractionResponse("Leaderboard sent!", s, i.Interaction)
}

func UpdateLeader(s *discordgo.Session) {
	// Update the embed message with the new leaderboard
	embedObj, err := shared.DBClient.Embed.FindFirst().Exec(context.Background())
	if err != nil {
		log.Error().Err(err).Msg("Failed to find embed record")
		return
	}

	topUsers := getTopUsers()
	var fields []*discordgo.MessageEmbedField

	// Create the fields for the embed
	for i, user := range topUsers {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:  strconv.Itoa(i+1) + ". " + user.DiscordUsername,
			Value: ":rose: " + user.FlowersReceivedCount + " roses received",
		})
	}

	embed := &discordgo.MessageEmbed{
		Title:  ":rose: Rose Leaderboard :rose:",
		Color:  0xFF0000,
		Fields: fields,
	}

	_, err = s.ChannelMessageEditEmbed(embedObj.ChannelID, embedObj.MessageID, embed)
	if err != nil {
		log.Error().Err(err).Msg("Failed to update leaderboard")
	}
}
