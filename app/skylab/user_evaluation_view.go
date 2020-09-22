package skylab

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	sq "github.com/bokwoon95/go-structured-query/postgres"
	"github.com/bokwoon95/nusskylabx/helpers/flash"
	"github.com/bokwoon95/nusskylabx/helpers/formx"
	"github.com/bokwoon95/nusskylabx/helpers/urlparams"
	"github.com/bokwoon95/nusskylabx/tables"
)

func (skylb Skylab) UserEvaluationView(role string) http.HandlerFunc {
	if !Contains(Roles(), role) {
		panic(fmt.Errorf("%s is not a valid skylab role", role))
	}
	return func(w http.ResponseWriter, r *http.Request) {
		skylb.Log.TraceRequest(r)
		r = skylb.SetRoleSection(w, r, role, SectionPreserve)
		var data UserEvaluationView
		var msgs = make(map[string][]string)
		userEvaluationID, err := urlparams.Int(r, "userEvaluationID")
		if err != nil {
			skylb.BadRequest(w, r, err.Error())
			return
		}
		render := func(data UserEvaluationView, msgs map[string][]string) {
			funcs := formx.Funcs(nil, skylb.Policy)
			r, _ = skylb.SetFlashMsgs(w, r, msgs)
			skylb.Render(w, r, data, funcs, "app/skylab/user_evaluation_view.html", "helpers/formx/render_form_results.html")
		}
		ue := tables.V_USER_EVALUATIONS()
		err = sq.WithDefaultLog(sq.Lstats).
			From(ue).
			Where(ue.USER_EVALUATION_ID.EqInt(userEvaluationID)).
			SelectRowx((&data.Evaluation).RowMapper(ue)).
			Fetch(skylb.DB)
		if err != nil {
			switch {
			case errors.Is(err, sql.ErrNoRows):
				skylb.BadRequest(w, r, fmt.Sprintf("No user evaluation found for userEvaluationID %d", userEvaluationID))
			default:
				msgs[flash.Error] = []string{err.Error()}
				w.WriteHeader(http.StatusInternalServerError)
				render(data, msgs)
			}
			return
		}
		canEdit, _ := r.Context().Value(ContextCanEditEvaluation).(bool)
		switch role {
		case RoleStudent:
			data.SubmissionURL = "google.com"
			if canEdit {
				data.EditURL = "google.com"
				data.SubmitURL = "google.com"
			}
		case RoleAdviser:
			data.SubmissionURL = AdviserSubmission + "/" + strconv.Itoa(data.Evaluation.Evaluatee.SubmissionID)
			data.EditURL = AdviserUserEvaluation + "/" + strconv.Itoa(userEvaluationID) + "/edit"
			data.SubmitURL = AdviserUserEvaluation + "/" + strconv.Itoa(userEvaluationID) + "/submit"
		}
		render(data, msgs)
	}
}
