package main

import "strings"

const (
	delimiter = " - "
)

// readPlayerList extracts all player information from the listplayers output
func readPlayerList(list string) (serverName string, players []connectedPlayer, err error) {
	lines := strings.Split(strings.ReplaceAll(list, "\r\n", "\n"), "\n")
	cutStart := strings.Index(lines[0], delimiter) + 3
	cutEnd := strings.LastIndex(lines[0], " ")
	serverName = lines[0][cutStart:cutEnd]

	players = make([]connectedPlayer, 0, len(lines)-2)
	for _, line := range lines[2:] {
		if line == "" {
			continue
		}
		split := strings.Split(line, delimiter)
		displayName := strings.Join(split[:len(split)-5], delimiter)
		playfabId := split[len(split)-5]
		if playfabId == "NULL" {
			continue
		}
		players = append(players, connectedPlayer{
			DisplayName: displayName,
			PlayfabId:   playfabId,
		})
	}
	return
}
