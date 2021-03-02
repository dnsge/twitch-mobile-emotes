package session

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

const twitchValidateEndpoint = "https://id.twitch.tv/oauth2/validate"

var httpClient = http.Client{
	Timeout: time.Second * 5,
}

type validateResponse struct {
	UserID string `json:"user_id"`
}

func GetUserIDFromOAuth(oauth string) (string, error) {
	req, err := http.NewRequest("GET", twitchValidateEndpoint, nil)
	if err != nil {
		return "", err
	}

	token := strings.TrimPrefix(oauth, "oauth:")
	req.Header.Set("Authorization", "OAuth "+token)
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	var v validateResponse
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return "", err
	}

	return v.UserID, nil
}
