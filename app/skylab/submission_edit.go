package skylab

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	sq "github.com/bokwoon95/go-structured-query/postgres"
	"github.com/bokwoon95/nusskylabx/helpers/flash"
	"github.com/bokwoon95/nusskylabx/helpers/headers"
	"github.com/bokwoon95/nusskylabx/helpers/urlparams"
	"github.com/bokwoon95/nusskylabx/tables"
)

func (skylb Skylab) SubmissionEdit(role string) http.HandlerFunc {
	if !Contains(Roles(), role) {
		panic(fmt.Errorf("%s is not a valid skylab role", role))
	}
	return func(w http.ResponseWriter, r *http.Request) {
		skylb.Log.TraceRequest(r)
		headers.DoNotCache(w)
		msgs := make(map[string][]string)
		submissionID, err := urlparams.Int(r, "submissionID")
		if err != nil {
			skylb.BadRequest(w, r, err.Error())
			return
		}
		var submission Submission
		var peerEvaluations []TeamEvaluation
		var adviserEvaluation, mentorEvaluation UserEvaluation
		var previewURL, updateURL, submitURL string
		render := func() {
			r = skylb.SetRoleSection(w, r, role, skylb.getSectionFromSubmissionID(submissionID, role))
			r, _ = skylb.SetFlashMsgs(w, r, msgs)
			data := map[string]interface{}{
				"Submission":        submission,
				"PeerEvaluations":   peerEvaluations,
				"AdviserEvaluation": adviserEvaluation,
				"MentorEvaluation":  mentorEvaluation,
				"PreviewURL":        previewURL,
				"UpdateURL":         updateURL,
				"SubmitURL":         submitURL,
			}
			skylb.Wender(w, r, data, "app/skylab/submission_edit.html")
		}
		s := tables.V_SUBMISSIONS()
		err = sq.WithDefaultLog(sq.Lstats).
			From(s).
			Where(s.SUBMISSION_ID.EqInt(submissionID)).
			SelectRowx(submission.RowMapper(s)).
			Fetch(skylb.DB)
		if err != nil {
			switch {
			case errors.Is(err, sql.ErrNoRows):
				skylb.BadRequest(w, r, fmt.Sprintf("No submission found for submissionID %d", submissionID))
			default:
				msgs[flash.Error] = []string{err.Error()}
				w.WriteHeader(http.StatusInternalServerError)
				render()
			}
			return
		}
		switch role {
		case RoleStudent:
			previewURL = StudentSubmission + "/" + strconv.Itoa(submissionID) + "/preview"
			updateURL = StudentSubmission + "/" + strconv.Itoa(submissionID) + "/update"
			submitURL = StudentSubmission + "/" + strconv.Itoa(submissionID) + "/submit"
		}

		// Team evaluations
		te := tables.V_TEAM_EVALUATIONS()
		var teamEvaluation TeamEvaluation
		err = sq.WithDefaultLog(sq.Lstats).
			From(te).
			Where(te.SUBMISSION_ID.EqInt(submissionID)).
			Selectx(teamEvaluation.RowMapper(te), func() {
				peerEvaluations = append(peerEvaluations, teamEvaluation)
			}).
			Fetch(skylb.DB)
		if err != nil {
			skylb.InternalServerError(w, r, err)
			return
		}

		// Adviser + Mentor Evaluations
		ue := tables.V_USER_EVALUATIONS()
		var userEvaluation UserEvaluation
		err = sq.WithDefaultLog(sq.Lstats).
			From(ue).
			Where(ue.SUBMISSION_ID.EqInt(submissionID)).
			Selectx(userEvaluation.RowMapper(ue), func() {
				switch userEvaluation.Role {
				case RoleAdviser:
					adviserEvaluation = userEvaluation
				case RoleMentor:
					mentorEvaluation = userEvaluation
				}
			}).
			Fetch(skylb.DB)
		if err != nil {
			skylb.InternalServerError(w, r, err)
			return
		}
		render()
	}
}
