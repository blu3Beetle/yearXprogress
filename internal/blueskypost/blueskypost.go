// Package blueskypost posts a status update to Bluesky via the AT Protocol.
// Each run logs in fresh with a handle + app password; there's no long-lived
// token to store between runs.
package blueskypost

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	createSessionURL = "https://bsky.social/xrpc/com.atproto.server.createSession"
	createRecordURL  = "https://bsky.social/xrpc/com.atproto.repo.createRecord"
)

// Credentials is a handle plus an app password (Settings -> App Passwords),
// not the account's real login password.
type Credentials struct {
	Handle      string
	AppPassword string
}

type sessionResponse struct {
	AccessJwt string `json:"accessJwt"`
	Did       string `json:"did"`
}

type createRecordRequest struct {
	Repo       string     `json:"repo"`
	Collection string     `json:"collection"`
	Record     postRecord `json:"record"`
}

type postRecord struct {
	Type      string `json:"$type"`
	Text      string `json:"text"`
	CreatedAt string `json:"createdAt"`
}

func Post(ctx context.Context, creds Credentials, text string) (string, error) {
	session, err := login(ctx, creds)
	if err != nil {
		return "", fmt.Errorf("bluesky login: %w", err)
	}

	body, err := json.Marshal(createRecordRequest{
		Repo:       session.Did,
		Collection: "app.bsky.feed.post",
		Record: postRecord{
			Type:      "app.bsky.feed.post",
			Text:      text,
			CreatedAt: time.Now().UTC().Format(time.RFC3339),
		},
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, createRecordURL, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+session.AccessJwt)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("creating bluesky post: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bluesky api error (status %d): %s", resp.StatusCode, respBody)
	}

	var result struct {
		URI string `json:"uri"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("decoding bluesky response: %w", err)
	}
	return result.URI, nil
}

func login(ctx context.Context, creds Credentials) (*sessionResponse, error) {
	body, err := json.Marshal(map[string]string{
		"identifier": creds.Handle,
		"password":   creds.AppPassword,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, createSessionURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("login failed (status %d): %s", resp.StatusCode, respBody)
	}

	var session sessionResponse
	if err := json.Unmarshal(respBody, &session); err != nil {
		return nil, fmt.Errorf("decoding session response: %w", err)
	}
	return &session, nil
}
