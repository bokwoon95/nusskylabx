package advisers

import (
	"net/http"

	sq "github.com/bokwoon95/go-structured-query/postgres"
	"github.com/bokwoon95/nusskylabx/tables"

	"github.com/bokwoon95/nusskylabx/app/skylab"

	"github.com/bokwoon95/nusskylabx/helpers/headers"
)

func (adv Advisers) MilestoneUserEvaluation(section string) http.HandlerFunc {
	milestone := milestoneFromSection(section)
	return func(w http.ResponseWriter, r *http.Request) {
		adv.skylb.Log.TraceRequest(r)
		user, _ := r.Context().Value(skylab.ContextUser).(skylab.User)
		r = adv.skylb.SetRoleSection(w, r, skylab.RoleAdviser, section)
		headers.DoNotCache(w)
		var evaluations []skylab.UserEvaluation
		ue := tables.V_USER_EVALUATIONS()
		userEvaluation := &skylab.UserEvaluation{}
		err := sq.WithDefaultLog(sq.Lstats).
			From(ue).
			Where(
				ue.EVALUATOR_USER_ID.EqInt(user.Roles[skylab.RoleAdviser]),
				ue.COHORT.EqString(adv.skylb.CurrentCohort()),
				ue.MILESTONE.EqString(milestone),
			).
			Selectx(
				userEvaluation.RowMapper(ue),
				func() { evaluations = append(evaluations, *userEvaluation) },
			).
			Fetch(adv.skylb.DB)

		if err != nil {
			adv.skylb.InternalServerError(w, r, err)
			return
		}
		data := map[string]interface{}{
			"Milestone":   milestone,
			"Evaluations": evaluations,
			"Period":      skylab.Period{},
		}
		if len(evaluations) > 0 {
			data["Period"] = evaluations[0].EvaluationForm.Period
		}
		adv.skylb.Render(w, r, data, "app/advisers/milestone_user_evaluation.html")
	}
}
