package students

import (
	"net/http"

	sq "github.com/bokwoon95/go-structured-query/postgres"
	"github.com/bokwoon95/nusskylabx/app/skylab"
	"github.com/bokwoon95/nusskylabx/tables"

	"github.com/bokwoon95/nusskylabx/helpers/headers"
)

func (stu Students) MilestoneTeamEvaluation(section string) http.HandlerFunc {
	milestone := milestoneFromSection(section)
	return func(w http.ResponseWriter, r *http.Request) {
		stu.skylb.Log.TraceRequest(r)
		r = stu.skylb.SetRoleSection(w, r, skylab.RoleStudent, section)
		user, _ := r.Context().Value(skylab.ContextUser).(skylab.User)
		headers.DoNotCache(w)
		type Data struct {
			Milestone   string
			Evaluations []skylab.TeamEvaluation
		}
		var data Data
		evaluation := &skylab.TeamEvaluation{}
		data.Milestone = milestone
		te, urs := tables.V_TEAM_EVALUATIONS(), tables.USER_ROLES_STUDENTS()
		err := sq.WithDefaultLog(sq.Lstats).
			From(te).
			Where(
				te.COHORT.EqString(stu.skylb.CurrentCohort()),
				te.STAGE.EqString(skylab.StageEvaluation),
				te.MILESTONE.EqString(milestone),
				te.EVALUATOR_TEAM_ID.In(
					sq.Select(urs.TEAM_ID).From(urs).Where(urs.USER_ROLE_ID.EqInt(user.Roles[skylab.RoleStudent])),
				)).
			Selectx(
				evaluation.RowMapper(te),
				func() { data.Evaluations = append(data.Evaluations, *evaluation) },
			).
			Fetch(stu.skylb.DB)
		if err != nil {
			stu.skylb.InternalServerError(w, r, err)
			return
		}
		stu.skylb.Render(w, r, data, nil, "app/students/milestone_team_evaluation.html")
	}
}
