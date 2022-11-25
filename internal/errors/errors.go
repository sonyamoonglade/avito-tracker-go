package errors

// ErrKind is enum
// To prevent from ErrKind("some random error")...
type ErrKind struct {
	K string
}

var (
	InternalKind = ErrKind{"internal"}

	// Not sure about this yet...
	DomainKind = ErrKind{"domain"}
)

type ApplicationError struct {
	stacktrace []string
	err        error
	kind       ErrKind
}

func (ae *ApplicationError) Error() string {
	return ae.err.Error()
}

func (ae *ApplicationError) PrintStacktrace() string {
	var buff string
	for _, trace := range ae.stacktrace {
		buff += trace
	}

	return buff
}

func (ae *ApplicationError) addTrace(trace string) {
	ae.stacktrace = append([]string{trace, "."}, ae.stacktrace...)
}

func WrapInternal(err error, context string) error {
	return &ApplicationError{
		kind:       InternalKind,
		err:        err,
		stacktrace: []string{context},
	}
}

func WrapDomain(err error) error {
	return &ApplicationError{
		kind:       DomainKind,
		err:        err,
		stacktrace: nil,
	}
}

func ChainInternal(err error, context string) error {
	ae, ok := err.(*ApplicationError)
	if !ok {
		return err
	}

	ae.addTrace(context)

	return ae
}
