package skylab

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	sq "github.com/bokwoon95/go-structured-query/postgres"
	"github.com/bokwoon95/nusskylabx/helpers/formx"
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
		var data TeamEvaluationView
		var msgs = make(map[string][]string)
		teamEvaluationID, err := urlparams.Int(r, "teamEvaluationID")
		if err != nil {
			skylb.BadRequest(w, r, err.Error())
			return
		}
		te := tables.V_TEAM_EVALUATIONS()
		err = sq.WithDefaultLog(sq.Lstats).
			From(te).
			Where(te.TEAM_EVALUATION_ID.EqInt(teamEvaluationID)).
			SelectRowx((&data.TeamEvaluation).RowMapper(te)).
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
				data.SubmitURL = StudentTeamEvaluation + "/" + strconv.Itoa(teamEvaluationID) + "/submit"
				data.EditURL = StudentTeamEvaluation + "/" + strconv.Itoa(teamEvaluationID) + "/edit"
			}
			data.SubmissionURL = StudentSubmission + "/" + strconv.Itoa(data.TeamEvaluation.Evaluatee.SubmissionID)
		case RoleAdviser:
			data.SubmissionURL = AdviserSubmission + "/" + strconv.Itoa(data.TeamEvaluation.Evaluatee.SubmissionID)
		}
		funcs := formx.Funcs(nil, skylb.Policy)
		r, _ = skylb.SetFlashMsgs(w, r, msgs)
		skylb.Render(w, r, data, funcs, "app/skylab/team_evaluation_view.html", "helpers/formx/render_form_results.html")
	}
}
