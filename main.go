package main

import (
	"errors"
	"fmt"
	"strings"
)

type State string

const (
	Degraded    State = "degraded"
	Unavailable State = "unavaiable"
)

func main() {
	errorList()
	wrappedError()
}

func errorList() {
	esReturnNil()
	esReturnSingleError()
	esReturnMultipleErrors()
	serrs := esCombineDifferentErrors()
	result := esProcessErrors(serrs)
	fmt.Println(result)

}

func wrappedError() {
	wrpReturnNil()
	wrpReturnSingleError()
	wrpReturnMultipleErrors()
	serrs := wrpCombineDifferentErrors()
	result := wrpProcessErrors(serrs)
	fmt.Println(result)

}

func esReturnSingleError() *StateError {
	// CONS
	//  * returning single error requires builder

	// NOTE also possible is more verbose
	// b := StateErrorBuilder{}
	// return b.AddDegraded("foobar").Errors()

	return &StateError{State: Degraded, Msg: "first single error"}
}

func wrpReturnSingleError() *WrappedStateError {
	// CONS
	// note can't be `func wrpReturnSingleError() error` to avoid typecasting
	// to use methods like - Append
	return NewDegradedError("first single")
}

func esReturnNil() StateErrorList {
	return nil
}

func wrpReturnNil() *WrappedStateError {
	return nil
}

func esReturnMultipleErrors() StateErrorList {
	b := StateErrorBuilder{}
	// pros:
	//   * builder takes care or nil
	//   * O(1) operation
	// cons:
	// *
	return b.
		AddDegraded("multiple").
		AddDegraded("another error").
		AddUnavailable("for some reason").
		Errors()
}

func wrpReturnMultipleErrors() *WrappedStateError {
	first := NewDegradedError("multiple")

	// pros:
	//   *

	// cons:
	//   * O(N) operation
	//   * first can't be `nil`
	//

	// if first == nil {
	//  is needed before calling Append
	// }

	first.
		Append(NewDegradedError("another error")).
		Append(NewUnavailable("for some reason"))
	return first
}

func esCombineDifferentErrors() StateErrorList {
	// combine 2  or more  error lists into one
	nilErr := esReturnNil()
	first := esReturnSingleError()
	multiple := esReturnMultipleErrors()
	third := esReturnMultipleErrors()

	b := StateErrorBuilder{}
	return b.Append(nilErr).
		Add(first).
		Append(multiple).
		Append(nilErr).
		Append(third).
		Errors()
}

func wrpCombineDifferentErrors() *WrappedStateError {
	// combine 2  or more  error lists into one
	nilErr := wrpReturnNil()
	first := wrpReturnSingleError()
	multiple := wrpReturnMultipleErrors()
	third := wrpReturnMultipleErrors()

	// cons
	//  * can't use `nilErr` for Append
	// pros
	//  * no need for builder but need to handle `nil` (which can be cumbersome)

	// 🤯 can you find what's wrong with this?
	// return first.Append(nilErr).
	//   Append(multiple).
	//   Append(third)

	first.Append(nilErr).
		Append(multiple).
		Append(third)
	return first
}

func esProcessErrors(serrs StateErrorList) string {
	sb := strings.Builder{}

	// process all errors
	for _, serr := range serrs {
		sb.WriteString(serr.Error())
		sb.WriteString("\n")
	}
	return sb.String()

}

func wrpProcessErrors(errs *WrappedStateError) string {
	sb := strings.Builder{}

	unwrapAll(errs, func(err *WrappedStateError) bool {
		sb.WriteString(err.Error())
		sb.WriteString("\n")
		return true
	})
	// process all errors
	return sb.String()

}

// StateError is a simple error, represents a State and a message for user
type StateError struct {
	State State
	Msg   string
}

func (e *StateError) Error() string {
	return fmt.Sprintf("StateError: %s: %s", e.State, e.Msg)
}

// StateErrorList is only a list of StateErrors and isn't an error but can be
// by implementing `Error() string`
type StateErrorList []*StateError

// StateErrorBuilder makes is convenient to work with list of Errors, handle nil,
// combine another StateErrorList etc
type StateErrorBuilder struct {
	errors StateErrorList
}

func (b *StateErrorBuilder) addError(s State, msg string) *StateErrorBuilder {
	b.errors = append(b.errors, &StateError{State: s, Msg: msg})
	return b
}

func (b *StateErrorBuilder) Add(serrs ...*StateError) *StateErrorBuilder {
	return b.Append(serrs)
}

func (b *StateErrorBuilder) AddDegraded(msg string) *StateErrorBuilder {
	return b.addError(Degraded, msg)

}
func (b *StateErrorBuilder) AddUnavailable(msg string) *StateErrorBuilder {
	return b.addError(Unavailable, msg)
}

func (b *StateErrorBuilder) AddIfNotNil(err error, s State) *StateErrorBuilder {
	if err == nil {
		return b
	}

	return b.addError(s, err.Error())
}

func (b *StateErrorBuilder) Append(serrList ...StateErrorList) *StateErrorBuilder {
	if len(serrList) == 0 {
		return b
	}

	for _, serrs := range serrList {
		if len(serrs) == 0 {
			continue
		}
		// NOTE: nils could sneak in
		b.errors = append(b.errors, serrs...)
	}

	return b
}

func (b *StateErrorBuilder) Errors() StateErrorList {
	return b.errors
}

type WrappedStateError struct {
	State State
	Msg   string

	// NOTE: this could just be *WrappedStateError instead of error but
	// if we want to use `errors.Unwarp`, we need to use interface - error
	wrapped error
}

var _ error = (*WrappedStateError)(nil)

func (se *WrappedStateError) Error() string {
	return fmt.Sprintf("WrappedStateError: %s: %s", se.State, se.Msg)
}

func (se *WrappedStateError) Unwrap() error {
	// fmt.Printf("\t ~~> unwrap %q: -> %v\n", se.Msg, se.wrapped)
	return se.wrapped
}

func (se *WrappedStateError) Append(err *WrappedStateError) *WrappedStateError {
	// fmt.Printf(" append: %q -> %v  \n", se.Msg, err)
	if err == nil {
		// fmt.Printf("nil err so not wrapping")
		return se
	}

	if se.wrapped != nil {
		// fmt.Printf(" ~~> dig: %q -> %q  \n", se.Msg, err.Msg)
		wrp := se.wrapped.(*WrappedStateError)
		return wrp.Append(err)
	}

	// fmt.Printf("wrap: %q -> %q  \n", se.Msg, err.Msg)

	se.wrapped = err
	return se.wrapped.(*WrappedStateError)
}

func NewDegradedError(msg string) *WrappedStateError {
	return &WrappedStateError{State: Degraded, Msg: msg, wrapped: nil}
}
func NewUnavailable(msg string) *WrappedStateError {
	return &WrappedStateError{State: Unavailable, Msg: msg, wrapped: nil}
}

func unwrapAll(head *WrappedStateError, fn func(*WrappedStateError) bool) {
	// fmt.Println("== unwarp ==")

	var current error = head
	for {

		if current == nil {
			// fmt.Println("... breaking: current is nil")
			break
		}

		var se *WrappedStateError
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
