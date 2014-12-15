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

func (e statusError) Error() string {
	return fmt.Sprintf("%s[%s]", http.StatusText(e.statusCode), e.text)
}

func ErrResponse(err error) dispatch.Response {
	if err == nil {
		return &dispatch.SimpleResponse{}
	}

	if e, ok := err.(StatusError); ok {
		return &dispatch.SimpleResponse{
			Err: e,
		}
	}

	switch err.(type) {
	case dispatch.ProtocolNotImplementError:
		return &dispatch.SimpleResponse{
			Err: statusError{http.StatusNotImplemented, err.Error()},
		}
	case dispatch.DestNotFoundError:
		return &dispatch.SimpleResponse{
			Err: statusError{http.StatusNotFound, err.Error()},
		}
	case dispatch.ContextCanceledError:
		return &dispatch.SimpleResponse{
			Err: statusError{http.StatusRequestTimeout, err.Error()},
		}
	case dispatch.PanicError:
		return &dispatch.SimpleResponse{
			Err: statusError{http.StatusInternalServerError, err.Error()},
		}
	default:
		return &dispatch.SimpleResponse{
			Err: statusError{http.StatusInternalServerError, err.Error()},
		}
	}
}
