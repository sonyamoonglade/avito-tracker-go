package parser

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"parser/internal/timer"
)

type RingParser struct {
	// List of URLs to parse
	targets   []string
	targetlen int32

	// Keep track of what urls are targets. See RingParser.AddTarget
	urls   map[string]struct{}
	offset int32

	// Protect data
	mu *sync.RWMutex
	wg *sync.WaitGroup

	parser Parser

	// Control maximum time a single parsing process can take
	timeout time.Duration
	timer   timer.Timer

	// 1 - running
	// 0 - stopped
	state uint32

	out      chan *ParseResult
	shutdown chan struct{}
}

func NewRingParser(parser Parser, timer timer.Timer, parsingTimeout time.Duration, queueLength int) *RingParser {
	return &RingParser{
		parser:   parser,
		timer:    timer,
		timeout:  parsingTimeout,
		offset:   0,
		mu:       new(sync.RWMutex),
		wg:       new(sync.WaitGroup),
		targets:  make([]string, 0),
		state:    1,
		urls:     make(map[string]struct{}),
		shutdown: make(chan struct{}),
		out:      make(chan *ParseResult, queueLength),
	}
}

// Run spawns a goroutine that performs a parsing within an interval
func (rp *RingParser) Run(interval time.Duration) {
	rp.timer.Every(interval, rp.parse)
}

func (rp *RingParser) Close() {
	// Set state to 'stopped'
	atomic.StoreUint32(&rp.state, 0)

	// Wait for all parsing goroutines to end
	rp.wg.Wait()
	// Signal that no more updates are sent
	close(rp.out)

	// Stop emitting intervals
	rp.timer.Stop()

}

func (rp *RingParser) Out() <-chan *ParseResult {
	return rp.out
}

func (rp *RingParser) AddTarget(url string) {
	rp.mu.RLock()
	_, ok := rp.urls[url]
	rp.mu.RUnlock()

	// URL is already being parsed
	if ok {
		return
	}

	// Add url
	rp.mu.Lock()
	rp.targets = append(rp.targets, url)
	rp.urls[url] = struct{}{}
	rp.mu.Unlock()

	atomic.AddInt32(&rp.targetlen, 1)
	fmt.Println("added: ", url)
}

func (rp *RingParser) parse() {

	targetlen := atomic.LoadInt32(&rp.targetlen)
	if targetlen == 0 {
		return
	}

	rp.mu.RLock()
	url := rp.targets[rp.offset]
	rp.mu.RUnlock()

	mx := targetlen - 1
	// rp.offset == mx
	if atomic.LoadInt32(&rp.offset) == mx {
		// rp.offset = 0
		atomic.StoreInt32(&rp.offset, 0)
	} else {
		// rp.offset += 1
		atomic.AddInt32(&rp.offset, 1)
	}

	if !rp.checkState() {
		return
	}

	// Access rp.parser there because otherwise it will Data Race
	parser := rp.parser
	rp.wg.Add(1)
	go func() {
		// rp.parser !!! DATA RACE
		data := parser.Parse(rp.timeout, url)
		if !rp.checkState() {
			// Return gracefully
			// Do not write on potentially closed rp.out channel
			rp.wg.Done()
			return
		}
		rp.out <- data
		rp.wg.Done()
	}()

}

func (rp *RingParser) checkState() bool {
	// Check if RingParser.Close() has been called. See RingParser.Close
	if atomic.LoadUint32(&rp.state) == 0 {
		// Stopped
		return false
	}

	return true
}
