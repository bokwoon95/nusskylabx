package students

import (
	"database/sql"
	"errors"
	"net/http"

	sq "github.com/bokwoon95/go-structured-query/postgres"
	"github.com/bokwoon95/nusskylabx/app/skylab"
	"github.com/bokwoon95/nusskylabx/tables"
)

func (stu Students) Team(w http.ResponseWriter, r *http.Request) {
	stu.skylb.Log.TraceRequest(r)
	r = stu.skylb.SetRoleSection(w, r, skylab.RoleStudent, skylab.StudentTeam)
	type Data struct {
		Team skylab.Team
	}
	var data Data
	user, _ := r.Context().Value(skylab.ContextUser).(skylab.User)
	studentUserRoleID := user.Roles[skylab.RoleStudent]
	t := tables.V_TEAMS()
	err := sq.WithLog(stu.skylb.Log, sq.Lverbose).
		From(t).
		Where(sq.Int(studentUserRoleID).In(sq.Fields{t.STUDENT1_USER_ROLE_ID, t.STUDENT2_USER_ID})).
		SelectRowx((&data.Team).RowMapper(t)).
		Fetch(stu.skylb.DB)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			stu.StudentNoTeam(w, r)
		default:
			stu.skylb.InternalServerError(w, r, err)
		}
		return
	}
	stu.skylb.Render(w, r, data, nil, "app/students/team.html")
}
