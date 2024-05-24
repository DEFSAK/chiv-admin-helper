package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"google.golang.org/api/idtoken"
	"io"
	"log"
	"net/http"
	"os"
	"slices"
	"strings"
)

var (
	validateClient *http.Client
	banClient      *http.Client
)

func init() {
	var err error
	ctx := context.Background()
	// figure out credential location
	path := ""
	if len(os.Args) > 1 {
		path = os.Args[1]
	} else {
		fmt.Print("Enter the path to your credentials file and press enter: ")
		_, _ = fmt.Scan(&path)
	}
	// login with credentials
	validateClient, err = idtoken.NewClient(ctx, validateUrl, idtoken.WithCredentialsFile(path))
	if err != nil {
		log.Fatal(err)
	}
	banClient, err = idtoken.NewClient(ctx, banUrl, idtoken.WithCredentialsFile(path))
	if err != nil {
		log.Fatal(err)
	}
}

// validatePlayers sends a list of players to the validation endpoint and returns all information sorted by display name
func validatePlayers(serverName string, players []connectedPlayer) (validatedPlayers []validatedPlayer, err error) {
	reqParams := struct {
		CheckWantedBoard bool              `json:"check_wanted_board"`
		ServerName       string            `json:"server_name"`
		Players          []connectedPlayer `json:"players"`
	}{
		CheckWantedBoard: true,
		ServerName:       serverName,
		Players:          players,
	}
	body, _ := json.Marshal(reqParams)
	req, _ := http.NewRequest(http.MethodPost, validateUrl, bytes.NewReader(body))
	resp, _ := validateClient.Do(req)
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("call to validation backend failed with status code %d", resp.StatusCode)
		return
	}
	respBody, _ := io.ReadAll(resp.Body)

	respData := outputSchema{}
	_ = json.Unmarshal(respBody, &respData)
	validatedPlayers = respData.ValidatedPlayers
	slices.SortFunc(validatedPlayers, func(a, b validatedPlayer) int {
		return strings.Compare(a.DisplayName, b.DisplayName)
	})
	return
}

// banPlayer adds the given player to the wanted board and returns a pre-written ban command
func banPlayer(playfabId string, charges []string) (banCommand string, err error) {
	reqParams := struct {
		PlayFabId string   `json:"playfab_id"`
		Charges   []string `json:"charges"`
	}{
		PlayFabId: playfabId,
		Charges:   charges,
	}
	body, _ := json.Marshal(reqParams)
	req, _ := http.NewRequest(http.MethodPost, banUrl, bytes.NewReader(body))
	resp, _ := banClient.Do(req)
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("failed to ban player: permission denied")
		return
	}

	respData := outputSchema{}
	_ = json.Unmarshal(respBody, &respData)
	banCommand = respData.BanCommand
	return
}
