package wrapped

import (
	"errors"
	"fmt"
)

type State string

const (
	Degraded    State = "degraded"
	Unavailable State = "unavailable"
)

type StateError struct {
	State   State
	Msg     string
	Wrapped *StateError
}

func (se StateError) Error() string {
	return fmt.Sprintf("WrappedStateError: %s: %s", se.State, se.Msg)
}

func (se StateError) Unwrap() error {
	// fmt.Printf("\t ~~> unwrap %q: -> %v\n", se.Msg, se.wrapped)
	return se.Wrapped
}

func (se *StateError) Append(err *StateError) *StateError {
	if se.Wrapped == nil {
		fmt.Println("appending ", err)
		se.Wrapped = err
		return se
	}
	return se.Wrapped.Append(err)
}

func (se StateError) Report() string {
	if se.Wrapped == nil {
		return se.Error()
	} else {
		return fmt.Sprintf("%v->%v", se.Error(), se.Wrapped.Report())
	}
}

func JoinErrLists(l1 error, l2 error) error {
	var e, v StateError
	if errors.As(l1, &e) {
		if errors.As(l2, &v) {
			e.Append(&v)
			return e
		}
		// this should probably panic if l1 is not a StateError
	}
	return l1
}
