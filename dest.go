package rdispatch

import (
	"net/http"

	"github.com/huangml/dispatch"
)

type RemoteDestAdapter interface {
	BuildRequest(r dispatch.Request, remoteAddr string, method string) (*http.Request, error)
	ResolveResponse(r *http.Response) dispatch.Response
	HTTPMethod(m RemoteMethod) string
}

type RemoteDest struct {
	remoteAddr string
	adapter    RemoteDestAdapter
}

func NewRemoteDest(remoteAddr string, adapter RemoteDestAdapter) *RemoteDest {
	if adapter == nil {
		adapter = defaultDestAdapter{}
	}

	return &RemoteDest{
		remoteAddr: remoteAddr,
		adapter:    adapter,
	}
}

func (d *RemoteDest) Call(ctx *dispatch.Context, r dispatch.Request) dispatch.Response {
	return d.doRemoteRequest(r, MethodCall)
}

func (d *RemoteDest) Send(r dispatch.Request) dispatch.Response {
	return d.doRemoteRequest(r, MethodSend)
}

func (d *RemoteDest) doRemoteRequest(r dispatch.Request, method RemoteMethod) dispatch.Response {
	rr, err := d.adapter.BuildRequest(r, d.remoteAddr, d.adapter.HTTPMethod(method))
	if rr == nil || err != nil {
		return dispatch.NewSimpleResponse(nil, statusError{http.StatusInternalServerError, err.Error()})
	}

	var client http.Client
	rs, err := client.Do(rr)
	if err != nil {
		return dispatch.NewSimpleResponse(nil, statusError{http.StatusInternalServerError, err.Error()})
	}

	return d.adapter.ResolveResponse(rs)
}

type defaultDestAdapter struct{}

func (d defaultDestAdapter) BuildRequest(r dispatch.Request, remoteAddr string, method string) (*http.Request, error) {
	return BuildRequest(r, remoteAddr, method)
}

func (d defaultDestAdapter) ResolveResponse(r *http.Response) dispatch.Response {
	return ResolveResponse(r)
}

func (d defaultDestAdapter) HTTPMethod(m RemoteMethod) string {
	return HTTPMethod(m)
}
