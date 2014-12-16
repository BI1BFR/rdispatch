package rdispatch

import (
	"fmt"
	"net/http"
	"time"

	"github.com/huangml/dispatch"
)

type Auth struct {
	UserName string
	Password string
}

type RemoteRequest struct {
	dispatch.SimpleRequest
	*Auth
	TimeOut time.Duration
}

type StatusError interface {
	StatusCode() int
	Text() string
	Error() string
}

type statusError struct {
	statusCode int
	text       string
}

func (e statusError) StatusCode() int {
	return e.statusCode
}

func (e statusError) Text() string {
	return e.text
}

func (e statusError) Error() string {
	return fmt.Sprintf("%s[%s]", http.StatusText(e.statusCode), e.text)
}

func ToStatusError(err error) error {
	if err == nil {
		return nil
	}

	if e, ok := err.(StatusError); ok {
		return e
	}

	switch err.(type) {
	case dispatch.ProtocolNotImplementError:
		return statusError{http.StatusNotImplemented, err.Error()}

	case dispatch.DestNotFoundError:
		return statusError{http.StatusNotFound, err.Error()}

	case dispatch.ContextCanceledError:
		return statusError{http.StatusRequestTimeout, err.Error()}

	case dispatch.PanicError:
		return statusError{http.StatusInternalServerError, err.Error()}

	default:
		return statusError{http.StatusInternalServerError, err.Error()}
	}
}
