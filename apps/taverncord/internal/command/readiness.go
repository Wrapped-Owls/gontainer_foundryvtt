package command

import (
	"context"
	"fmt"
	"time"
)

const confirmWindow = 90 * time.Second

func (pc *ProfileCommands) awaitReady(
	ctx context.Context,
	isReady func(StatusData) bool,
) (StatusData, error) {
	const confirmPoll = 3 * time.Second
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
			status, err := pc.client.Status(ctx)
			if err != nil {
				continue
			}
			if isReady(status) {
				return status, nil
			}
		}
	}
}

func (pc *ProfileCommands) confirm(
	ctx context.Context,
	r Responder,
	done string,
	isReady func(StatusData) bool,
) error {
	status, err := pc.awaitReady(ctx, isReady)
	if err != nil {
		pc.logger.Warn("foundry did not come back in time", "err", err)
		return r.Edit(ctx, fmt.Sprintf(
			"\u26a0\ufe0f %s, but the server has not answered in %s. Check `/foundry logs`.",
			done, confirmWindow,
		))
	}
	return r.Edit(ctx, fmt.Sprintf("✅ %s - 🟢 back online%s.", done, worldSuffix(status)))
}

func worldSuffix(status StatusData) string {
	if status.WorldActive && status.World != "" {
		return fmt.Sprintf(" on world `%s`", status.World)
	}
	return " at the setup screen"
}

func cycled(before StatusData) func(StatusData) bool {
	wasUp := before.Online && before.UptimeMS > 0
	return func(now StatusData) bool {
		return now.Online && (!wasUp || now.UptimeMS < before.UptimeMS)
	}
}
