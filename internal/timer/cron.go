package timer

import (
	"time"

	"github.com/robfig/cron"
)

type Cron struct {
	cron *cron.Cron
}

func NewCronTimer() *Cron {
	return &Cron{
		cron: cron.New(),
	}
}

func (c *Cron) Every(interval time.Duration, f func()) {
	panic("")
}
