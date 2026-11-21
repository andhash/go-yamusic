package yamusic

import (
	"errors"
	"fmt"
	"strings"
)

// Error is a refused API call. Status is the HTTP status. A 200 with an
// error object in the body keeps Status 200 and still returns *Error.
// Name is the API error code (for example "not-found") when the body has one.
type Error struct {
	Status  int
	Name    string
	Message string
}

func (e *Error) Error() string {
	if e == nil {
		return "yamusic: <nil>"
	}
	msg := strings.TrimSpace(e.Message)
	if msg == "" {
		msg = e.Name
	}
	if msg == "" {
		msg = "request failed"
	}
	if e.Status != 0 && e.Name != "" {
		return fmt.Sprintf("yamusic: %s (%d %s)", msg, e.Status, e.Name)
	}
	if e.Status != 0 {
		return fmt.Sprintf("yamusic: %s (%d)", msg, e.Status)
	}
	if e.Name != "" {
		return fmt.Sprintf("yamusic: %s (%s)", msg, e.Name)
	}
	return "yamusic: " + msg
}

// IsUnauthorized reports whether err is a 401 or 403 from the API.
func IsUnauthorized(err error) bool {
	var api *Error
	return errors.As(err, &api) && (api.Status == 401 || api.Status == 403)
}

// IsNotFound reports whether err is a 404, or a 200 whose error name is not-found.
func IsNotFound(err error) bool {
	var api *Error
	if !errors.As(err, &api) {
		return false
	}
	return api.Status == 404 || api.Name == "not-found"
}

// IsBadRequest reports whether err is a 400.
func IsBadRequest(err error) bool {
	var api *Error
	return errors.As(err, &api) && api.Status == 400
}
