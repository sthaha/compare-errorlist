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

type statusReporter struct {
}

type Status interface {
	State() string
	Message() string
}

type StatusReport interface {
	Degraded() Status
	Unavailable() Status
}

func (sr *statusReporter) ReportStatus(s StatusReport) {
	degraded := s.Degraded()
	unavailable := s.Unavailable()

	fmt.Println("== Status Report ==")
	if degraded != nil {
		fmt.Printf("\t degraded: %s\n", degraded.Message())
	}

	if unavailable != nil {
		fmt.Printf("\t unavailable: %s\n", unavailable.Message())
	}
	fmt.Println("== Status Report ==")
}

func errorList() {
	fmt.Printf("%v \n", esReturnNil())

	single := esReturnSingleError()
	fmt.Printf("%v \n", single)

	multiple := esReturnMultipleErrors()
	fmt.Printf("%v \n", multiple)

	serrs := esCombineDifferentErrors()
	result := esProcessErrors(serrs)

	fmt.Println(result)

	report := statusReportFromListStateErrors(serrs)
	sr := statusReporter{}
	sr.ReportStatus(report)
}

func statusReportFromWrappedStateErrors(errs *wrapped.StateError) StatusReport {
	report := &statusReport{}

	degradedReasons := []string{}
	unavailableReasons := []string{}

	wrapped.ForEach(errs, func(err wrapped.StateError) bool {
		switch err.State {
		case wrapped.Degraded:
			degradedReasons = append(degradedReasons, err.Msg)

		case wrapped.Unavailable:
			unavailableReasons = append(unavailableReasons, err.Msg)
		}
		return true
	})

	if len(degradedReasons) > 0 {
		msg := strings.Join(degradedReasons, ", ")
		report.degraded = &wrappedStatus{wrapped.NewDegradedError(msg)}
	}

	if len(unavailableReasons) > 0 {
		msg := strings.Join(unavailableReasons, ", ")
		report.unavailable = &wrappedStatus{wrapped.NewUnavailableError(msg)}
	}

	return report
}

func statusReportFromListStateErrors(serrs list.StateErrors) StatusReport {
	report := &statusReport{}

	degradedReasons := []string{}
	unavailableReasons := []string{}

	for _, err := range serrs {
		switch err.State {
		case list.Degraded:
			degradedReasons = append(degradedReasons, err.Msg)

		case list.Unavailable:
			unavailableReasons = append(unavailableReasons, err.Msg)
		}
	}

	if len(degradedReasons) > 0 {
		msg := strings.Join(degradedReasons, ", ")
		report.degraded = &listStatus{list.NewDegradedError(msg)}
	}

	if len(unavailableReasons) > 0 {
		msg := strings.Join(unavailableReasons, ", ")
		report.unavailable = &listStatus{list.NewUnavailableError(msg)}
	}

	return report
}

// adapter for converting status error to Status
type listStatus struct {
	err *list.StateError
}

func (s *listStatus) State() string {
	return string(s.err.State)
}

func (s *listStatus) Message() string {
	return s.err.Msg
}

// adapter for converting wrapped.StateError to Status
type wrappedStatus struct {
	err *wrapped.StateError
}

func (s *wrappedStatus) State() string {
	return string(s.err.State)
}

func (s *wrappedStatus) Message() string {
	return s.err.Msg
}

type statusReport struct {
	degraded    Status
	unavailable Status
}

var _ StatusReport = (*statusReport)(nil)

func (s *statusReport) Unavailable() Status {
	return s.unavailable
}

func (s *statusReport) Degraded() Status {
	return s.degraded
}

func wrappedError() {
	fmt.Printf("%v \n", wrpReturnNil())

	single := wrpReturnSingleError()
	fmt.Printf("%v \n", single)

	multiple := wrpReturnMultipleErrors()
	fmt.Printf("%v \n", multiple)

	serrs := wrpCombineDifferentErrors()
	result := wrpProcessErrors(serrs)

	fmt.Println(result)

	report := statusReportFromWrappedStateErrors(serrs)
	sr := statusReporter{}
	sr.ReportStatus(report)

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
		Append(wrapped.NewUnavailableError("for some reason"))
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

	wrapped.ForEach(errs, func(err wrapped.StateError) bool {
		sb.WriteString(err.Error())
		sb.WriteString("\n")
		return true
	})
	// process all errors
	return sb.String()

}
