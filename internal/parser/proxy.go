package parser

import (
	goerrors "errors"
	"fmt"
	"parser/internal/errors"
)

type UpdateHandler func(result *ParseResult) error

// Proxy handles output from `rcvq` and handles it via `updateHandler`
// On error the callback `onError` is executed
type Proxy struct {
	rcvq          <-chan *ParseResult
	shutdown      chan struct{}
	updateHandler UpdateHandler
	onError       func(err error)
}

func NewProxy(rcvq <-chan *ParseResult, updateHandler UpdateHandler, onError func(err error)) *Proxy {
	return &Proxy{rcvq: rcvq, updateHandler: updateHandler, onError: onError, shutdown: make(chan struct{})}
}

// Run starts listening to rcvq and execute updateHandler
// To stop running caller should close rcvq channel
func (p *Proxy) Run() {
	for update := range p.rcvq {
		fmt.Printf("proxy rsv: %+v\n", update)
		// Parsing result occured
		if err := update.Err(); err != nil {
			// Report parsed text on which error has occured
			if goerrors.Is(err, ErrURLUnavailable) {
				p.Report(update.Raw())
			}
			p.onError(err)
			continue
		}

		err := p.updateHandler(update)
		var appErr *errors.ApplicationError
		if goerrors.As(err, &appErr) {
			fmt.Println("trace: ", appErr.PrintStacktrace(), "error: ", appErr.Error())
			continue
		}

		if err != nil {
			p.onError(err)
		}
	}
}

// Report should be called when ErrURLUnavailable occurs.
// Mainly for debugging purposes
func (p *Proxy) Report(text *string) {
	// TODO: logger
	fmt.Printf("err: %s occured with: \n\t%s\n", ErrURLUnavailable.Error(), *text)
}
