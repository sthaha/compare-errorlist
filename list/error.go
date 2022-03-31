package list

import (
	"fmt"
)

type State string

const (
	Degraded    State = "degraded"
	Unavailable State = "unavaiable"
)

// StateError is a simple error, represents a State and a message for user.
type StateError struct {
	State State
	Msg   string
}

func NewDegradedError(msg string) *StateError {
	return &StateError{State: Degraded, Msg: msg}
}

func NewUnavailable(msg string) *StateError {
	return &StateError{State: Unavailable, Msg: msg}
}

func (e *StateError) Error() string {
	return fmt.Sprintf("StateError: %s: %s", e.State, e.Msg)
}

// StateErrors is only a list of StateErrors and isn't an error but can be
// by implementing `Error() string`.
type StateErrors []*StateError

// StateErrorBuilder makes is convenient to work with list of Errors, handle nil,
// combine another StateErrors etc.
type StateErrorBuilder struct {
	errors StateErrors
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

func (b *StateErrorBuilder) Append(serrList ...StateErrors) *StateErrorBuilder {
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

func (b *StateErrorBuilder) Errors() StateErrors {
	return b.errors
}
