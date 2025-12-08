package main

import (
	"context"
	"fmt"
	"os"

	"github.com/pzsp-teams/cli/internal/client"
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
	bulkMessageDemo()
}

func bulkMessageDemo() {
	log := initializers.Logger
	ctx := context.TODO()

	dataFile, err := os.Open("data.yaml")
	if err != nil {
		log.Error("Failed to open data.yaml", "error", err)
		os.Exit(1)
	}

	templateFile, err := os.Open("message.txt")
	if err != nil {
		_ = dataFile.Close()
		log.Error("Failed to open message.txt", "error", err)
		os.Exit(1)
	}

	parser := &templates.YAMLParser{}
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

	for recipient, message := range messages {
		log.Debug("Rendered message", "recipient", recipient, "content", message)
	}

	senderConfig := newSenderConfig()
	authConfig := loadAuthConfig()
	teamsClient, err := client.NewTeamsClient(ctx, authConfig, senderConfig)
	if err != nil {
		log.Error("Error creating Teams client", "error", err)
		os.Exit(1)
	}

	teamName := "pzsp2"
	results := messaging.SendToChannels(ctx, teamsClient, teamName, messages)

	successCount := 0
	for _, result := range results {
		if result.Error == nil {
			successCount++
			log.Info("Message sent successfully",
				"channel", result.ChannelRef,
				"message_id", result.Message.ID)
		} else {
			log.Error("Failed to send message",
				"channel", result.ChannelRef,
				"error", result.Error)
		}
	}

	log.Info("Bulk message demo completed",
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
