package rdispatch

import (
	"net/http"

	"github.com/huangml/dispatch"
)

type RemoteDispatcher struct {
	dispatch.Dispatcher
	adapter RemoteDispatcherAdapter
}

func (d *RemoteDispatcher) ServeHTTP(w http.ResponseWriter, r *http.Request) {

}
