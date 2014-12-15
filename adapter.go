package rdispatch

import (
	"bytes"
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

type RemoteDispatcherAdapter interface {
	Method(r *http.Request) RemoteMethod
	ResolveRequest(r *http.Request) dispatch.Request
	WriteResponse(w http.ResponseWriter, r dispatch.Response)
}

type RemoteDestAdapter interface {
	BuildRequest(r dispatch.Request, addr, method string) *http.Request
	ResolveResponse(r *http.Response) dispatch.Response
}

type defaultDispatcherAdapter struct{}

func (d defaultDispatcherAdapter) Method(r *http.Request) RemoteMethod {
	return parseMethodFromHTTP(r)
}

func (d defaultDispatcherAdapter) ResolveRequest(r *http.Request) dispatch.Request {
	return ResolveRequest(r)
}

func (d defaultDispatcherAdapter) WriteResponse(w http.ResponseWriter, r dispatch.Response) {
	WriteResponse(w, r)
}

type defaultDestAdapter struct{}

func (d defaultDestAdapter) BuildRequest(r dispatch.Request, addr, method string) *http.Request {
	return BuildRequest(r, addr, method)
}

func (d defaultDestAdapter) ResolveResponse(r *http.Response) dispatch.Response {
	return ResolveResponse(r)
}

func parseMethodFromHTTP(r *http.Request) RemoteMethod {
	if r.Method == "PUT" {
		return MethodSend
	} else {
		return MethodCall
	}
}

func ResolveRequest(r *http.Request) dispatch.Request {
	return &RemoteRequest{
		SimpleRequest: dispatch.SimpleRequest{
			Proto: r.RequestURI,
			Sink:  parseSinkFromHTTP(r.Body, r.Header),
		},
		Auth:    parseAuthFromHTTP(r),
		TimeOut: parseTimeOutFromHTTP(r),
	}
}

func WriteResponse(w http.ResponseWriter, r dispatch.Response) {
	if sink := r.Body(); sink != nil {
		defer w.Write(sink.Bytes())
		w.Header().Set(ContentTypeKey, contentTypeToHTTP(sink.ContentType))
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
	var rsp dispatch.SimpleResponse

	if r.StatusCode != http.StatusOK {
		rsp.Err = statusError{
			statusCode: r.StatusCode,
		}
	}

	rsp.Sink = parseSinkFromHTTP(r.Body, r.Header)

	return &rsp
}

func BuildRequest(r dispatch.Request, addr, method string) *http.Request {
	sink := r.Body()
	if sink == nil {
		req, _ := http.NewRequest(method, addr, nil)
		return req
	}

	req, err := http.NewRequest(method, addr, bytes.NewBuffer(sink.Bytes()))
	if err != nil {
		return nil
	}

	req.Header.Set(ContentTypeKey, contentTypeToHTTP(sink.ContentType))
	if r, ok := r.(*RemoteRequest); ok {
		if r.Auth != nil {
			req.SetBasicAuth(r.Auth.UserName, r.Auth.Password)
		}
		if r.TimeOut > 0 {
			req.Header.Set(TimeOutKey, r.TimeOut.String())
		}
	}

	return req
}

func parseSinkFromHTTP(body io.ReadCloser, header http.Header) *dispatch.Sink {
	b, err := ioutil.ReadAll(body)
	if err != nil {
		return nil
	}
	c := contentTypeFromHTTP(header.Get(ContentTypeKey))
	s := dispatch.NewBytesSink(b)
	s.ContentType = c
	return s
}

func parseAuthFromHTTP(r *http.Request) *Auth {
	if username, password, ok := r.BasicAuth(); ok {
		return &Auth{
			UserName: username,
			Password: password,
		}
	}
	return nil
}

func parseTimeOutFromHTTP(r *http.Request) time.Duration {
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

func contentTypeFromHTTP(v string) dispatch.ContentType {
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

func contentTypeToHTTP(c dispatch.ContentType) string {
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
