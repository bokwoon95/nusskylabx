package app

import (
	"net/http"
)

func (ap App) LoginPage(w http.ResponseWriter, r *http.Request) {
	ap.skylb.Log.TraceRequest(r)
	ap.skylb.Wender(w, r, nil, "app/login_page.html")
}
