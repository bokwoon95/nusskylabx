package advisers

import (
	"net/http"

	sq "github.com/bokwoon95/go-structured-query/postgres"
	"github.com/bokwoon95/nusskylabx/helpers/timeutil"
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
		type Data struct {
			Milestone   string
			Evaluations []skylab.UserEvaluation
			Period      skylab.Period
		}
		var data Data
		data.Milestone = milestone
		ue := tables.V_USER_EVALUATIONS()
		userEvaluation := &skylab.UserEvaluation{}
		err := sq.WithLog(adv.skylb.Log, sq.Lstats).
			From(ue).
			Where(
				ue.EVALUATOR_USER_ID.EqInt(user.Roles[skylab.RoleAdviser]),
				ue.COHORT.EqString(adv.skylb.CurrentCohort()),
				ue.MILESTONE.EqString(milestone),
			).
			Selectx(
				userEvaluation.RowMapper(ue),
				func() { data.Evaluations = append(data.Evaluations, *userEvaluation) },
			).
			Fetch(adv.skylb.DB)

		if err != nil {
			adv.skylb.InternalServerError(w, r, err)
			return
		}
		if len(data.Evaluations) > 0 {
			data.Period = data.Evaluations[0].EvaluationForm.Period
		}
		funcs := timeutil.Funcs(nil)
		adv.skylb.Render(w, r, data, funcs, "app/advisers/milestone_user_evaluation.html")
	}
}
