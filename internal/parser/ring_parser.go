package parser

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"parser/internal/timer"
)

type RingParserOptions struct {
	Parser         Parser
	ParsingTimeout time.Duration
	Timer          timer.Timer
	// Buffer length of .Out() chan
	OutChanBuff int32
}

type RingParser struct {
	// List of URLs to parse
	targets   []string
	targetlen int32
	offset    int32

	// Keep track of what urls are targets. See RingParser.AddTarget
	urls map[string]struct{}

	// Protect data
	mu *sync.RWMutex

	parser Parser

	// Control maximum time a single parsing process can take
	timeout time.Duration
	timer   timer.Timer

	out      chan *ParseResult
	shutdown chan struct{}
}

func NewRingParser(opts *RingParserOptions) *RingParser {
	return &RingParser{
		offset:   0,
		parser:   opts.Parser,
		timer:    opts.Timer,
		timeout:  opts.ParsingTimeout,
		mu:       new(sync.RWMutex),
		targets:  make([]string, 0),
		urls:     make(map[string]struct{}),
		shutdown: make(chan struct{}),
		out:      make(chan *ParseResult, opts.OutChanBuff),
	}
}

// Run spawns a goroutine that performs a parsing within an interval
func (rp *RingParser) Run(interval time.Duration) {
	rp.timer.Every(interval, rp.parse)
}

func (rp *RingParser) Close() {
	// Stop emitting intervals
	rp.timer.Stop()

	// Trigger shutdown
	close(rp.shutdown)
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

	select {
	case <-rp.shutdown:
		rp.onClose()
		return
	default:
		rp.out <- rp.parser.Parse(rp.timeout, url)
	}
}

func (rp *RingParser) onClose() {
	close(rp.out)
}
