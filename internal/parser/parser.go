package parser

import (
	"time"
)

type ParseResult struct {
	Title string
	Price float64
	Err   error
}

func (pr *ParseResult) Error() error {
	return pr.Err
}

type Parser interface {
	Parse(timeout time.Duration, url string) (*ParseResult, error)
}
