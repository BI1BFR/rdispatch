package rdispatch

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/huangml/dispatch"
)

type RemoteMethod int

const (
	MethodCall RemoteMethod = iota
	MethodSend
)

func (m RemoteMethod) String() string {
	switch m {
	case MethodCall:
		return "MethodCall"
	case MethodSend:
		return "MethodSend"
	default:
		return fmt.Sprintf("RemoteMethod(%d)", m)
	}
}

func HTTPMethod(m RemoteMethod) string {
	if m == MethodSend {
		return "POST"
	} else {
		return "PUT"
	}
}

func ParseMethodFromHTTP(r *http.Request) RemoteMethod {
	if r.Method == "PUT" {
		return MethodSend
	} else {
		return MethodCall
	}
}

func ResolveRequest(r *http.Request) dispatch.Request {
	return &RemoteRequest{
		SimpleRequest: dispatch.NewSimpleRequest(r.RequestURI, r.RequestURI, ParseSinkFromHTTP(r.Body, r.Header)),
		Auth:          ParseAuthFromHTTP(r),
		TimeOut:       ParseTimeOutFromHTTP(r),
	}
}

func WriteResponse(w http.ResponseWriter, r dispatch.Response) {
	if sink := r.Body(); sink != nil {
		defer w.Write(sink.Bytes())
		w.Header().Set(ContentTypeKey, ContentTypeToHTTP(sink.ContentType))
	}

	if r.Error() != nil {
		if e, ok := r.Error().(StatusError); ok {
			w.WriteHeader(e.StatusCode())
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

func ResolveResponse(r *http.Response) dispatch.Response {
	if r.StatusCode != http.StatusOK && r.StatusCode != http.StatusAccepted {
		return dispatch.NewSimpleResponse(nil, statusError{statusCode: r.StatusCode})
	} else {
		return dispatch.NewSimpleResponse(ParseSinkFromHTTP(r.Body, r.Header), nil)
	}
}

func BuildRequest(r dispatch.Request, remoteAddr string, method string) (*http.Request, error) {
	sink := r.Body()

	var buffer *bytes.Buffer
	if sink != nil {
		buffer = bytes.NewBuffer(sink.Bytes())
	}

	req, err := http.NewRequest(method, remoteAddr, buffer)
	if err != nil || req == nil {
		return nil, err
	}

	if sink != nil {
		req.Header.Set(ContentTypeKey, ContentTypeToHTTP(sink.ContentType))
	}

	if r, ok := r.(*RemoteRequest); ok {
		if r.Auth != nil {
			req.SetBasicAuth(r.Auth.UserName, r.Auth.Password)
		}
		if r.TimeOut > 0 {
			req.Header.Set(TimeOutKey, r.TimeOut.String())
		}
	}

	return req, err
}

func ParseSinkFromHTTP(body io.ReadCloser, header http.Header) *dispatch.Sink {
	b, err := ioutil.ReadAll(body)
	if err != nil {
		return nil
	}
	c := ContentTypeFromHTTP(header.Get(ContentTypeKey))
	s := dispatch.NewBytesSink(b)
	s.ContentType = c
	return s
}

func ParseAuthFromHTTP(r *http.Request) *Auth {
	if username, password, ok := r.BasicAuth(); ok {
		return &Auth{
			UserName: username,
			Password: password,
		}
	}
	return nil
}

func ParseTimeOutFromHTTP(r *http.Request) time.Duration {
	if t, err := time.ParseDuration(r.Header.Get(TimeOutKey)); err == nil {
		return t
	}

	return 0
}

const (
	OctetStream = "application/octet-stream"
	XProtoBuf   = "application/x-protobuf"
	TextPlain   = "text/plain"

	TimeOutKey     = "X-Dispatch-Timeout"
	ContentTypeKey = "Content-Type"
)

func ContentTypeFromHTTP(v string) dispatch.ContentType {
	switch v {
	case OctetStream:
		return dispatch.Bytes
	case TextPlain:
		return dispatch.Text
	case XProtoBuf:
		return dispatch.ProtoBuf
	default:
		return dispatch.Bytes
	}
}

func ContentTypeToHTTP(c dispatch.ContentType) string {
	switch c {
	case dispatch.Bytes:
		return OctetStream
	case dispatch.Text:
		return TextPlain
	case dispatch.ProtoBuf:
		return XProtoBuf
	default:
		return OctetStream
	}
}
