package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"google.golang.org/api/idtoken"
	"io"
	"net/http"
	"time"
)

const (
	validateUrl = "https://europe-west3-prj-prd-chiv-01.cloudfunctions.net/func-prd-validate_players"
	banUrl      = "https://europe-west3-prj-prd-chiv-01.cloudfunctions.net/func-prd-ban_player"
)

type backendService struct {
	validateClient *http.Client
	banClient      *http.Client
}

// newBackendService creates an authenticated client for validation and banning
func newBackendService(credentialsPath string) (svc backendService, err error) {
	ctx := context.Background()
	credentialsFile := idtoken.WithCredentialsFile(credentialsPath)
	svc.validateClient, err = idtoken.NewClient(ctx, validateUrl, credentialsFile)
	if err != nil {
		err = fmt.Errorf("authentication failed: %w", err)
		return
	}
	svc.banClient, err = idtoken.NewClient(ctx, banUrl, credentialsFile)
	if err != nil {
		err = fmt.Errorf("authentication failed: %w", err)
		return
	}
	return
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

// validatePlayers sends a list of players to the validation endpoint and returns all information
func (svc backendService) validatePlayers(serverName string, players []connectedPlayer) (validatedPlayers []validatedPlayer, err error) {
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
	resp, err := svc.validateClient.Do(req)
	if err != nil {
		err = fmt.Errorf("call to validate backend failed: %w", err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("call to validate backend failed with status code %d", resp.StatusCode)
		return
	}

	respBody, _ := io.ReadAll(resp.Body)
	respData := struct {
		ValidatedPlayers []validatedPlayer `json:"validated_players"`
	}{}
	_ = json.Unmarshal(respBody, &respData)
	validatedPlayers = respData.ValidatedPlayers
	return
}

// banPlayer adds the given player to the wanted board and returns a pre-written ban command
func (svc backendService) banPlayer(playfabId string, charges []string) (banCommand string, err error) {
	reqParams := struct {
		PlayFabId string   `json:"playfab_id"`
		Charges   []string `json:"charges"`
	}{
		PlayFabId: playfabId,
		Charges:   charges,
	}
	body, _ := json.Marshal(reqParams)
	req, _ := http.NewRequest(http.MethodPost, banUrl, bytes.NewReader(body))
	resp, err := svc.banClient.Do(req)
	if err != nil {
		err = fmt.Errorf("call to ban backend failed: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("call to validate backend failed with status code %d", resp.StatusCode)
		return
	}

	respBody, _ := io.ReadAll(resp.Body)
	respData := struct {
		BanCommand string `json:"ban_command"`
	}{}
	_ = json.Unmarshal(respBody, &respData)
	banCommand = respData.BanCommand
	return
}
