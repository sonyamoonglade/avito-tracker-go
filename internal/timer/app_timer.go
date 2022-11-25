package timer

import (
	"time"
)

type AppTimer struct {
	shutdown chan struct{}
	t        *time.Ticker
}

func NewAppTimer() Timer {
	return &AppTimer{
		shutdown: make(chan struct{}),
	}
}

func (at *AppTimer) Every(interval time.Duration, f func()) {

	if at.t == nil {
		at.t = time.NewTicker(interval)
	}

	for {
		select {
		case <-at.shutdown:
			at.t.Stop()
			return
		case <-at.t.C:
			f()
		}
	}
}

func (at *AppTimer) Stop() {
	close(at.shutdown)
	at.t.Stop()
}
