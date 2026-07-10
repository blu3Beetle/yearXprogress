// Command yearxprogress posts a year-progress update whenever the
// percentage changes, run as a one-shot process on a GitHub Actions schedule.
package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/blu3Beetle/yearXprogress/internal/blueskypost"
	"github.com/blu3Beetle/yearXprogress/internal/progress"
	"github.com/blu3Beetle/yearXprogress/internal/state"
)

// target is one platform the bot posts to.
type target struct {
	name string
	post func(ctx context.Context, text string) (string, error)
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	bskyCreds := blueskypost.Credentials{
		Handle:      mustEnv("BLUESKY_HANDLE"),
		AppPassword: mustEnv("BLUESKY_APP_PASSWORD"),
	}
	targets := []target{
		{
			name: "bluesky",
			post: func(ctx context.Context, text string) (string, error) {
				return blueskypost.Post(ctx, bskyCreds, text)
			},
		},
	}

	store := state.NewStore(mustEnv("GIST_TOKEN"), mustEnv("GIST_ID"))

	snap := progress.Compute(time.Now())

	last, err := store.Load(ctx)
	if err != nil {
		log.Fatalf("loading state from gist: %v", err)
	}

	if last.Year == snap.Year && last.LastPercentage == snap.Percentage {
		log.Printf("already posted %d%% for %d, skipping", snap.Percentage, snap.Year)
		return
	}

	succeeded := 0
	for _, t := range targets {
		id, err := t.post(ctx, snap.Text)
		if err != nil {
			log.Printf("%s: post failed: %v", t.name, err)
			continue
		}
		log.Printf("%s: posted %s", t.name, id)
		succeeded++
	}

	if succeeded == 0 {
		log.Fatalf("all targets failed, not saving state so this percentage is retried next run")
	}

	newState := state.State{Year: snap.Year, LastPercentage: snap.Percentage}
	if err := store.Save(ctx, newState); err != nil {
		log.Fatalf("posted but failed to save state: %v", err)
	}
	log.Printf("state saved: year=%d percentage=%d", newState.Year, newState.LastPercentage)
}

func mustEnv(name string) string {
	v := os.Getenv(name)
	if v == "" {
		log.Fatalf("missing required environment variable %s", name)
	}
	return v
}
