// Package state persists the bot's last-posted year/percentage in a GitHub
// Gist, so the workflow has somewhere to remember state between runs without
// needing a database or committing bot activity into the project repo.
package state

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	apiBase  = "https://api.github.com/gists"
	fileName = "state.json"
)

// State is the last percentage the bot posted, so it can dedupe repeat runs
// within the same hour and detect year rollover.
type State struct {
	Year           int `json:"year"`
	LastPercentage int `json:"lastPercentage"`
}

// Store reads and writes State to a single file inside a GitHub Gist.
type Store struct {
	Token  string
	GistID string
	client *http.Client
}

func NewStore(token, gistID string) *Store {
	return &Store{Token: token, GistID: gistID, client: http.DefaultClient}
}

type gistResponse struct {
	Files map[string]struct {
		Content string `json:"content"`
	} `json:"files"`
}

// Load fetches the current state. A gist with no state file yet (first run)
// returns a zero-value State rather than an error.
func (s *Store) Load(ctx context.Context) (State, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiBase+"/"+s.GistID, nil)
	if err != nil {
		return State{}, err
	}
	s.setHeaders(req)

	resp, err := s.client.Do(req)
	if err != nil {
		return State{}, fmt.Errorf("fetching gist: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return State{}, err
	}
	if resp.StatusCode != http.StatusOK {
		return State{}, fmt.Errorf("gist GET returned %d: %s", resp.StatusCode, body)
	}

	var gr gistResponse
	if err := json.Unmarshal(body, &gr); err != nil {
		return State{}, fmt.Errorf("decoding gist response: %w", err)
	}

	file, ok := gr.Files[fileName]
	if !ok || file.Content == "" {
		return State{}, nil
	}

	var st State
	if err := json.Unmarshal([]byte(file.Content), &st); err != nil {
		return State{}, fmt.Errorf("decoding state.json: %w", err)
	}
	return st, nil
}

// Save writes the given state back to the gist, overwriting state.json.
func (s *Store) Save(ctx context.Context, st State) error {
	content, err := json.Marshal(st)
	if err != nil {
		return err
	}

	payload := map[string]any{
		"files": map[string]any{
			fileName: map[string]string{
				"content": string(content),
			},
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, apiBase+"/"+s.GistID, bytes.NewReader(body))
	if err != nil {
		return err
	}
	s.setHeaders(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("updating gist: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("gist PATCH returned %d: %s", resp.StatusCode, respBody)
	}
	return nil
}

func (s *Store) setHeaders(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+s.Token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "yearXprogress-bot")
}
