package students

import (
	"net/http"

	"github.com/bokwoon95/nusskylabx/app/skylab"
)

func (stu Students) Dashboard(w http.ResponseWriter, r *http.Request) {
	stu.skylb.Log.TraceRequest(r)
	r = stu.skylb.SetRoleSection(w, r, skylab.RoleStudent, skylab.StudentDashboard)
	stu.skylb.Render(w, r, nil, nil, "app/students/dashboard.html")
}
