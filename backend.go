package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"google.golang.org/api/idtoken"
	"io"
	"net/http"
	"slices"
	"time"
)

const (
	validateUrl = "https://europe-west3-prj-prd-chiv-01.cloudfunctions.net/func-prd-validate_players"
	actionUrl   = "https://europe-west3-prj-prd-chiv-01.cloudfunctions.net/func-prd-player_action"
)

type backendService struct {
	validateClient *http.Client
	actionClient   *http.Client
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
	svc.actionClient, err = idtoken.NewClient(ctx, actionUrl, credentialsFile)
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
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		err = fmt.Errorf("call to validate backend failed: %s", string(respBody))
		return
	}

	respData := struct {
		ValidatedPlayers []validatedPlayer `json:"validated_players"`
	}{}
	_ = json.Unmarshal(respBody, &respData)
	validatedPlayers = respData.ValidatedPlayers
	for i, player := range validatedPlayers {
		if slices.Contains(localTrustList, player.PlayfabId) && player.WantedLevel == "suspicious" {
			validatedPlayers[i].WantedLevel = ""
			validatedPlayers[i].BanCommand = ""
		}
	}
	return
}

// playerAction executes an action that targets a single player. For example banning, unbanning or trusting.
// These action may result in a command that should be run on the server.
func (svc backendService) playerAction(action, playfabId string, params map[string]any) (outputCommand string, err error) {
	reqParams := struct {
		Action     string         `json:"action"`
		PlayFabId  string         `json:"playfab_id"`
		Parameters map[string]any `json:"parameters"`
	}{
		Action:     action,
		PlayFabId:  playfabId,
		Parameters: params,
	}
	body, _ := json.Marshal(reqParams)
	req, _ := http.NewRequest(http.MethodPost, actionUrl, bytes.NewReader(body))
	resp, err := svc.actionClient.Do(req)
	if err != nil {
		err = fmt.Errorf("call to player action backend failed: %w", err)
	}
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		err = fmt.Errorf("call to player action backend failed: %s", string(respBody))
		return
	}

	respData := struct {
		OutputCommand string `json:"output_command"`
	}{}
	_ = json.Unmarshal(respBody, &respData)
	outputCommand = respData.OutputCommand
	return
}
