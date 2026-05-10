package cmd

import (
	"context"
	"os"
	"os/signal"
)

func signalNotifyContext(
	parent context.Context,
	sig ...os.Signal,
) (context.Context, context.CancelFunc) {
	return signal.NotifyContext(parent, sig...)
}
