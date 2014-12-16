package rdispatch

import (
	"net/http"

	"github.com/huangml/dispatch"
)

type RemoteDest struct {
	addr    string
	adapter RemoteDestAdapter
}

func NewRemoteDest(addr string, adapter RemoteDestAdapter) *RemoteDest {
	if adapter == nil {
		adapter = &defaultDestAdapter{}
	}

	return &RemoteDest{
		addr:    addr,
		adapter: adapter,
	}
}

func (d *RemoteDest) Call(ctx *dispatch.Context, r dispatch.Request) dispatch.Response {
	return d.doRemoteRequest(r, MethodCall)
}

func (d *RemoteDest) Send(r dispatch.Request) dispatch.Response {
	return d.doRemoteRequest(r, MethodSend)
}

func (d *RemoteDest) doRemoteRequest(r dispatch.Request, method RemoteMethod) dispatch.Response {
	rr := d.adapter.BuildRequest(r, d.addr, method)
	var client http.Client
	rs, err := client.Do(rr)
	if err != nil {
		return ErrResponse(err)
	}
	return d.adapter.ResolveResponse(rs)
}
