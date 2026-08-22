package command

import (
	"context"
	"fmt"
	"time"
)

const (
	// confirmWindow is generous: a cold start re-reads the world.
	confirmWindow = 90 * time.Second
	confirmPoll   = 3 * time.Second
)

// awaitReady polls because a 202 means queued, not that Foundry is back.
func (pc *ProfileCommands) awaitReady(
	ctx context.Context,
	isReady func(StatusData) bool,
) (StatusData, error) {
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, confirmWindow)
	defer cancel()

	ticker := time.NewTicker(confirmPoll)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return StatusData{}, fmt.Errorf("wait for foundry: %w", ctx.Err())
		case <-ticker.C:
			data, err := pc.client.Status(ctx)
			if err != nil {
				// Unreachable is the expected state mid-restart; keep waiting.
				continue
			}
			if isReady(data) {
				return data, nil
			}
		}
	}
}

// confirm edits the pending reply with what actually happened. done is past tense,
// e.g. "Switched to **alice**".
func (pc *ProfileCommands) confirm(
	ctx context.Context,
	r Responder,
	done string,
	isReady func(StatusData) bool,
) error {
	data, err := pc.awaitReady(ctx, isReady)
	if err != nil {
		pc.logger.Warn("foundry did not come back in time", "err", err)
		return r.Edit(ctx, fmt.Sprintf(
			"⚠️ %s, but the server has not answered in %s. Check `/foundry logs`.",
			done, confirmWindow,
		))
	}
	return r.Edit(ctx, fmt.Sprintf("✅ %s - 🟢 back online%s.", done, worldSuffix(data)))
}

func worldSuffix(data StatusData) string {
	if data.WorldActive && data.World != "" {
		return fmt.Sprintf(" on world `%s`", data.World)
	}
	return " at the setup screen"
}

// cycled reports that the process actually restarted, not that the outgoing one is
// still answering. Only a reset uptime proves it: the manager marks the new profile
// active before the child is spawned, so "online and active" is true too early.
func cycled(before StatusData) func(StatusData) bool {
	wasUp := before.Online && before.UptimeMS > 0
	return func(now StatusData) bool {
		return now.Online && (!wasUp || now.UptimeMS < before.UptimeMS)
	}
}
