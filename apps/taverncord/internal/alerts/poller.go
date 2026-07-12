package alerts

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/taverncord/internal/command"
)

// EventSource fetches detected events at or after a cursor.
type EventSource interface {
	Events(ctx context.Context, since int) (command.EventsData, error)
}

// Sender posts a message to a Discord channel.
type Sender interface {
	SendMessage(channelID, content string) error
}

// Poller periodically fetches events from the manager and forwards new ones to a
// Discord channel. It is safe to run in its own goroutine and stops on ctx.Done.
type Poller struct {
	source    EventSource
	sender    Sender
	channelID string
	interval  time.Duration
	logger    *slog.Logger
	cursor    int
}

// New builds a Poller that polls source every interval and posts to channelID.
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

// Run polls until ctx is cancelled. It primes the cursor on start so only events
// occurring after startup are announced, then posts each new event.
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
		if err := p.sender.SendMessage(p.channelID, format(e)); err != nil {
			p.logger.Warn("failed to post alert", "err", err)
		}
	}
	p.cursor = next
}

func (p *Poller) fetch(ctx context.Context) ([]command.EventItem, int, error) {
	data, err := p.source.Events(ctx, p.cursor)
	if err != nil {
		return nil, p.cursor, err
	}
	return data.Events, data.Next, nil
}

func format(e command.EventItem) string {
	icon := "⚠️"
	if e.Kind == "crash" {
		icon = "🔴"
	}
	return fmt.Sprintf("%s **Foundry %s**: %s", icon, e.Kind, e.Message)
}
