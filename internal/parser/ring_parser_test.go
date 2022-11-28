package parser

import (
	"sync/atomic"
	"testing"
	"time"

	"parser/internal/timer"
	"parser/internal/urlcache"

	"github.com/stretchr/testify/require"
)

var (
	mockParseResult = NewParseResult("mock", 100.0, "url")
)

type NoOpParser struct{}

func (np *NoOpParser) Parse(timeout time.Duration, url string) *ParseResult {
	return mockParseResult
}

type NoOpUrlCacher struct{}

func (np *NoOpUrlCacher) Set(url string) {
	return
}

func (no *NoOpUrlCacher) ShouldParse(url string) bool {
	return true
}

func TestRingParser(t *testing.T) {
	t.Run("test can add", func(t *testing.T) {
		t.Parallel()

		ringParser := NewRingParser(&RingParserOptions{
			Parser:         new(NoOpParser),
			ParsingTimeout: 10,
			Timer:          new(timer.AppTimer),
			OutChanBuff:    2,
		})

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
		t.Parallel()

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

	t.Run("can use urlCache", func(t *testing.T) {
		t.Parallel()

		ringParser := rpWithURLs()
		// Replace NoOp for real impl.
		ringParser.urlCache = urlcache.NewUrlCache(time.Second * 10 /* cache TTL */)

		// Add simple reader from output chan
		var countUpdates int32
		go func() {
			for update := range ringParser.Out() {
				require.EqualValues(t, mockParseResult, update)
				atomic.AddInt32(&countUpdates, 1)
			}
		}()

		ringParser.Run(time.Millisecond * 100)

		time.Sleep(time.Second * 2)

		ringParser.Close()
		// Time slept is 2 seconds (2000ms).
		// Parsing interval is 100ms, theoretically 20 updates will reach reader.
		// Actually, we test urlCache here. Each url set to cache inside ringParser
		// would be parsed not more often than 10s (cache TTL).
		// So, expect only 5 updates. (len of ringParser.targets)
		require.Equal(t, int32(len(ringParser.targets)) /* expected count updates */, atomic.LoadInt32(&countUpdates))
	})

}

func rpWithURLs() *RingParser {

	ringParser := NewRingParser(&RingParserOptions{
		Parser:         new(NoOpParser),
		UrlCache:       new(NoOpUrlCacher),
		ParsingTimeout: 10,
		Timer:          timer.NewAppTimer(),
		OutChanBuff:    2,
	})

	urls := []string{
		"abcd",
		"efgh",
		"zxcv",
		"fhdia",
		"qiwnx",
	}

	for _, url := range urls {
		ringParser.AddTarget(url)
	}

	return ringParser
}
