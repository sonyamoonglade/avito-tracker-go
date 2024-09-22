package parser

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"parser/internal/timer"
	"parser/internal/urlcache"
)

type RingParserOptions struct {
	Parser         Parser
	UrlCache       urlcache.UrlCacher
	ParsingTimeout time.Duration
	Timer          timer.Timer
	// Buffer length of .Out() chan
	OutChanBuff int32
}

type RingParser struct {
	// List of URLs to parse.
	// Main operands to perform ring operations
	targets []string
	offset  int32

	targetlen int32

	// Keep track of what urls are targets. See RingParser.AddTarget
	urls map[string]struct{}

	// Protect data
	mu *sync.RWMutex

	parser Parser

	// Why we need UrlCache?
	//
	// Both Avito and Wildberries have caching systems.
	// When advert's price is updated Program sees it and notifies subscribers.
	// On the next iteration Program might get older web-page (cached one).
	// If old web-page is parsed, Program will notify the reverse change in price.
	// We have to avoid that duplication and therefore should use UrlCache.
	// In the result, Program will wait between parsing the same advert (url) the TTL defined in urlCache
	//
	// Example:
	// Example avert has price = 1000;
	// Iteration 1;
	// At some time it changes to -> 800: Program notifies change from 1000 to 800.
	// Iteration 2;
	// ... next iteration: gets page from cache;
	// Example advert again has price = 1000 (cached page), as a result of parsing. (Iteration 2)
	// Program notifies change from 800 (updated in db after Iteration 1) to 1000 (price in cached page)
	urlCache urlcache.UrlCacher

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
		urlCache: opts.UrlCache,
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

	// Get url to parse
	rp.mu.RLock()
	url := rp.targets[rp.offset]
	rp.mu.RUnlock()

	// Beforehand check if url should be parsed
	should := rp.urlCache.ShouldParse(url)
	if !should {
		fmt.Println("hitting cache")
		return
	}

	mx := targetlen - 1
	// If reached end of targets
	// rp.offset == mx
	if atomic.LoadInt32(&rp.offset) == mx {
		// Move offset ptr to start of array
		// rp.offset = 0
		atomic.StoreInt32(&rp.offset, 0)
	} else {
		// Next item
		// rp.offset += 1
		atomic.AddInt32(&rp.offset, 1)
	}

	select {
	case <-rp.shutdown:
		rp.onClose()
		return
	default:
		rp.out <- rp.parser.Parse(rp.timeout, url)
		// After successful parsing cache the url
		rp.urlCache.Set(url)
	}
}

func (rp *RingParser) onClose() {
	close(rp.out)
}
