package main

import (
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
	wrpReturnNil()
	wrpReturnSingleError()
	wrpReturnMultipleErrors()
	serrs := wrpCombineDifferentErrors()
	result := wrpProcessErrors(serrs)
	fmt.Println(result)

}

func esReturnSingleError() *list.StateError {
	_ = &list.StateError{State: list.Degraded, Msg: "foobar"}
	return list.NewDegradedError("first single error")
}

func wrpReturnSingleError() *wrapped.StateError {
	// NOTE: can't be `func wrpReturnSingleError() error` to avoid typecasting
	// and to use methods like - Append
	_ = &wrapped.StateError{State: wrapped.Degraded, Msg: "foobar"}
	return wrapped.NewDegradedError("first single")
}

func esReturnNil() list.StateErrors {
	return nil
}

func wrpReturnNil() *wrapped.StateError {
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

func wrpReturnMultipleErrors() *wrapped.StateError {
	first := wrapped.NewDegradedError("multiple")

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
		Append(wrapped.NewDegradedError("another error")).
		Append(wrapped.NewUnavailable("for some reason"))
	return first
}

func esCombineDifferentErrors() list.StateErrors {
	// combine 2  or more  error lists into one
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

func wrpCombineDifferentErrors() *wrapped.StateError {
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

func esProcessErrors(serrs list.StateErrors) string {
	sb := strings.Builder{}

	// process all errors
	for _, serr := range serrs {
		sb.WriteString(serr.Error())
		sb.WriteString("\n")
	}
	return sb.String()

}

func wrpProcessErrors(errs *wrapped.StateError) string {
	sb := strings.Builder{}

	wrapped.ForEach(errs, func(err *wrapped.StateError) bool {
		sb.WriteString(err.Error())
		sb.WriteString("\n")
		return true
	})
	// process all errors
	return sb.String()

}
