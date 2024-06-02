package main

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"github.com/gen2brain/beeep"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"
)

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

// printTable nicely formats the list of validated players and adds coloring and audio clues
func printTable(validatedPlayers []validatedPlayer) {
	slices.SortFunc(validatedPlayers, func(a, b validatedPlayer) int {
		return strings.Compare(a.DisplayName, b.DisplayName)
	})

	maxDisplayNameLength := 0
	for _, player := range validatedPlayers {
		width := utf8.RuneCountInString(player.DisplayName)
		if width > maxDisplayNameLength {
			maxDisplayNameLength = width
		}
	}

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
		if player.WantedLevel == "wanted" {
			go beeep.Beep(beeep.DefaultFreq, 100)
		}
		fmt.Println(styles[player.WantedLevel].Render(strings.Join(lines, "\n")))
	}
	fmt.Println()
}
