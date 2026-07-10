# yearXprogress

A bot that posts an update on how much of the year has passed:

```
YearProgress - ▓▓▓▓▓▓▓▓▓░░░░░░░░░░░ 47%
```

It checks in every 3 hours, but only actually posts when the percentage has gone up since the last post. Runs as a one-shot binary triggered by a GitHub Actions cron schedule — no server, no database, nothing to host. Each run:

1. Computes how much of the current year has elapsed (leap-year aware, via Go's `time.YearDay()`)
2. Loads the last posted percentage from a GitHub Gist
3. Skips posting if that percentage hasn't changed
4. Posts the update to Bluesky
5. Saves the new percentage back to the Gist

## Why a Gist instead of a database?

Didn't want to keep a database running for one number storage. And committing the state back into this repo every few hours is too noisy. A Gist does the job fine- can read and update over the API, sitting in the side.

## Architecture

```
main.go
├── internal/progress     — year-progress % + bar rendering (no dependencies)
├── internal/state        — reads/writes state.json in a GitHub Gist
└── internal/blueskypost  — logs in and posts a status update to Bluesky
```
