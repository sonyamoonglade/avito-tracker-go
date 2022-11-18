package timer

import "time"

type Timer interface {
	On(interval time.Duration, f func())
}
