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
	timer  timer.Timer

	out chan *ParseResult
}

func NewRingParser(parser Parser, timer timer.Timer, initialTargetURLs ...string) *RingParser {
	if initialTargetURLs == nil {
		initialTargetURLs = make([]string, 0, 0)
	}

	urls := make(map[string]struct{}, len(initialTargetURLs))
	for _, url := range initialTargetURLs {
		urls[url] = struct{}{}
	}

	return &RingParser{
		parser:  parser,
		mu:      new(sync.Mutex),
		offset:  0,
		targets: initialTargetURLs,
		urls:    urls,
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
	_, ok := rp.urls[url]
	rp.mu.Unlock()

	// URL is already being parsed
	if ok {
		return
	}

	// Add url
	rp.mu.Lock()
	rp.targets = append(rp.targets, url)
	rp.mu.Unlock()
}

func (rp *RingParser) Out() chan *ParseResult {
	return rp.out
}

func (rp *RingParser) parse() {

	fmt.Println("start parsing...")

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

	fmt.Printf("%+v\n", result)
	rp.out <- &ParseResult{
		Title: result.Title,
		Price: result.Price,
		Err:   err,
	}
}
