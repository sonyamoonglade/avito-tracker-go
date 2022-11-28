package proxy

import (
	goerrors "errors"
	"fmt"
	"parser/internal/domain/services"
	"parser/internal/errors"
	"parser/internal/parser"
)

// Proxy handles output from `rcvq` and handles it via `updateHandler`
// On error the callback `onError` is executed
type Proxy struct {
	rcvq          <-chan *parser.ParseResult
	shutdown      chan struct{}
	updateHandler services.UpdateHandler
	onError       func(err error)
}

func NewProxy(rcvq <-chan *parser.ParseResult, updateHandler services.UpdateHandler, onError func(err error)) *Proxy {
	return &Proxy{rcvq: rcvq, updateHandler: updateHandler, onError: onError, shutdown: make(chan struct{})}
}

// Run starts listening to rcvq and execute updateHandler
// To stop running caller should close rcvq channel
func (p *Proxy) Run() {
	for update := range p.rcvq {
		fmt.Printf("proxy rsv: %+v\n", update)

		// Parsing result occured
		if err := update.Err(); err != nil {
			p.handleError(err, update)
			continue
		}

		p.handleUpdate(update)
	}
}

// Report should be called when ErrURLUnavailable occurs.
// Mainly for debugging purposes
func (p *Proxy) Report(text *string) {
	// TODO: logger
	fmt.Printf("err: %s occured with: \n\t%s\n", parser.ErrURLUnavailable.Error(), *text)
}

func (p *Proxy) handleError(err error, update *parser.ParseResult) {
	p.onError(err)

	// Report web-page that caused an error
	if goerrors.Is(err, parser.ErrURLUnavailable) {
		p.Report(update.Raw())
	}
}

func (p *Proxy) handleUpdate(update *parser.ParseResult) {

	err := p.updateHandler(update)
	// TODO: handle errors somewhere else
	// TODO: proxy is just proxy :D
	var appErr *errors.ApplicationError
	if goerrors.As(err, &appErr) {
		// TODO: upgrade logger
		fmt.Println("trace: ", appErr.PrintStacktrace(), "error: ", appErr.Error())
		return
	}

	// updateHandler never returns ErrURLUnavailable.
	// Can pass nil as second argument
	p.handleError(err, nil)
}
