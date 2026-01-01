package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	channelcreation "github.com/pzsp-teams/cli/internal/channel_creation"
	"github.com/pzsp-teams/cli/internal/client"
	"github.com/pzsp-teams/cli/internal/file_readers"
	"github.com/pzsp-teams/cli/internal/initializers"
	"github.com/pzsp-teams/cli/internal/logger"
	"github.com/pzsp-teams/cli/internal/messaging"
	"github.com/pzsp-teams/cli/internal/templates"
)

func init() {
	verbose := false
	for _, arg := range os.Args {
		if arg == "-v" || arg == "--verbose" {
			verbose = true
			break
		}
	}

	logFile, err := os.Create("preview.log")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create log file: %v\n", err)
		os.Exit(1)
	}

	stderrLevel := logger.LevelError
	if verbose {
		stderrLevel = logger.LevelDebug
	}

	initializers.InitMultiOutputLogger(initializers.MultiOutputConfig{
		StderrLevel:         stderrLevel,
		FileLevel:           logger.LevelDebug,
		FileWriter:          logFile,
		StderrOmitTimestamp: !verbose,
		FileOmitTimestamp:   false,
	})
}

func main() {
	// bulkMessageDemo("channels", "preview/channel_message.txt", "preview/channel_message_data.yaml", "pzsp2z1teams", false, true)
	bulkMessageDemo("chats", "preview/chat_message.txt", "preview/chat_message_data.yaml", "", false, true)
	bulkMessageDemo("chats", "preview/group_chat_message.txt", "preview/group_chat_message_data.yaml", "", false, true)
	createChannelsDemo()
	// charmDemo()
}

func mapExtensionToDecodeFunc(extension string) (file_readers.DecodeFunc, error) {
	return file_readers.GetDecoderByExtension(extension)
}

func bulkMessageDemo(targetType, messageFileName, dataFileName, teamName string, dryRun, ignoreError bool) {
	log := initializers.Logger
	ctx := context.TODO()

	dataFile, err := os.Open(dataFileName)
	if err != nil {
		log.Error("Failed to open data file", "file", dataFileName, "error", err)
		os.Exit(1)
	}

	templateFile, err := os.Open(messageFileName)
	if err != nil {
		_ = dataFile.Close()
		log.Error("Failed to open message file", "file", messageFileName, "error", err)
		os.Exit(1)
	}

	extension := filepath.Ext(dataFile.Name())[1:]
	parser, err := mapExtensionToDecodeFunc(extension)
	if err != nil {
		log.Error("Failed to get decode function", "error", err)
		os.Exit(1)
	}

	messageParser, err := templates.NewMessageParser(templateFile, dataFile, parser)
	_ = templateFile.Close()
	_ = dataFile.Close()
	if err != nil {
		log.Error("Failed to create message parser", "error", err)
		os.Exit(1)
	}

	messages, err := messageParser.Parse()
	if err != nil {
		log.Error("Failed to render messages", "error", err)
		os.Exit(1)
	}

	senderConfig := newSenderConfig()
	authConfig := loadAuthConfig()
	cacheConfig := newCacheConfig()
	teamsClient, err := client.NewTeamsClient(ctx, authConfig, senderConfig, cacheConfig)
	if err != nil {
		log.Error("Error creating Teams client", "error", err)
		os.Exit(1)
	}

	switch targetType {
	case "channels":
		if teamName == "" {
			log.Error("Team name is required for sending to channels")
			os.Exit(1)
		}
		log.Info("Sending messages to channels", "team", teamName, "count", len(messages), "dryRun", dryRun)
		results := teamsClient.ChannelSender.SendToChannels(ctx, teamName, messages, dryRun, ignoreError)
		printChannelResults(results, dryRun)

	case "chats":
		log.Info("Sending messages to chats", "count", len(messages), "dryRun", dryRun)
		results := teamsClient.ChatSender.SendToChats(ctx, messages, dryRun, ignoreError)
		printChatResults(results, dryRun)

	default:
		log.Error("Invalid target type", "targetType", targetType, "validTypes", "channels, chats")
		os.Exit(1)
	}
}

func printChannelResults(results []messaging.ChannelSendResult, dryRun bool) {
	log := initializers.Logger
	if dryRun {
		for _, res := range results {
			if res.Error != nil {
				log.Warn("Would fail", "channel", res.ChannelRef, "error", res.Error)
			} else {
				log.Info("Would send", "channel", res.ChannelRef, "message", res.Message)
			}
		}
	} else {
		for _, res := range results {
			if res.Error != nil {
				log.Error("Failed", "channel", res.ChannelRef, "error", res.Error)
			}
		}
	}
}

func printChatResults(results []messaging.ChatSendResult, dryRun bool) {
	log := initializers.Logger
	if dryRun {
		for _, res := range results {
			if res.Error != nil {
				log.Warn("Would fail", "chat", res.ChatRef, "error", res.Error)
			} else {
				log.Info("Would send", "chat", res.ChatRef, "message", res.Message)
			}
		}
	}
}

func createChannelsDemo() {
	log := initializers.Logger
	ctx := context.TODO()

	senderConfig := newSenderConfig()
	authConfig := loadAuthConfig()
	cacheConfig := newCacheConfig()
	teamsClient, err := client.NewTeamsClient(ctx, authConfig, senderConfig, cacheConfig)
	if err != nil {
		log.Error("Error creating Teams client", "error", err)
		os.Exit(1)
	}
	dataFile, err := os.Open("preview/channels.csv")
	if err != nil {
		log.Error("Failed to open channels file", "error", err)
		os.Exit(1)
	}
	extension := filepath.Ext(dataFile.Name())[1:]

	channelData, err := channelcreation.ParseChannelsDataByExtension(dataFile, extension)
	teamRef := "pzsp2z1teams"
	if err != nil {
		log.Error("Failed to parse channels data", "error", err)
		_ = dataFile.Close()
		os.Exit(1)
	}
	_ = dataFile.Close()
	successCount := 0
	results := teamsClient.ChannelCreator.CreateChannels(ctx, teamRef, channelData, true, true)
	for _, result := range results {
		if result.Error == nil {
			log.Info("Channel operation successful",
				"channel", result.ChannelName,
				"channel_id", result.ChannelID,
				"status", result.Status,
				"members", result.MemberRefs,
				"owners", result.OwnerRefs)
			successCount++
		} else {
			log.Error("Failed to perform channel operation",
				"channel", result.ChannelName,
				"error", result.Error,
				"status", result.Status)
		}
	}
	log.Info("Channel creation demo completed",
		"total", len(results),
		"successful", successCount,
		"failed", len(results)-successCount)
}

func charmDemo() {
	log := initializers.Logger
	log.Debug("this is a debug message")
	log.Info("application started successfully")
	log.Warn("this is a warning", "code", "WARN001")
	log.Error("this is an error", "error", "something went wrong")

	log.Info("=== Logging with Context (With) ===")
	userLogger := log.With("user_id", "12345", "role", "admin")
	userLogger.Info("user logged in")
	userLogger.Info("user accessed dashboard")

	requestLogger := userLogger.With("request_id", "req-abc-123")
	requestLogger.Info("processing request", "endpoint", "/api/users")
	requestLogger.Warn("slow request detected", "duration_ms", 2500)

	log.Debug("debug level message")
	log.Info("info level message")
	log.Warn("warning level message")
	log.Error("error level message")
}
