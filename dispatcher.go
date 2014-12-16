package rdispatch

import (
	"net/http"

	"github.com/huangml/dispatch"
)

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
			return &dispatch.SimpleResponse{Err: statusError{http.StatusBadRequest, ""}}
		}

		if d.adapter.Method(r) == MethodSend {
			err := d.Send(rr)
			return &dispatch.SimpleResponse{Err: ToStatusError(err)}
		}

		// MethodCall TODO: TimeOut
		return d.Call(dispatch.NewContext(), rr)
	}()

	d.adapter.WriteResponse(w, rs)
}
