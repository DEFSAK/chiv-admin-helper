package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

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

func executeCommand(command string, players []validatedPlayer, svc backendService) (err error) {
	rd := strings.NewReader(command)
	var action string
	var index int
	n, err := fmt.Fscan(rd, &action, &index)
	if err != nil || n != 2 {
		err = errors.New("invalid command format")
		return
	}
	if index < 0 || index >= len(players) {
		err = errors.New("invalid player number")
		return
	}
	switch action {
	case "kick":
		// Generate a one time kick command
		fmt.Printf("kickbyid %s\n", players[index].PlayfabId)
	case "ban":
		// Ban a player globally
		charges := make([]string, 0, 1)
		var charge, banCommand string
		for {
			n, err = fmt.Fscan(rd, &charge)
			if err != nil || n != 1 {
				if errors.Is(err, io.EOF) {
					err = nil
				}
				break
			} else {
				charges = append(charges, charge)
			}
		}
		banCommand, err = svc.banPlayer(players[index].PlayfabId, charges)
		fmt.Println(banCommand)
	default:
		err = errors.New("command not recognized")
	}
	return
}
