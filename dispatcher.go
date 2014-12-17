package rdispatch

import (
	"net/http"
	"time"

	"github.com/huangml/dispatch"
)

type RemoteDispatcherAdapter interface {
	Method(r *http.Request) RemoteMethod
	ResolveRequest(r *http.Request) dispatch.Request
	WriteResponse(w http.ResponseWriter, r dispatch.Response)
}

type RemoteDispatcher struct {
	*dispatch.Dispatcher
	adapter RemoteDispatcherAdapter
}

func NewRemoteDispatcher(d *dispatch.Dispatcher, adapter RemoteDispatcherAdapter) *RemoteDispatcher {
	if adapter == nil {
		adapter = defaultDispatcherAdapter{}
	}

	return &RemoteDispatcher{
		Dispatcher: d,
		adapter:    adapter,
	}
}

func (d *RemoteDispatcher) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rr := d.adapter.ResolveRequest(r)
	if rr == nil {
		d.adapter.WriteResponse(w, dispatch.NewSimpleResponse(nil, statusError{http.StatusBadRequest, ""}))
		return
	}

	var rsp dispatch.Response

	switch d.adapter.Method(r) {
	case MethodCall:
		t := time.Second * 10
		if req, ok := rr.(*RemoteRequest); ok {
			t = req.TimeOut
		}
		rsp = d.Call(dispatch.NewContextWithTimeOut(t), rr)
	case MethodSend:
		err := d.Send(rr)
		rsp = dispatch.NewSimpleResponse(nil, ToStatusError(err))
	default:
		rsp = dispatch.NewSimpleResponse(nil, statusError{http.StatusBadRequest, ""})
	}

	d.adapter.WriteResponse(w, rsp)
}

type defaultDispatcherAdapter struct{}

func (d defaultDispatcherAdapter) Method(r *http.Request) RemoteMethod {
	return ParseMethodFromHTTP(r)
}

func (d defaultDispatcherAdapter) ResolveRequest(r *http.Request) dispatch.Request {
	return ResolveRequest(r)
}

func (d defaultDispatcherAdapter) WriteResponse(w http.ResponseWriter, r dispatch.Response) {
	WriteResponse(w, r)
}
