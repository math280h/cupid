package main

import (
	"context"
	"math280h/cupid/db"
	"math280h/cupid/internal/board"
	"math280h/cupid/internal/shared"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var s *discordgo.Session //nolint:gochecknoglobals // This is the Discord session

var (
	commands = []*discordgo.ApplicationCommand{ //nolint:gochecknoglobals // This is a list of commands
		{
			Name:        "rose",
			Description: "Give a rose to someone",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "user",
					Description: "The user to add a note to",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "text",
					Description: "The text of the note",
					Required:    true,
				},
				{
					Type: discordgo.ApplicationCommandOptionBoolean,
					Name: "private",
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{
							Name:  "Public",
							Value: true,
						},
						{
							Name:  "Private",
							Value: false,
						},
					},
					Description: "Whether the text is private, defaults to private",
					Required:    false,
				},
			},
		},
		{
			Name:        "board",
			Description: "Create a leaderboard",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionChannel,
					Name:        "channel",
					Description: "The channel to create the leaderboard in",
					Required:    true,
				},
			},
		},
	}

	commandHandlers = map[string]func( //nolint:gochecknoglobals // This is a map of commands to their handlers
		s *discordgo.Session,
		i *discordgo.InteractionCreate,
	){
		"rose":  roseCommand,
		"board": board.CreateBoard,
	}
)

func onReady(s *discordgo.Session, _ *discordgo.Ready) {
	log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
}

func roseCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Debug().Msg("rose command called")
	// Get the user to give the rose to
	user := i.ApplicationCommandData().Options[0].UserValue(s)
	// Get the text of the rose
	text := i.ApplicationCommandData().Options[1].StringValue()
	// Get whether the rose is private
	var private bool
	// If there is no Options[2], it means the rose is private
	if len(i.ApplicationCommandData().Options) < 3 {
		log.Debug().Msg("No private option, defaulting to private")
		private = true
	} else {
		private = i.ApplicationCommandData().Options[2].BoolValue()
	}

	// Get the target user object
	targetUserObj, targetUserErr := shared.FindOrCreateUser(user)
	if targetUserErr != nil {
		log.Error().Err(targetUserErr).Msg("Cannot find or create user")
		shared.SimpleEphemeralInteractionResponse("Cannot find or create user", s, i.Interaction)
		return
	}

	// Get the giver user object
	giverUserObj, giverUserErr := shared.FindOrCreateUser(i.Member.User)
	if giverUserErr != nil {
		log.Error().Err(giverUserErr).Msg("Cannot find or create user")
		shared.SimpleEphemeralInteractionResponse("Cannot find or create user", s, i.Interaction)
		return
	}

	// Create the rose
	_, err := shared.DBClient.Flower.CreateOne(
		// Create the rose
		db.Flower.Giver.Link(
			db.User.ID.Equals(giverUserObj.ID),
		),
		db.Flower.Receiver.Link(
			db.User.ID.Equals(targetUserObj.ID),
		),
		db.Flower.Message.Set(text),
		db.Flower.Public.Set(private),
	).Exec(context.Background())
	if err != nil {
		log.Error().Err(err).Msg("Cannot create rose")
		shared.SimpleEphemeralInteractionResponse("Cannot find or create user", s, i.Interaction)
		return
	}

	board.UpdateLeader(s)

	// Send a DM to the target user
	channel, err := s.UserChannelCreate(targetUserObj.DiscordID)
	if err != nil {
		log.Error().Err(err).Msg("Cannot create DM channel")
		shared.SimpleEphemeralInteractionResponse("Cannot create DM channel", s, i.Interaction)
		return
	}

	embed := &discordgo.MessageEmbed{
		Title:       "🌹 You received a rose! 🌹",
		Description: text,
		Color:       0xFF0000, // Red color
		Footer: &discordgo.MessageEmbedFooter{
			Text: "From: " + "Your Secret Admirer xoxo",
		},
	}
	_, err = s.ChannelMessageSendEmbed(channel.ID, embed)
	if err != nil {
		log.Error().Err(err).Msg("Cannot send DM to user")
		shared.SimpleEphemeralInteractionResponse("Cannot send DM to user", s, i.Interaction)
		return
	}

	// Respond to the interaction
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Rose command called",
			Flags:   64,
		},
	})
	if err != nil {
		log.Error().Err(err).Msg("Cannot respond to the interaction")
	}
}

func main() {
	shared.Init()
	shared.InitDB()

	// Set up logging
	log.Logger = log.With().Caller().Logger() //nolint:reassign // This is the only way to enable caller information
	log.Logger = log.Output(                  //nolint:reassign // This is the only way to set the output
		zerolog.ConsoleWriter{Out: os.Stderr},
	)

	var err error
	s, err = discordgo.New("Bot " + *shared.BotToken)
	if err != nil {
		log.Fatal().Err(err).Msg("Invalid bot parameters")
	}

	s.Identify.Intents = discordgo.IntentsGuildMessages |
		discordgo.IntentsDirectMessages |
		discordgo.IntentsGuildMembers |
		discordgo.IntentMessageContent

	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type == discordgo.InteractionApplicationCommand {
			// Make sure it's an application command (e.g., /mycommand)
			if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
				h(s, i)
			}
			return
		}
	})
	s.AddHandler(onReady)

	// Open the session to begin listening for events
	err = s.Open()
	if err != nil {
		log.Fatal().Err(err).Msg("Cannot open the session")
	}

	// Register available slash commands
	log.Info().Msg("Adding commands...")
	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
	for i, v := range commands {
		log.Debug().Msgf("Adding command: %v", v.Name)
		var cmd *discordgo.ApplicationCommand
		cmd, err = s.ApplicationCommandCreate(s.State.User.ID, *shared.GuildID, v)
		if err != nil {
			log.Fatal().Err(err).Msgf("Cannot create '%v' command", v.Name)
		}
		registeredCommands[i] = cmd
	}
	log.Info().Msg("Commands added successfully")

	defer s.Close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Info().Msg("Bot is now running. Press CTRL+C to exit.")
	<-stop

	defer func() {
		if err = shared.DBClient.Prisma.Disconnect(); err != nil {
			panic(err)
		}
	}()
}
