// Copyright 2015 go-swagger maintainers
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package errors

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-openapi/errors"
)

const (
	customResponseCode = 600
)

// DefaultHandler provides a default implementation.
var DefaultHandler = &ErrorHandler{ContentType: "application/vnd.cia.v1+json"}

// ErrorHandler defines a simple struct to hold the Content-Type value to use when handling errors.
type ErrorHandler struct {
	ContentType string
}

// ServeError the error handler interface implemenation.
func (h *ErrorHandler) ServeError(rw http.ResponseWriter, r *http.Request, err error) {
	rw.Header().Set("Content-Type", h.ContentType)

	switch e := err.(type) {
	case *errors.MethodNotAllowedError:
		rw.Header().Add("Allow", strings.Join(err.(*errors.MethodNotAllowedError).Allowed, ","))
		rw.WriteHeader(asHTTPCode(int(e.Code())))

		if r == nil || r.Method != http.MethodHead {
			_, _ = rw.Write(errorAsJSON(e))
		}
	case *errors.CompositeError:
		// CompositeError has a hard-coded code of 422 to maintain backwards-compatibility with
		// previous usages of go-openapi/errors; as such this workarounds that scenario to allow
		// overriding the code to be 400.
		rw.WriteHeader(http.StatusBadRequest)

		if r == nil || r.Method != http.MethodHead {
			b, _ := (&domainError{Code: fmt.Sprintf("%s-%03d", Default, http.StatusBadRequest), Message: e.Error()}).MarshalBinary()
			_, _ = rw.Write(b)
		}
	case errors.Error:
		rw.WriteHeader(asHTTPCode(int(e.Code())))

		if r == nil || r.Method != http.MethodHead {
			_, _ = rw.Write(errorAsJSON(e))
		}
	default:
		rw.WriteHeader(http.StatusInternalServerError)

		if r == nil || r.Method != http.MethodHead {
			_, _ = rw.Write(errorAsJSON(errors.New(http.StatusInternalServerError, err.Error())))
		}
	}
}

func errorAsJSON(err errors.Error) []byte {
	b, _ := (&domainError{Code: fmt.Sprintf("%s-%03d", Default, err.Code()), Message: err.Error()}).MarshalBinary()

	return b
}

func asHTTPCode(input int) int {
	if input >= customResponseCode {
		return http.StatusBadRequest
	}

	return input
}
