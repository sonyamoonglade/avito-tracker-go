package timer

import (
	"time"
)

type AppTimer struct {
	shutdown chan struct{}
}

func NewAppTimer() Timer {
	return &AppTimer{
		shutdown: make(chan struct{}),
	}
}

func (at *AppTimer) Every(interval time.Duration, f func()) {
	go func() {
		emitter := time.NewTicker(interval)
		for {
			select {
			// Exit emitting goroutine once reached shutdown state
			case <-at.shutdown:
				return
			// Emit an action (call func)
			case _, ok := <-emitter.C:
				if !ok {
					return
				}

				f()
			}
		}
	}()
}

func (at *AppTimer) Stop() {
	close(at.shutdown)
}
