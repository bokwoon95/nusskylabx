package students

import (
	"net/http"

	"github.com/bokwoon95/nusskylabx/app/skylab"
	"github.com/bokwoon95/nusskylabx/helpers/headers"
)

func (stu Students) ListFeedbacks(w http.ResponseWriter, r *http.Request) {
	stu.skylb.Log.TraceRequest(r)
	r = stu.skylb.SetRoleSection(w, r, skylab.RoleStudent, skylab.StudentListFeedbacks)
	headers.DoNotCache(w)

	type Data struct {
		TeamFeedbacks []skylab.TeamFeedback
		UserFeedbacks []skylab.UserFeedback
	}
	var data Data
	stu.skylb.Render(w, r, data, nil, "app/students/list_feedbacks.html")
}
