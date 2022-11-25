package parser

import (
	"fmt"
	"parser/internal/timer"
	"sync"
	"time"
)

type RingParser struct {
	// List of URLs to parse
	targets []string
	// Keep track of what urls are targets. See RingParser.AddTarget
	urls   map[string]struct{}
	offset int

	// Protect data
	mu *sync.Mutex
	wg *sync.WaitGroup

	parser Parser
	// Control maximum time a single parsing process can take
	timeout time.Duration
	timer   timer.Timer

	out      chan *ParseResult
	shutdown chan struct{}
}

func NewRingParser(parser Parser, timer timer.Timer, parsingTimeout time.Duration, queueLength int) *RingParser {
	return &RingParser{
		parser:   parser,
		mu:       &sync.Mutex{},
		wg:       new(sync.WaitGroup),
		shutdown: make(chan struct{}),
		offset:   0,
		timer:    timer,
		targets:  make([]string, 0),
		urls:     make(map[string]struct{}),
		timeout:  parsingTimeout,
		out:      make(chan *ParseResult, queueLength),
	}
}

// Run spawns a gorotine that performs a parsing within an interval
func (rp *RingParser) Run(interval time.Duration) {
	go rp.timer.Every(interval, rp.parse)
}

func (rp *RingParser) Close() {
	rp.timer.Stop()
	// signal all parsing goroutines to end
	rp.wg.Wait()
	close(rp.out)
}

func (rp *RingParser) Out() <-chan *ParseResult {
	return rp.out
}

func (rp *RingParser) AddTarget(url string) {
	rp.mu.Lock()
	_, ok := rp.urls[url]
	rp.mu.Unlock()

	// URL is already being parsed
	if ok {
		return
	}

	// Add url
	rp.mu.Lock()
	rp.targets = append(rp.targets, url)
	rp.urls[url] = struct{}{}
	rp.mu.Unlock()

	fmt.Println("added: ", url)
}

func (rp *RingParser) parse() {

	rp.mu.Lock()
	targetlen := len(rp.targets)
	if targetlen == 0 {
		rp.mu.Unlock()
		return
	}

	// Mutex is locked here
	url := rp.targets[rp.offset]

	mx := targetlen - 1
	if rp.offset == mx {
		rp.offset = 0
	} else {
		rp.offset += 1
	}
	rp.mu.Unlock()

	rp.wg.Add(1)
	go func() {
		data := rp.parser.Parse(rp.timeout, url)
		rp.out <- data
		rp.wg.Done()
	}()
}
