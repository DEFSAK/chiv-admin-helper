package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/charmbracelet/log"
	"io"
	"os"
	"strconv"
	"strings"
)

var localTrustList = make([]string, 0)

// watchStdin monitors console input and notifies the channel when a new command is received
func watchStdin(ctx context.Context) (events chan string) {
	events = make(chan string)
	stdin := bufio.NewReader(os.Stdin)
	go func() {
		for {
			if ctx.Err() != nil {
				close(events)
				return
			}
			command, _ := stdin.ReadString('\n')
			command = strings.TrimSpace(command)
			if command == "" {
				continue
			}
			events <- command
		}
	}()
	return
}

func executeCommand(command string, players []validatedPlayer, svc backendService) (outputCommand string, err error) {
	// Parse command and arguments
	rd := strings.NewReader(command)
	args := make([]string, 0, 2)
	var arg string
	for {
		_, err = fmt.Fscan(rd, &arg)
		if err != nil {
			break
		} else {
			args = append(args, arg)
		}
	}
	if !errors.Is(err, io.EOF) || len(args) < 2 {
		err = errors.New("invalid command format")
		return
	}
	index, err := strconv.Atoi(args[1])
	if err != nil || index < 0 || index >= len(players) {
		index = -1
	}
	err = nil

	// Execute command
	switch args[0] {
	case "kick":
		// Generate a one time kick command
		if index == -1 {
			err = errors.New("invalid player number")
			break
		}
		outputCommand = "kickbyid " + players[index].PlayfabId
	case "ban":
		// Ban a player globally
		if index == -1 {
			err = errors.New("invalid player number")
			break
		}
		if len(args) < 3 {
			err = errors.New("ban requires at least 1 reason")
			break
		}
		outputCommand, err = svc.playerAction("ban", players[index].PlayfabId, map[string]any{
			"charges": args[2:],
		})
	case "banbyid":
		// Ban a player that is not currently in the lobby
		if len(args) < 3 {
			err = errors.New("banbyid requires at least 1 reason")
			break
		}
		outputCommand, err = svc.playerAction("ban", args[1], map[string]any{
			"charges": args[2:],
		})
	case "unbanbyid":
		outputCommand, err = svc.playerAction("unban", args[1], nil)
	case "trust":
		// Trust a player so they won't show as suspicious
		if index == -1 {
			err = errors.New("invalid player number")
			break
		}
		_, err = svc.playerAction("trust", players[index].PlayfabId, nil)
		log.Info("This action may take up to 15 minutes to apply globally")
		// Mark the player trusted on this client immediately
		localTrustList = append(localTrustList, players[index].PlayfabId)
	default:
		err = errors.New("command not recognized")
	}
	return
}
