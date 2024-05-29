package main

import (
	"bufio"
	"context"
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

func executeCommand(command string, players []validatedPlayer) (err error) {

	return
}
