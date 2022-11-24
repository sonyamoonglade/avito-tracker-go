package parser

import (
	"time"
)

type Parser interface {
	Parse(timeout time.Duration, url string) *ParseResult
}

type ParseResult struct {
	url string

	title string
	price float64
	err   error
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

func NewParseResult(title string, price float64, URL string) *ParseResult {
	return &ParseResult{title: title, price: price, url: URL, err: nil}
}

func NewParseResultWithError(err error) *ParseResult {
	return &ParseResult{title: "", price: 0.0, err: err}
}
