package discordadapter

import (
	"context"
	"errors"
	"sync/atomic"
	"time"

	"github.com/bwmarrin/discordgo"
)

var ErrGatewayUnreachable = errors.New("discord gateway unreachable")

type gatewayWatch struct {
	offlineSince atomic.Pointer[time.Time] // fed by events: a stuck reconnect holds the session lock
}

func (w *gatewayWatch) onConnect(*discordgo.Session, *discordgo.Connect) {
	w.offlineSince.Store(nil)
}

func (w *gatewayWatch) onDisconnect(*discordgo.Session, *discordgo.Disconnect) {
	now := time.Now()
	w.offlineSince.CompareAndSwap(nil, &now)
}

func (w *gatewayWatch) run(ctx context.Context, limit time.Duration) error {
	const checkInterval = 30 * time.Second
	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if since := w.offlineSince.Load(); since != nil && time.Since(*since) >= limit {
				return ErrGatewayUnreachable
			}
		}
	}
}
