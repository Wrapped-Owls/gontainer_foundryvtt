package command

import (
	"context"
	"fmt"
)

func (pc *ProfileCommands) Restart(ctx context.Context, r Responder, interrupt Interrupt) error {
	before, err := pc.client.Status(ctx)
	if err != nil {
		pc.logger.Warn("status before restart failed", "err", err)
	}

	if err = r.Send(ctx, "⏳ Restarting Foundry...", Public); err != nil {
		return err
	}
	if err = pc.client.Restart(ctx, interrupt); err != nil {
		pc.logger.Error("restart failed", "err", err)
		return r.Edit(ctx, fmt.Sprintf("❌ Restart failed: %s", err.Error()))
	}

	return pc.confirm(ctx, r, "Restarted", cycled(before))
}
