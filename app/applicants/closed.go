package applicants

import (
	"net/http"
)

func (apt Applicants) Closed(w http.ResponseWriter, r *http.Request) {
	apt.skylb.Log.TraceRequest(r)
	apt.skylb.Render(w, r, nil, nil, "app/applicants/Closed.html")
}
