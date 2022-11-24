package parser

import "fmt"

type UpdateHandler func(result *ParseResult) error

type Proxy struct {
	rcvq          <-chan *ParseResult
	updateHandler UpdateHandler
	onError       func(err error)
}

func NewProxy(rcvq <-chan *ParseResult, updateHandler UpdateHandler, onError func(err error)) *Proxy {
	return &Proxy{rcvq: rcvq, updateHandler: updateHandler, onError: onError}
}

// Run starts listening to rcvq and execute handler
func (p *Proxy) Run() {
	for update := range p.rcvq {
		fmt.Printf("proxy rsv: %+v\n", update)
		err := p.updateHandler(update)
		if err != nil {
			p.onError(err)
		}
	}
}
