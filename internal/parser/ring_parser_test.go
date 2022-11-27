package parser

import (
	"sync/atomic"
	"testing"
	"time"

	"parser/internal/timer"

	"github.com/stretchr/testify/require"
)

type NoOpParser struct{}

var mockParseResult = NewParseResult("mock", 100.0, "url")

func (np *NoOpParser) Parse(timeout time.Duration, url string) *ParseResult {
	return mockParseResult
}

func TestRingParser(t *testing.T) {
	t.Run("test can add", func(t *testing.T) {
		ringParser := NewRingParser(new(NoOpParser), timer.NewAppTimer(), time.Second*1 /* parsing timeout */, 1)
		urls := []string{
			"abcd",
			"efgh",
			"zxcv",
			"fhdia",
			"qiwnx",
		}

		for _, url := range urls {
			// prevents from adding same url twice so no effect
			// testing purposes...
			ringParser.AddTarget(url)
			ringParser.AddTarget(url)
		}

		require.True(t, len(ringParser.targets) == len(urls))
	})

	t.Run("test can run, read, gracefully close", func(t *testing.T) {
		ringParser := rpWithURLs()
		// Add simple reader from output chan
		var countUpdates int32
		go func() {
			for update := range ringParser.Out() {
				require.EqualValues(t, mockParseResult, update)
				atomic.AddInt32(&countUpdates, 1)
			}
		}()

		// Spawns a goroutine under the hood
		ringParser.Run(time.Second * 1 /* Parsing interval */)

		// Sleep for abit longer than 2000ms(2s) because of some overhead
		time.Sleep(time.Millisecond * 2005)

		ringParser.Close()
		// Can run two complete parsings.
		// ringParser.Run has interval - 1s
		// Time asleep - 2.005s =>
		// Two complete parsings could be performed
		require.Equal(t, int32(2) /* expected count updates */, atomic.LoadInt32(&countUpdates))
	})

}

func rpWithURLs() *RingParser {

	rp := NewRingParser(new(NoOpParser), timer.NewAppTimer(), time.Second*1 /* parsing timeout */, 1)
	urls := []string{
		"abcd",
		"efgh",
		"zxcv",
		"fhdia",
		"qiwnx",
	}

	for _, url := range urls {
		rp.AddTarget(url)
	}

	return rp
}
