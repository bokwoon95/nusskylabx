package advisers

import (
	"net/http"

	sq "github.com/bokwoon95/go-structured-query/postgres"
	"github.com/bokwoon95/nusskylabx/app/skylab"
	"github.com/bokwoon95/nusskylabx/tables"

	"github.com/bokwoon95/nusskylabx/helpers/headers"
)

func (adv Advisers) MilestoneTeamEvaluation(section string) http.HandlerFunc {
	milestone := milestoneFromSection(section)
	return func(w http.ResponseWriter, r *http.Request) {
		adv.skylb.Log.TraceRequest(r)
		user, _ := r.Context().Value(skylab.ContextUser).(skylab.User)
		r = adv.skylb.SetRoleSection(w, r, skylab.RoleAdviser, section)
		headers.DoNotCache(w)
		var evaluationGroups [][]skylab.TeamEvaluation
		t, te := tables.TEAMS(), tables.V_TEAM_EVALUATIONS()
		evaluation := &skylab.TeamEvaluation{}
		// teamGroups maps a team to the list of evaluations evaluating
		// it. The key is the teamID, the value is the index of the slice of
		// team evaluations in evaluationGroups.
		teamGroups := make(map[int]int)
		advisers_teams := sq.
			Select(t.TEAM_ID).
			From(t).
			Where(t.ADVISER_USER_ROLE_ID.EqInt(user.Roles[skylab.RoleAdviser])).
			CTE("advisers_teams")
		err := sq.WithDefaultLog(sq.Lstats).
			With(advisers_teams).
			From(te).
			Where(
				te.MILESTONE.EqString(milestone),
				te.EVALUATEE_TEAM_ID.In(sq.Select(advisers_teams["team_id"]).From(advisers_teams)),
				te.EVALUATOR_TEAM_ID.In(sq.Select(advisers_teams["team_id"]).From(advisers_teams)),
			).
			OrderBy(
				te.EVALUATEE_TEAM_ID,
				te.EVALUATOR_TEAM_ID,
			).
			Selectx(
				evaluation.RowMapper(te),
				func() {
					if i, ok := teamGroups[evaluation.Evaluatee.Team.TeamID]; ok {
						// append to an existing evaluation group
						evaluationGroups[i] = append(evaluationGroups[i], *evaluation)
					} else {
						// create a new evaluation group and append it in
						evaluationGroups = append(evaluationGroups, []skylab.TeamEvaluation{*evaluation})
						teamGroups[evaluation.Evaluatee.Team.TeamID] = len(evaluationGroups) - 1
					}
				},
			).
			Fetch(adv.skylb.DB)
		if err != nil {
			adv.skylb.InternalServerError(w, r, err)
			return
		}
		data := make(map[string]interface{})
		data["Milestone"] = milestone
		data["EvaluationGroups"] = evaluationGroups
		adv.skylb.Wender(w, r, data, "app/advisers/milestone_team_evaluation.html")
	}
}
