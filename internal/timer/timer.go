package timer

import (
	"time"
)

type Timer interface {
	Every(interval time.Duration, f func())
	Stop()
}
