package shared

import (
	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

func SimpleEphemeralInteractionResponse(
	content string,
	session *discordgo.Session,
	interaction *discordgo.Interaction,
) {
	err := session.InteractionRespond(interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
			Flags:   64,
		},
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to respond to interaction")
	}
}
