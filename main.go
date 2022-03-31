package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/sthaha/errors/list"
	"github.com/sthaha/errors/wrapped"
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
	fmt.Println("single wrapped: ", wrpReturnSingleError())
	fmt.Println("multi wrapped: ", wrpReturnMultipleErrors())
	serrs := wrpCombineDifferentErrors()
	result := wrpProcessErrors(serrs)
	fmt.Println(result)

}

func esReturnSingleError() *list.StateError {
	_ = &list.StateError{State: list.Degraded, Msg: "foobar"}
	return list.NewDegradedError("first single error")
}

func esReturnNil() list.StateErrors {
	return nil
}

func esReturnMultipleErrors() list.StateErrors {
	b := list.StateErrorBuilder{}
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

func esCombineDifferentErrors() list.StateErrors {
	// combine 2  or more  error lists into one

	// I don't understand why we would want to append a nil error ever? a
	// nil error means no error, so there is nothing to append?
	nilErr := esReturnNil()
	first := esReturnSingleError()
	multiple := esReturnMultipleErrors()
	third := esReturnMultipleErrors()

	b := list.StateErrorBuilder{}
	return b.Append(nilErr).
		Add(first).
		Append(multiple).
		Append(nilErr).
		Append(third).
		Errors()
}

func esProcessErrors(serrs list.StateErrors) string {
	sb := strings.Builder{}

	// process all errors
	for _, serr := range serrs {
		sb.WriteString(serr.Error())
		sb.WriteString("\n")
	}
	return sb.String()

}

// using a struct literal for error construction is just for brevity, this could
// also be handled by constructors
// the big advantage in my optinion is that except the functions that create,
// combine or report this custom error can just treat it as errors, see the
// function signatures

func wrpReturnSingleError() error {
	return wrapped.StateError{State: wrapped.Unavailable, Msg: "some reason"}
}

func wrpReturnMultipleErrors() error {
	e1 := wrapped.StateError{State: wrapped.Degraded, Msg: "multiple"}
	e2 := wrapped.StateError{State: wrapped.Degraded, Msg: "another error", Wrapped: &e1}
	e3 := wrapped.StateError{State: wrapped.Unavailable, Msg: "some other reason", Wrapped: &e2}
	return e3
}

func wrpCombineDifferentErrors() error {
	// combine 2  or more  error lists into one
	first := wrpReturnSingleError()
	second := wrpReturnMultipleErrors()
	third := wrpReturnMultipleErrors()
	first = wrapped.JoinErrLists(first, wrapped.JoinErrLists(second, third))

	return first
}

func wrpProcessErrors(errs error) string {
	var e wrapped.StateError
	if errors.As(errs, &e) {
		return e.Report()
	}
	return errs.Error()

}
