package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
)

var ErrUnauthorized = errors.New("unauthorized")
var ErrServiceUnavailable = errors.New("identity service unavailable")

type IdentityResponse struct {
	UserID string `json:"user_id"`
}

func ValidateToken(token string) (string, error) {
	identityURL := os.Getenv("IDENTITY_SERVICE_URL")
	if identityURL == "" {
		identityURL = "http://identity-service.local"
	}

	url := fmt.Sprintf("%s/validate", identityURL)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", ErrServiceUnavailable
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", ErrServiceUnavailable
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return "", ErrUnauthorized
	}
	if resp.StatusCode != http.StatusOK {
		return "", ErrServiceUnavailable
	}

	var identity IdentityResponse
	if err := json.NewDecoder(resp.Body).Decode(&identity); err != nil {
		return "", ErrServiceUnavailable
	}

	return identity.UserID, nil
}

func ExtractBearerToken(r *http.Request) (string, bool) {
	header := r.Header.Get("Authorization")
	if !strings.HasPrefix(header, "Bearer ") {
		return "", false
	}
	token := strings.TrimPrefix(header, "Bearer ")
	if token == "" {
		return "", false
	}
	return token, true
}
