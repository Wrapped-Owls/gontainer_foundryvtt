package alerts

import (
	"context"
	"log/slog"
	"sync"
	"testing"
	"testing/synctest"
	"time"

	"github.com/wrapped-owls/gontainer_foundryvtt/apps/taverncord/internal/command"
)

type fakeSource struct {
	mu   sync.Mutex
	data command.EventsData
	err  error
}

func (f *fakeSource) Events(_ context.Context, since int) (command.EventsData, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.err != nil {
		return command.EventsData{}, f.err
	}
	if since >= f.data.Next {
		return command.EventsData{Next: f.data.Next}, nil
	}
	return f.data, nil
}

func (f *fakeSource) set(d command.EventsData) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.data = d
}

type fakeSender struct {
	mu   sync.Mutex
	msgs []string
}

func (f *fakeSender) SendMessage(_ string, content string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.msgs = append(f.msgs, content)
	return nil
}

func (f *fakeSender) count() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.msgs)
}

func TestPoller_announcesOnlyNewEvents(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		source := &fakeSource{data: command.EventsData{
			Events: []command.EventItem{{Kind: "error", Message: "old"}},
			Next:   1,
		}}
		sender := &fakeSender{}
		ctx, cancel := context.WithCancel(context.Background())

		p := New(source, sender, "chan", time.Second, slog.New(slog.DiscardHandler))
		done := make(chan struct{})
		go func() {
			p.Run(ctx)
			close(done)
		}()

		// After priming, the pre-existing event must not be announced.
		synctest.Wait()
		if sender.count() != 0 {
			t.Fatalf("priming should not announce existing events, got %d", sender.count())
		}

		// A new crash arrives; the next tick announces it exactly once.
		source.set(command.EventsData{
			Events: []command.EventItem{{Kind: "crash", Message: "boom"}},
			Next:   2,
		})
		time.Sleep(time.Second)
		synctest.Wait()
		if sender.count() != 1 {
			t.Fatalf("expected one alert, got %d", sender.count())
		}

		// A further tick with no new events must not re-announce.
		time.Sleep(time.Second)
		synctest.Wait()
		if sender.count() != 1 {
			t.Fatalf("expected no re-announce, got %d", sender.count())
		}

		cancel()
		<-done
	})
}
