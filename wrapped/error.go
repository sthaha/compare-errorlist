package wrapped

import (
	"errors"
	"fmt"
)

type State string

const (
	Degraded    State = "degraded"
	Unavailable State = "unavaiable"
)

type StateError struct {
	State State
	Msg   string

	// NOTE: this could just be *StateError instead of error but
	// if we want to use `errors.Unwarp`, we need to use interface - error
	wrapped error
}

var _ error = (*StateError)(nil)

func NewDegradedError(msg string) *StateError {
	return &StateError{State: Degraded, Msg: msg, wrapped: nil}
}

func NewUnavailable(msg string) *StateError {
	return &StateError{State: Unavailable, Msg: msg, wrapped: nil}
}

func (se *StateError) Error() string {
	return fmt.Sprintf("StateError: %s: %s", se.State, se.Msg)
}

func (se *StateError) Unwrap() error {
	// fmt.Printf("\t ~~> unwrap %q: -> %v\n", se.Msg, se.wrapped)
	return se.wrapped
}

func (se *StateError) Append(err *StateError) *StateError {
	// fmt.Printf(" append: %q -> %v  \n", se.Msg, err)
	if err == nil {
		// fmt.Printf("nil err so not wrapping")
		return se
	}

	if se.wrapped != nil {
		// fmt.Printf(" ~~> dig: %q -> %q  \n", se.Msg, err.Msg)
		wrp := se.wrapped.(*StateError)
		return wrp.Append(err)
	}

	// fmt.Printf("wrap: %q -> %q  \n", se.Msg, err.Msg)

	se.wrapped = err
	return se.wrapped.(*StateError)
}

func ForEach(head *StateError, fn func(*StateError) bool) {
	// fmt.Println("== unwarp ==")

	var current error = head
	for {

		if current == nil {
			// fmt.Println("... breaking: current is nil")
			break
		}

		var se *StateError
		if !errors.As(current, &se) {
			// ⚠️   NOTE: this shouldn't happen but if does
			panic(" something went wrong here ...")
		}

		if next := fn(se); !next {
			// fmt.Println("... breaking processing on request")
			break
		}
		// NOTE: we know this is a wrapped state error
		// so instead of errors.Wrap() .. we could just use current.wrapped 🤨
		// and not have to use `errors.Unwrap`
		current = errors.Unwrap(current)
	}
}
