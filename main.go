package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"github.com/gen2brain/beeep"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	delimiter   = " - "
	validateUrl = "https://europe-west3-prj-prd-chiv-01.cloudfunctions.net/func-prd-validate_players"
	banUrl      = "https://europe-west3-prj-prd-chiv-01.cloudfunctions.net/func-prd-ban_player"
)

type outputSchema struct {
	ValidatedPlayers []validatedPlayer `json:"validated_players"`
	BanCommand       string            `json:"ban_command"`
}

type validatedPlayer struct {
	PlayfabId   string    `json:"playfab_id"`
	DisplayName string    `json:"display_name"`
	Aliases     []string  `json:"aliases"`
	CreatedAt   time.Time `json:"created_at"`
	Platform    string    `json:"platform"`
	BanCommand  string    `json:"ban_command"`
	WantedFor   []string  `json:"wanted_for"`
	WantedLevel string    `json:"wanted_level"`
}

type connectedPlayer struct {
	DisplayName string `json:"display_name"`
	PlayfabId   string `json:"playfab_id"`
}

var platforms = map[string]string{
	"unknown": "X",
	"console": "G",
	"PC":      " ",
}

var styles = map[string]lipgloss.Style{
	"":           lipgloss.NewStyle(),
	"suspicious": lipgloss.NewStyle().Background(lipgloss.Color("#FFA500")).Foreground(lipgloss.Color("#000000")),
	"wanted":     lipgloss.NewStyle().Background(lipgloss.Color("#FF0000")),
}

func main() {
	ctx := context.Background()
	stdinEvents := watchStdin(ctx)
	clipboardEvents := watchClipboard(ctx, time.Millisecond*50)
	var validatedPlayers []validatedPlayer
	fmt.Println("chiv-admin-helper is ready. Use the listplayers command to validate players. Press ctrl+c to abort.")
	for {
		select {
		case event := <-stdinEvents:
			// Read command from stdin
			fmt.Printf("%q\n", event)
			split := strings.SplitN(event, " ", 3)
			nr, _ := strconv.Atoi(split[1])
			switch split[0] {
			case "kick":
				fmt.Printf("kickbyid %s \"\"\n", validatedPlayers[nr].PlayfabId)
			case "ban":
				charges := strings.Split(split[2], " ")
				banCommand, err := banPlayer(validatedPlayers[nr].PlayfabId, charges)
				if err != nil {
					fmt.Println(err)
				} else {
					fmt.Println(banCommand)
				}
			}
		case event := <-clipboardEvents:
			// Read listplayers from clipboard
			if strings.HasPrefix(event, "ServerName - ") {
				serverName, players := readPlayerList(event)
				validatedPlayers, _ = validatePlayers(serverName, players)
				printTable(validatedPlayers)
			}
		}
	}
}

// readPlayerList extracts all player information from the listplayers output
func readPlayerList(list string) (serverName string, players []connectedPlayer) {
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

// printTable nicely formats the list of validated players and adds coloring and audio clues
func printTable(validatedPlayers []validatedPlayer) {
	maxDisplayNameLength := 0
	for _, player := range validatedPlayers {
		if len(player.DisplayName) > maxDisplayNameLength {
			maxDisplayNameLength = len(player.DisplayName)
		}
	}

	fmt.Println()
	for i, player := range validatedPlayers {
		aliases := strings.Join(player.Aliases, ", ")
		lines := make([]string, 1)
		lines[0] = fmt.Sprintf(
			"%2d)  %-16s  %s  %1s  %-"+strconv.Itoa(maxDisplayNameLength)+"s (%s)",
			i,
			player.PlayfabId,
			player.CreatedAt.Format("2006-01-02 15:04"),
			platforms[player.Platform],
			player.DisplayName,
			aliases,
		)
		if len(player.WantedFor) > 0 {
			lines = append(lines, "Wanted for: "+strings.Join(player.WantedFor, ", "))
		}
		if player.BanCommand != "" {
			lines = append(lines, player.BanCommand)
		}
		fmt.Println(styles[player.WantedLevel].Render(strings.Join(lines, "\n")))
		if player.WantedLevel != "" {
			go beeep.Beep(beeep.DefaultFreq, beeep.DefaultDuration)
			time.Sleep(time.Millisecond * 100)
		}
	}
	fmt.Println()
}

// watchStdin monitors console input and notifies the channel when a new command is received
func watchStdin(ctx context.Context) (events chan string) {
	events = make(chan string)
	stdin := bufio.NewReader(os.Stdin)
	go func() {
		for {
			if ctx.Err() == nil {
				command, _ := stdin.ReadString('\n')
				command = strings.TrimRight(command, "\r\n")
				if command == "" {
					continue
				}
				events <- command
			} else {
				close(events)
			}
		}
	}()
	return
}
