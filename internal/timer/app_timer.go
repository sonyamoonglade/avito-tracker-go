package timer

import (
	"time"
)

type AppTimer struct {
	t *time.Ticker
}

func (at *AppTimer) Every(interval time.Duration, f func()) {

	if at.t == nil {
		at.t = time.NewTicker(interval)
	}

	for {
		<-at.t.C
		f()
	}
}

func (at *AppTimer) Stop() {
	at.t.Stop()
}
