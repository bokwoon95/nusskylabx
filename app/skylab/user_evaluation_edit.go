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

func (skylb Skylab) UserEvaluationEdit(role string) http.HandlerFunc {
	if !Contains(Roles(), role) {
		panic(fmt.Errorf("%s is not a valid skylab role", role))
	}
	return func(w http.ResponseWriter, r *http.Request) {
		skylb.Log.TraceRequest(r)
		r = skylb.SetRoleSection(w, r, role, SectionPreserve)
		var data UserEvaluationEdit
		var msgs = make(map[string][]string)
		userEvaluationID, err := urlparams.Int(r, "userEvaluationID")
		if err != nil {
			skylb.InternalServerError(w, r, err)
			return
		}
		ue := tables.V_USER_EVALUATIONS()
		err = sq.WithLog(skylb.Log, sq.Lstats).
			From(ue).
			Where(ue.USER_EVALUATION_ID.EqInt(userEvaluationID)).
			SelectRowx((&data.Evaluation).RowMapper(ue)).
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
			data.SubmissionURL = AdviserSubmission + "/" + strconv.Itoa(data.Evaluation.Evaluatee.SubmissionID)
			data.PreviewURL = AdviserUserEvaluation + "/" + strconv.Itoa(userEvaluationID) + "/preview"
			data.UpdateURL = AdviserUserEvaluation + "/" + strconv.Itoa(userEvaluationID) + "/update"
			data.SubmitURL = AdviserUserEvaluation + "/" + strconv.Itoa(userEvaluationID) + "/submit"
		}
		funcs := formx.Funcs(nil, skylb.Policy)
		r, _ = skylb.SetFlashMsgs(w, r, msgs)
		skylb.Render(w, r, data, funcs, "app/skylab/user_evaluation_edit.html", "helpers/formx/render_form.html")
	}
}
