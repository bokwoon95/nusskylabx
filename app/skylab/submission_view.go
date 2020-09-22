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

func (skylb Skylab) SubmissionView(role string) http.HandlerFunc {
	if !Contains(Roles(), role) {
		panic(fmt.Errorf("%s is not a valid skylab role", role))
	}
	return func(w http.ResponseWriter, r *http.Request) {
		skylb.Log.TraceRequest(r)
		var data SubmissionViewData
		var msgs = make(map[string][]string)
		submissionID, err := urlparams.Int(r, "submissionID")
		if err != nil {
			skylb.BadRequest(w, r, err.Error())
			return
		}
		render := func(data SubmissionViewData, msgs map[string][]string) {
			funcs := formx.Funcs(nil, skylb.Policy)
			r = skylb.SetRoleSection(w, r, role, skylb.getSectionFromSubmissionID(submissionID, role))
			r, _ = skylb.SetFlashMsgs(w, r, msgs)
			skylb.Render(w, r, data, funcs, "app/skylab/submission_view.html", "helpers/formx/render_form_results.html")
		}
		s := tables.V_SUBMISSIONS()
		err = sq.WithDefaultLog(sq.Lverbose).
			From(s).
			Where(s.SUBMISSION_ID.EqInt(submissionID)).
			SelectRowx((&data.Submission).RowMapper(s)).
			Fetch(skylb.DB)
		if err != nil {
			switch {
			case errors.Is(err, sql.ErrNoRows):
				skylb.BadRequest(w, r, fmt.Sprintf("No submission found for submissionID %d", submissionID))
			default:
				msgs[flash.Error] = []string{err.Error()}
				w.WriteHeader(http.StatusInternalServerError)
				render(data, msgs)
			}
			return
		}
		canEdit, _ := r.Context().Value(ContextCanEditSubmission).(bool)
		if canEdit {
			switch role {
			case RoleStudent:
				data.EditURL = StudentSubmission + "/" + strconv.Itoa(submissionID) + "/edit"
				data.SubmitURL = StudentSubmission + "/" + strconv.Itoa(submissionID) + "/submit"
			case RoleAdviser:
			}
		}
		render(data, msgs)
	}
}

func (skylb Skylab) getSectionFromSubmissionID(submissionID int, role string) (section string) {
	s, f, p := tables.SUBMISSIONS(), tables.FORMS(), tables.PERIODS()
	var milestone string
	_ = sq.From(s).
		Join(f, f.FORM_ID.Eq(s.SUBMISSION_FORM_ID)).
		Join(p, p.PERIOD_ID.Eq(f.PERIOD_ID)).
		Where(s.SUBMISSION_ID.EqInt(submissionID)).
		SelectRowx(func(row *sq.Row) { milestone = row.String(p.MILESTONE) }).
		Fetch(skylb.DB)
	switch role {
	case RoleStudent:
		switch milestone {
		case Milestone1:
			section = StudentM1Submission
		case Milestone2:
			section = StudentM2Submission
		case Milestone3:
			section = StudentM3Submission
		default:
			section = StudentM1Submission
		}
	case RoleAdviser:
		section = SectionPreserve
	}
	if section == "" {
		section = SectionPreserve
	}
	return section
}
