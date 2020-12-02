package app

import (
	"net/http"

	"github.com/bokwoon95/nusskylabx/app/skylab"
	"github.com/bokwoon95/nusskylabx/helpers/urlparams"
)

func (ap App) User(w http.ResponseWriter, r *http.Request) {
	ap.skylb.Log.TraceRequest(r)
	r = ap.skylb.SetRoleSection(w, r, skylab.RolePreserve, "")
	asUser := r.FormValue("user") != ""
	asAdmin := r.FormValue("admin") != ""
	if (!asUser && !asAdmin) || (asUser && asAdmin) {
		asUser = true
	}
	user, _ := r.Context().Value(skylab.ContextUser).(skylab.User)
	admin, _ := r.Context().Value(skylab.ContextAdmin).(skylab.User)
	data := make(map[string]interface{})
	if asUser {
		data["User"] = user
		data["Role"] = "user"
	}
	if asAdmin {
		data["User"] = admin
		data["Role"] = "admin"
	}
	ap.skylb.Wender(w, r, "app/user.html", data)
}

func (ap App) UserUpdate(w http.ResponseWriter, r *http.Request) {
	ap.skylb.Log.TraceRequest(r)
	userID, err := urlparams.Int(r, "userID")
	if err != nil {
		ap.skylb.BadRequest(w, r, err.Error())
		return
	}
	displayname := r.FormValue("displayname")
	_, err = ap.skylb.DB.Exec(`UPDATE users SET displayname = $1 WHERE user_id = $2`, displayname, userID)
	if err != nil {
		ap.skylb.InternalServerError(w, r, err)
		return
	}
	var param string
	if role := r.FormValue("role"); role != "" {
		param = "?" + role + "=true"
	}
	http.Redirect(w, r, "/user"+param, http.StatusMovedPermanently)
}
