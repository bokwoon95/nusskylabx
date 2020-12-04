package skylab

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	sq "github.com/bokwoon95/go-structured-query/postgres"
	"github.com/bokwoon95/nusskylabx/helpers/flash"
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
		msgs := make(map[string][]string)
		userEvaluationID, err := urlparams.Int(r, "userEvaluationID")
		if err != nil {
			skylb.BadRequest(w, r, err.Error())
			return
		}
		var evaluation UserEvaluation
		var submissionURL, submitURL, editURL string
		render := func() {
			r, _ = skylb.SetFlashMsgs(w, r, msgs)
			data := map[string]interface{}{
				"Evaluation":    evaluation,
				"SubmissionURL": submissionURL,
				"SubmitURL":     submitURL,
				"EditURL":       editURL,
			}
			skylb.Wender(w, r, data, "app/skylab/user_evaluation_view.html")
		}
		ue := tables.V_USER_EVALUATIONS()
		err = sq.WithDefaultLog(sq.Lstats).
			From(ue).
			Where(ue.USER_EVALUATION_ID.EqInt(userEvaluationID)).
			SelectRowx(evaluation.RowMapper(ue)).
			Fetch(skylb.DB)
		if err != nil {
			switch {
			case errors.Is(err, sql.ErrNoRows):
				skylb.BadRequest(w, r, fmt.Sprintf("No user evaluation found for userEvaluationID %d", userEvaluationID))
			default:
				msgs[flash.Error] = []string{err.Error()}
				w.WriteHeader(http.StatusInternalServerError)
				render()
			}
			return
		}
		canEdit, _ := r.Context().Value(ContextCanEditEvaluation).(bool)
		switch role {
		case RoleStudent:
			submissionURL = "google.com"
			if canEdit {
				editURL = "google.com"
				submitURL = "google.com"
			}
		case RoleAdviser:
			submissionURL = AdviserSubmission + "/" + strconv.Itoa(evaluation.Evaluatee.SubmissionID)
			editURL = AdviserUserEvaluation + "/" + strconv.Itoa(userEvaluationID) + "/edit"
			submitURL = AdviserUserEvaluation + "/" + strconv.Itoa(userEvaluationID) + "/submit"
		}
		render()
	}
}
