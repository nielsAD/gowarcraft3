// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package capi

import "fmt"

// Errors
//
// Area Code Reason
// 8    1    Not Connected to chat
// 8    2    Bad request
// 6    5    Request timed out
// 6    8    Hit rate limit
var (
	Success           = Status{0, 0}
	ErrNotConnected   = Status{8, 1}
	ErrBadRequest     = Status{8, 2}
	ErrRequestTimeout = Status{6, 5}
	ErrRateLimit      = Status{6, 8}
)

// Status object
type Status struct {
	Area int `json:"area"`
	Code int `json:"code"`
}

// Error converts the Status object to an appropriate error message
func (s *Status) Error() string {
	if s == nil || (*s == Success) {
		return ""
	}

	switch *s {
	case ErrNotConnected:
		return "capi: Not connected"
	case ErrBadRequest:
		return "capi: Bad request"
	case ErrRequestTimeout:
		return "capi: Request timeout"
	case ErrRateLimit:
		return "capi: Rate limit"
	default:
		return fmt.Sprintf("capi: Unknown error (%+v)", *s)
	}
}

// Timeout error (able to retry later)
func (s *Status) Timeout() bool {
	return s.Area == 6
}
