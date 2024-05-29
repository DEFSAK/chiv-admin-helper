package main

import (
	"context"
	"fmt"
	"github.com/charmbracelet/log"
	"os"
	"os/signal"
	"strings"
	"time"
)

func main() {
	// Crash guard
	defer func() {
		err := recover()
		if err != nil {
			log.Error("Fatal error occurred", "err", err)
			log.Info("If this error persists please create a bug report")
			fmt.Println("Press Enter to close this window...")
			fmt.Scan(nil)
			os.Exit(1)
		}
	}()

	// Setup logger
	log.SetFormatter(log.TextFormatter)
	log.SetReportCaller(false)

	// Make sure the user has credentials
	credentialPath, err := setupCredentials()
	if err != nil {
		log.Error("Credential setup failed", "err", err)
	}

	// Login to backend
	svc, err := newBackendService(credentialPath)
	if err != nil {
		log.Error("Login to backend failed", "err", err)
	}

	// Setup watchers for commands, clipboard copy and os interrupts
	interrupts := make(chan os.Signal, 1)
	signal.Notify(interrupts, os.Interrupt)
	ctx, cancelWatchers := context.WithCancel(context.Background())
	clipboardEvents := watchClipboard(ctx, time.Millisecond*50)
	stdinEvents := watchStdin(ctx)

	// Start the main loop
	log.Info("Chiv admin helper is ready to use")
	log.Info("Use the listplayers command in game to validate players. Press Ctrl+C to abort")
	validatedPlayers := make([]validatedPlayer, 0)
mainLoop:
	for {
		select {
		case event := <-stdinEvents:
			err = executeCommand(event, validatedPlayers)
			if err != nil {
				log.Warn("Failed to execute command", "err", err)
				continue mainLoop
			}
		case event := <-clipboardEvents:
			// Read listplayers from clipboard
			if strings.HasPrefix(event, "ServerName - ") {
				serverName, players, err := readPlayerList(event)
				if err != nil {
					log.Warn("Failed to read player list", "err", err)
					continue mainLoop
				}
				validatedPlayers, _ = svc.validatePlayers(serverName, players)
				log.Info("Validated players", "count", len(validatedPlayers))
				printTable(validatedPlayers)
			}
		case <-interrupts:
			// Close program
			cancelWatchers()
			time.Sleep(time.Millisecond)
			break mainLoop
		}
	}
}
