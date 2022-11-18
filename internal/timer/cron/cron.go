package cron

import (
	"time"

	gocron "github.com/robfig/cron"
)

type Cron struct {
	cron *gocron.Cron
}

func NewCronManager() *Cron {
	return &Cron{
		cron: gocron.New(),
	}
}

func (c *Cron) On(interval time.Duration, f func()) {
	gocron.New().AddFunc("@every"+interval.String(), f)
}
