package parser

import (
	"errors"
	"time"
)

var (
	ErrURLUnavailable = errors.New("URL is unavailable")
)

type Parser interface {
	Parse(timeout time.Duration, url string) *ParseResult
}

type TargetAdder interface {
	AddTarget(url string)
}

type ParseResult struct {
	url string

	title string
	price float64
	err   error

	// original html that was parsed
	raw *string
}

func NewParseResult(title string, price float64, URL string) *ParseResult {
	return &ParseResult{title: title, price: price, url: URL, err: nil, raw: nil}
}

func NewParseResultWithError(err error, raw *string) *ParseResult {
	return &ParseResult{title: "", price: 0.0, err: err, raw: raw}
}
func (pr *ParseResult) Title() string {
	return pr.title
}

func (pr *ParseResult) Price() float64 {
	return pr.price
}

func (pr *ParseResult) Err() error {
	return pr.err
}

func (pr *ParseResult) URL() string {
	return pr.url
}

func (pr *ParseResult) Raw() *string {
	return pr.raw
}
