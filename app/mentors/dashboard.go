package mentors

import (
	"net/http"

	"github.com/bokwoon95/nusskylabx/app/skylab"
)

func (mnt Mentors) Dashboard(w http.ResponseWriter, r *http.Request) {
	mnt.skylb.Log.TraceRequest(r)
	r = mnt.skylb.SetRoleSection(w, r, skylab.RoleMentor, skylab.MentorDashboard)
	mnt.skylb.Render(w, r, nil, nil, "app/mentors/dashboard.html")
}
