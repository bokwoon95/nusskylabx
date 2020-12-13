package advisers

import (
	"net/http"
	"strconv"

	sq "github.com/bokwoon95/go-structured-query/postgres"
	"github.com/bokwoon95/nusskylabx/app/skylab"
	"github.com/bokwoon95/nusskylabx/tables"

	"github.com/bokwoon95/nusskylabx/helpers/formutil"
	"github.com/bokwoon95/nusskylabx/helpers/headers"
)

func (adv Advisers) EvaluatorEvaluatees(w http.ResponseWriter, r *http.Request) {
	adv.skylb.Log.TraceRequest(r)
	r = adv.skylb.SetRoleSection(w, r, skylab.RoleAdviser, skylab.AdviserEvaluatorEvaluatees)
	headers.DoNotCache(w)
	evaluatorEvaluatees := make(map[int]map[int]bool)
	user, _ := r.Context().Value(skylab.ContextUser).(skylab.User)

	// get list of teamIDs under adviser
	adviserTeamIDs, err := adv.getTeamIDs(user)
	if err != nil {
		adv.skylb.InternalServerError(w, r, err)
		return
	}

	// fill in evaluatorEvaluatees accordingly
	tp, t := tables.TEAM_EVALUATION_PAIRS(), tables.TEAMS()
	var evaluatorTeamID, evaluateeTeamID int
	err = sq.WithDefaultLog(sq.Lverbose).
		From(tp).
		Join(t, t.TEAM_ID.Eq(tp.EVALUATOR_TEAM_ID)).
		Where(
			tp.EVALUATEE_TEAM_ID.In(adviserTeamIDs),
			tp.EVALUATOR_TEAM_ID.In(adviserTeamIDs),
		).
		OrderBy(
			sq.Fieldf("array_position(ARRAY[?], ?)", skylab.ProjectLevels(), t.PROJECT_LEVEL),
		).
		Selectx(func(row *sq.Row) {
			evaluatorTeamID = row.Int(tp.EVALUATOR_TEAM_ID)
			evaluateeTeamID = row.Int(tp.EVALUATEE_TEAM_ID)
		}, func() {
			if evaluatorEvaluatees[evaluatorTeamID] == nil {
				evaluatorEvaluatees[evaluatorTeamID] = make(map[int]bool)
			}
			// Check this pair of evaluator and evaluatee
			evaluatorEvaluatees[evaluatorTeamID][evaluateeTeamID] = true
		}).
		Fetch(adv.skylb.DB)
	if err != nil {
		adv.skylb.InternalServerError(w, r, err)
		return
	}

	for i := range adviserTeamIDs {
		evaluatorTeamID := adviserTeamIDs[i]
		evaluateeTeamIDs := append(append([]int{}, adviserTeamIDs[:i]...), adviserTeamIDs[i+1:]...)
		for j := range evaluateeTeamIDs {
			evaluateeTeamID := evaluateeTeamIDs[j]
			if evaluatorEvaluatees[evaluatorTeamID] == nil {
				evaluatorEvaluatees[evaluatorTeamID] = make(map[int]bool)
			}
			if !evaluatorEvaluatees[evaluatorTeamID][evaluateeTeamID] {
				// Uncheck this pair of evaluator and evaluatee
				evaluatorEvaluatees[evaluatorTeamID][evaluateeTeamID] = false
			}
		}
	}

	vt := tables.V_TEAMS()
	var team skylab.Team
	teams := make(map[int]skylab.Team)
	err = sq.WithDefaultLog(sq.Lverbose).
		From(vt).
		Where(vt.ADVISER_USER_ROLE_ID.EqInt(user.Roles[skylab.RoleAdviser])).
		Selectx(func(row *sq.Row) {
			team.TeamID = row.Int(vt.TEAM_ID)
			team.Valid = team.TeamID != 0
			team.TeamName = row.String(vt.TEAM_NAME)
			team.ProjectLevel = row.String(vt.PROJECT_LEVEL)
			team.Student1.Displayname = row.String(vt.STUDENT1_DISPLAYNAME)
			team.Student2.Displayname = row.String(vt.STUDENT2_DISPLAYNAME)
		}, func() {
			teams[team.TeamID] = team
		}).
		Fetch(adv.skylb.DB)
	if err != nil {
		adv.skylb.InternalServerError(w, r, err)
		return
	}

	data := make(map[string]interface{})
	data["EvaluatorEvaluatees"] = evaluatorEvaluatees
	data["Teams"] = teams
	adv.skylb.Render(w, r, data, "app/advisers/evaluator_evaluatees.html")
}

func (adv Advisers) EvaluatorEvaluateesUpdate(w http.ResponseWriter, r *http.Request) {
	adv.skylb.Log.TraceRequest(r)
	user, _ := r.Context().Value(skylab.ContextUser).(skylab.User)
	userRoleID := user.Roles[skylab.RoleAdviser]
	if userRoleID == 0 {
		adv.skylb.NotAuthorized(w, r)
		return
	}

	// get list of teamIDs under adviser
	t := tables.TEAMS()
	var adviserTeamIDs []int64
	err := sq.WithDefaultLog(sq.Lstats).
		From(t).
		Where(t.ADVISER_USER_ROLE_ID.EqInt(userRoleID)).
		GroupBy(t.ADVISER_USER_ROLE_ID).
		SelectRowx(func(row *sq.Row) {
			row.ScanArray(&adviserTeamIDs, sq.Fieldf("array_agg(?)", t.TEAM_ID))
		}).
		Fetch(adv.skylb.DB)
	if err != nil {
		adv.skylb.InternalServerError(w, r, err)
		return
	}

	validTeamID := make(map[int]bool)
	for i := range adviserTeamIDs {
		validTeamID[int(adviserTeamIDs[i])] = true
	}
	_ = formutil.ParseForm(r)
	pairings := make(map[int]map[int]bool)
	for key, values := range r.Form {
		evaluatorTeamID, err := strconv.Atoi(key)
		if err != nil || !validTeamID[evaluatorTeamID] {
			continue
		}
		if pairings[evaluatorTeamID] == nil {
			pairings[evaluatorTeamID] = make(map[int]bool)
		}
		for _, value := range values {
			evaluateeTeamID, err := strconv.Atoi(value)
			if err != nil || !validTeamID[evaluatorTeamID] {
				continue
			}
			pairings[evaluatorTeamID][evaluateeTeamID] = true
		}
	}

	var fields sq.Fields
	var values sq.RowValues
	for i := range adviserTeamIDs {
		evaluatorTeamID := int(adviserTeamIDs[i])
		evaluateeTeamIDs := append(append([]int64{}, adviserTeamIDs[:i]...), adviserTeamIDs[i+1:]...)
		for j := range evaluateeTeamIDs {
			evaluateeTeamID := int(evaluateeTeamIDs[j])
			if pairings[evaluatorTeamID][evaluateeTeamID] {
				// pending insertion
				values = append(values, []interface{}{evaluatorTeamID, evaluateeTeamID})
			} else {
				// pending deletion
				fields = append(fields, sq.Fieldf("(?, ?)", evaluatorTeamID, evaluateeTeamID))
			}
		}
	}
	tp := tables.TEAM_EVALUATION_PAIRS()
	_, err = sq.InsertQuery{
		Log:           adv.skylb.Log,
		LogFlag:       sq.Lstats,
		IntoTable:     tp,
		InsertColumns: []sq.Field{tp.EVALUATOR_TEAM_ID, tp.EVALUATEE_TEAM_ID},
		RowValues:     values,
	}.OnConflict().DoNothing().Exec(adv.skylb.DB, sq.ErowsAffected)
	if err != nil {
		adv.skylb.InternalServerError(w, r, err)
		return
	}
	_, err = sq.WithDefaultLog(sq.Lstats).
		DeleteFrom(tp).
		Where(sq.RowValue{tp.EVALUATOR_TEAM_ID, tp.EVALUATEE_TEAM_ID}.In(fields)).
		Exec(adv.skylb.DB, sq.ErowsAffected)
	if err != nil {
		adv.skylb.InternalServerError(w, r, err)
		return
	}
	r, _ = adv.skylb.SetFlashMsgs(w, r, map[string][]string{"success": {"pairings updated!"}})
	http.Redirect(w, r, skylab.AdviserEvaluatorEvaluatees, http.StatusMovedPermanently)
}
