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

	var teamFeedbacks []skylab.TeamFeedback
	var userFeedbacks []skylab.UserFeedback
	data := map[string]interface{}{
		"TeamFeedbacks": teamFeedbacks,
		"UserFeedbacks": userFeedbacks,
	}
	stu.skylb.Render(w, r, data, "app/students/list_feedbacks.html")
}
