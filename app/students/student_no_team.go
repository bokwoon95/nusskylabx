package students

import (
	"net/http"

	"github.com/bokwoon95/nusskylabx/app/skylab"
)

func (stu Students) StudentNoTeam(w http.ResponseWriter, r *http.Request) {
	stu.skylb.Log.TraceRequest(r)
	r = stu.skylb.SetRoleSection(w, r, skylab.RoleStudent, skylab.SectionPreserve)
	stu.skylb.Render(w, r, nil, nil, "app/students/student_no_team.html")
}
