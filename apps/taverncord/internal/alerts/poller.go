package alerts

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/taverncord/internal/command"
)

type EventSource interface {
	Events(ctx context.Context, since int) (command.EventsData, error)
}

type Sender interface {
	SendMessage(channelID, content string) error
}

type Poller struct {
	source    EventSource
	sender    Sender
	channelID string
	interval  time.Duration
	logger    *slog.Logger
	cursor    int
}

func New(
	source EventSource,
	sender Sender,
	channelID string,
	interval time.Duration,
	logger *slog.Logger,
) *Poller {
	return &Poller{
		source:    source,
		sender:    sender,
		channelID: channelID,
		interval:  interval,
		logger:    logger,
	}
}

func (p *Poller) Run(ctx context.Context) {
	if _, next, err := p.fetch(ctx); err == nil {
		p.cursor = next
	}

	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.poll(ctx)
		}
	}
}

func (p *Poller) poll(ctx context.Context) {
	events, next, err := p.fetch(ctx)
	if err != nil {
		p.logger.Warn("alert poll failed", "err", err)
		return
	}
	for _, e := range events {
		if sendErr := p.sender.SendMessage(p.channelID, format(e)); sendErr != nil {
			p.logger.Warn("failed to post alert", "err", sendErr)
		}
	}
	p.cursor = next
}

func (p *Poller) fetch(ctx context.Context) ([]command.EventItem, int, error) {
	page, err := p.source.Events(ctx, p.cursor)
	if err != nil {
		return nil, p.cursor, err
	}
	return page.Events, page.Next, nil
}

func format(e command.EventItem) string {
	icon := "\u26a0\ufe0f"
	if e.Kind == "crash" {
		icon = "🔴"
	}
	return fmt.Sprintf("%s **Foundry %s**: %s", icon, e.Kind, e.Message)
}
