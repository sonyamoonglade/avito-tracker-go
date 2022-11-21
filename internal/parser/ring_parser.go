package parser

import (
	"parser/internal/timer"
	"sync"
	"time"
)

type RingParser struct {
	// List of URLs to parse
	targets []string
	offset  int

	mu     *sync.Mutex
	parser Parser

	timer timer.Timer

	out chan *ParseResult
}

func NewRingParser(parser Parser, timer timer.Timer, initialTargetURLs ...string) *RingParser {
	if initialTargetURLs == nil {
		initialTargetURLs = make([]string, 0, 0)
	}

	return &RingParser{
		parser:  parser,
		mu:      new(sync.Mutex),
		offset:  0,
		targets: initialTargetURLs,
		timer:   timer,
		out:     make(chan *ParseResult),
	}
}

func (rp *RingParser) Run(interval time.Duration) {
	rp.timer.Every(interval, rp.parse)
}

func (rp *RingParser) Stop() {
	close(rp.out)
	rp.timer.Stop()
}

func (rp *RingParser) AddTarget(url string) {
	rp.mu.Lock()
	rp.targets = append(rp.targets, url)
	rp.mu.Unlock()
}

func (rp *RingParser) Out() chan *ParseResult {
	return rp.out
}

func (rp *RingParser) parse() {

	rp.mu.Lock()
	url := rp.targets[rp.offset]

	mx := len(rp.targets) - 1
	if rp.offset == mx {
		rp.offset = 0
	} else {
		rp.offset += 1
	}
	rp.mu.Unlock()

	// move timeout from there
	result, err := rp.parser.Parse(time.Second*10, url)
	if err != nil {
		rp.out <- &ParseResult{
			Title: "",
			Price: 0.0,
			Err:   err,
		}

		return
	}

	rp.out <- &ParseResult{
		Title: result.Title,
		Price: result.Price,
		Err:   err,
	}
}
