package parser

type ParseResult struct {
	Title string
	Price float64
}

type Parser interface {
	Parse(url string) (*ParseResult, error)
}
