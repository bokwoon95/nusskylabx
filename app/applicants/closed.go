package applicants

import (
	"net/http"
)

func (apt Applicants) Closed(w http.ResponseWriter, r *http.Request) {
	apt.skylb.Log.TraceRequest(r)
	apt.skylb.Wender(w, r, nil, "app/applicants/closed.html")
}
