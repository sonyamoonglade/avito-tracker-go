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

	parser Parser
	// Control maximum time a single parsing process can take
	timeout time.Duration
	timer   timer.Timer

	out chan *ParseResult
}

func NewRingParser(parser Parser, timer timer.Timer, parsingTimeout time.Duration, queueLength int) *RingParser {
	return &RingParser{
		parser:  parser,
		mu:      &sync.Mutex{},
		offset:  0,
		targets: make([]string, 0),
		urls:    make(map[string]struct{}),
		timer:   timer,
		timeout: parsingTimeout,
		out:     make(chan *ParseResult, queueLength),
	}
}

func (rp *RingParser) Run(interval time.Duration) {
	rp.timer.Every(interval, rp.parse)
}

func (rp *RingParser) Close() {
	close(rp.out)
	rp.timer.Stop()
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

	rp.out <- rp.parser.Parse(rp.timeout, url)
}
