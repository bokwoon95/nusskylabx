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

func (skylb Skylab) TeamEvaluationView(role string) http.HandlerFunc {
	if !Contains(Roles(), role) {
		panic(fmt.Errorf("%s is not a valid skylab role", role))
	}
	return func(w http.ResponseWriter, r *http.Request) {
		skylb.Log.TraceRequest(r)
		r = skylb.SetRoleSection(w, r, role, SectionPreserve)
		msgs := make(map[string][]string)
		teamEvaluationID, err := urlparams.Int(r, "teamEvaluationID")
		if err != nil {
			skylb.BadRequest(w, r, err.Error())
			return
		}
		te := tables.V_TEAM_EVALUATIONS()
		var teamEvaluation TeamEvaluation
		var submitURL, submissionURL, editURL string
		err = sq.WithDefaultLog(sq.Lstats).
			From(te).
			Where(te.TEAM_EVALUATION_ID.EqInt(teamEvaluationID)).
			SelectRowx(teamEvaluation.RowMapper(te)).
			Fetch(skylb.DB)
		if err != nil {
			switch {
			case errors.Is(err, sql.ErrNoRows):
				skylb.BadRequest(w, r, fmt.Sprintf("No team evaluation found for teamEvaluationID %d", teamEvaluationID))
			default:
				skylb.InternalServerError(w, r, err)
			}
			return
		}
		canEdit, _ := r.Context().Value(ContextCanEditEvaluation).(bool)
		switch role {
		case RoleStudent:
			if canEdit {
				submitURL = StudentTeamEvaluation + "/" + strconv.Itoa(teamEvaluationID) + "/submit"
				editURL = StudentTeamEvaluation + "/" + strconv.Itoa(teamEvaluationID) + "/edit"
			}
			submissionURL = StudentSubmission + "/" + strconv.Itoa(teamEvaluation.Evaluatee.SubmissionID)
		case RoleAdviser:
			submissionURL = AdviserSubmission + "/" + strconv.Itoa(teamEvaluation.Evaluatee.SubmissionID)
		}
		r, _ = skylb.SetFlashMsgs(w, r, msgs)
		data := map[string]interface{}{
			"TeamEvaluation": teamEvaluation,
			"SubmitURL":      submitURL,
			"SubmissionURL":  submissionURL,
			"EditURL":        editURL,
		}
		skylb.Wender(w, r, data, "app/skylab/team_evaluation_view.html")
	}
}
