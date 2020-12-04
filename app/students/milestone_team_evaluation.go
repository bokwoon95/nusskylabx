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
		team_evaluations, user_roles_students := tables.V_TEAM_EVALUATIONS(), tables.USER_ROLES_STUDENTS()
		var evaluation skylab.TeamEvaluation
		var evaluations []skylab.TeamEvaluation
		err := sq.WithDefaultLog(sq.Lstats).
			From(team_evaluations).
			Where(
				team_evaluations.COHORT.EqString(stu.skylb.CurrentCohort()),
				team_evaluations.STAGE.EqString(skylab.StageEvaluation),
				team_evaluations.MILESTONE.EqString(milestone),
				team_evaluations.EVALUATOR_TEAM_ID.In(sq.
					Select(user_roles_students.TEAM_ID).
					From(user_roles_students).
					Where(user_roles_students.USER_ROLE_ID.EqInt(user.Roles[skylab.RoleStudent])),
				)).
			Selectx(evaluation.RowMapper(team_evaluations), func() {
				evaluations = append(evaluations, evaluation)
			}).
			Fetch(stu.skylb.DB)
		if err != nil {
			stu.skylb.InternalServerError(w, r, err)
			return
		}
		data := map[string]interface{}{
			"Milestone":   milestone,
			"Evaluations": evaluations,
		}
		stu.skylb.Wender(w, r, data, "app/students/milestone_team_evaluation.html")
	}
}
