package app

import (
	"net/http"
)

func (ap App) Landing(w http.ResponseWriter, r *http.Request) {
	ap.skylb.Log.TraceRequest(r)
	ap.skylb.Render(w, r, nil, "app/landing.html")
}
