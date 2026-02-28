package xhttp

import (
	"net/http"

	"github.com/raymondwongso/goerp/domain/xerror"
)

// ErrorMapper defines function that map error into proper http status
type ErrorMapper func(err error) int

// MapError map error to certain http status
// if err is xerror, it will return proper http status according to error code
// do note that this is the standard error mapping
// when you want to implement your own mapper, create a function that satisfy the ErrorMapper interface
// The typical and suggested flow is:
// 1. Inject ErrorMapper to your usecase
// 2. Call your error mapper
// 3. if return value is 0 (means that there are unknown error mapping), you can use this MapError default mapping
func MapError(err error) int {
	switch xerror.GetCode(err) {
	case xerror.CodeConflict:
		return http.StatusConflict
	case xerror.CodeDuplicate:
		return http.StatusConflict
	case xerror.CodeForbidden:
		return http.StatusForbidden
	case xerror.CodeInternal:
		return http.StatusInternalServerError
	case xerror.CodeInvalidParameter:
		return http.StatusBadRequest
	case xerror.CodeNotFound:
		return http.StatusNotFound
	case xerror.CodeNotImplemented:
		return http.StatusNotImplemented
	case xerror.CodeServiceUnavailable:
		return http.StatusServiceUnavailable
	case xerror.CodeTimeout:
		return http.StatusRequestTimeout
	case xerror.CodeUnauthorized:
		return http.StatusUnauthorized
	case xerror.CodeUnprocessable:
		return http.StatusUnprocessableEntity
	default:
		return 0
	}
}
