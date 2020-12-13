package skylab

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	sq "github.com/bokwoon95/go-structured-query/postgres"
	"github.com/bokwoon95/nusskylabx/helpers/urlparams"
	"github.com/bokwoon95/nusskylabx/tables"
)

func (skylb Skylab) UserEvaluationEdit(role string) http.HandlerFunc {
	if !Contains(Roles(), role) {
		panic(fmt.Errorf("%s is not a valid skylab role", role))
	}
	return func(w http.ResponseWriter, r *http.Request) {
		skylb.Log.TraceRequest(r)
		r = skylb.SetRoleSection(w, r, role, SectionPreserve)
		msgs := make(map[string][]string)
		userEvaluationID, err := urlparams.Int(r, "userEvaluationID")
		if err != nil {
			skylb.InternalServerError(w, r, err)
			return
		}
		ue := tables.V_USER_EVALUATIONS()
		var evaluation UserEvaluation
		var submissionURL, submitURL, previewURL, updateURL string
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
				skylb.InternalServerError(w, r, err)
			}
			return
		}
		switch role {
		case RoleStudent:
		case RoleAdviser:
			submissionURL = AdviserSubmission + "/" + strconv.Itoa(evaluation.Evaluatee.SubmissionID)
			previewURL = AdviserUserEvaluation + "/" + strconv.Itoa(userEvaluationID) + "/preview"
			updateURL = AdviserUserEvaluation + "/" + strconv.Itoa(userEvaluationID) + "/update"
			submitURL = AdviserUserEvaluation + "/" + strconv.Itoa(userEvaluationID) + "/submit"
		}
		r, _ = skylb.SetFlashMsgs(w, r, msgs)
		data := map[string]interface{}{
			"Evaluation":    evaluation,
			"SubmissionURL": submissionURL,
			"SubmitURL":     submitURL,
			"PreviewURL":    previewURL,
			"UpdateURL":     updateURL,
		}
		skylb.Render(w, r, data, "app/skylab/user_evaluation_edit.html")
	}
}
