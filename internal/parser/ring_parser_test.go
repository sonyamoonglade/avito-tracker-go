package parser

import (
	"parser/internal/timer"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type NoOpParser struct{}

var mockParseResult = NewParseResult("mock", 100.0, "url")

func (np *NoOpParser) Parse(timeout time.Duration, url string) *ParseResult {
	return mockParseResult
}

func TestRingParser(t *testing.T) {
	timer := timer.NewAppTimer()
	ringParser := NewRingParser(new(NoOpParser), timer, time.Second*1, 1)
	t.Run("test can add", func(t *testing.T) {
		urls := []string{
			"abcd",
			"efgh",
			"zxcv",
			"fhdia",
			"qiwnx",
		}

		for _, url := range urls {
			// Prevents from adding same url twice
			ringParser.AddTarget(url)
			ringParser.AddTarget(url)
		}

		require.True(t, len(ringParser.targets) == len(urls))
	})

	t.Run("test can run, read, gracefully close", func(t *testing.T) {
		// Wait time is 2 seconds but parsing interval is 1 second.
		// Therefore, expect 2 updates

		// Add simple reader from output
		var countUpdates int
		wg := &sync.WaitGroup{}
		wg.Add(1)
		go func() {
			for update := range ringParser.Out() {
				t.Log(update)
				require.EqualValues(t, mockParseResult, update)
				countUpdates += 1
			}
			t.Log("done!\n")
			wg.Done()
		}()

		time.Sleep(time.Microsecond)

		ringParser.Run(time.Second * 1 /* Interval */)

		time.Sleep(time.Second * 2)
		ringParser.Close()

		wg.Wait()
		require.Equal(t, 2 /* expected count updates */, countUpdates)
	})

}
