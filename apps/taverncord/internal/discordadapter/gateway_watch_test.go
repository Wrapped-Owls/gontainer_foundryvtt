package discordadapter

import (
	"context"
	"errors"
	"testing"
	"testing/synctest"
	"time"
)

const (
	limit   = 5 * time.Minute
	horizon = 2 * limit
)

func TestGatewayWatchRun(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		drive       func(watch *gatewayWatch)
		wantErr     error
		wantElapsed time.Duration
	}{
		{name: "stays connected", drive: func(*gatewayWatch) {}, wantElapsed: horizon},
		{
			name: "reconnects within limit",
			drive: func(watch *gatewayWatch) {
				watch.onDisconnect(nil, nil)
				synctest.Sleep(limit / 2)
				watch.onConnect(nil, nil)
			},
			wantElapsed: horizon,
		},
		{
			name: "reconnect restarts the offline clock",
			drive: func(watch *gatewayWatch) {
				watch.onDisconnect(nil, nil)
				synctest.Sleep(limit / 2)
				watch.onConnect(nil, nil)
				watch.onDisconnect(nil, nil)
			},
			wantErr:     ErrGatewayUnreachable,
			wantElapsed: limit/2 + limit,
		},
		{
			name:        "stays disconnected",
			drive:       func(watch *gatewayWatch) { watch.onDisconnect(nil, nil) },
			wantErr:     ErrGatewayUnreachable,
			wantElapsed: limit,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			synctest.Test(t, func(t *testing.T) {
				ctx, cancel := context.WithTimeout(t.Context(), horizon)
				defer cancel()

				watch := &gatewayWatch{}
				start := time.Now()
				result := make(chan error, 1)
				go func() { result <- watch.run(ctx, limit) }()
				testCase.drive(watch)

				err := <-result
				if !errors.Is(err, testCase.wantErr) {
					t.Fatalf("err = %v, want %v", err, testCase.wantErr)
				}
				if elapsed := time.Since(start); elapsed != testCase.wantElapsed {
					t.Fatalf("elapsed = %v, want %v", elapsed, testCase.wantElapsed)
				}
			})
		})
	}
}
