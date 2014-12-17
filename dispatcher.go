package rdispatch

import (
	"net/http"

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
		adapter = &defaultDispatcherAdapter{}
	}

	return &RemoteDispatcher{
		Dispatcher: d,
		adapter:    adapter,
	}
}

func (d *RemoteDispatcher) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rr := d.adapter.ResolveRequest(r)

	rs := func() dispatch.Response {
		if rr == nil {
			return dispatch.NewSimpleResponse(nil, statusError{http.StatusBadRequest, ""})
		}

		if d.adapter.Method(r) == MethodSend {
			if err := d.Send(rr); err != nil {
				return dispatch.NewSimpleResponse(nil, nil) // TODO: Send returns StatusAccepted.
			} else {
				return dispatch.NewSimpleResponse(nil, ToStatusError(err))
			}
		}

		// MethodCall TODO: TimeOut
		return d.Call(dispatch.NewContext(), rr)
	}()

	d.adapter.WriteResponse(w, rs)
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
