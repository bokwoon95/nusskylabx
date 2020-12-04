package advisers

import (
	"database/sql"
	"net/http"
	"strconv"

	sq "github.com/bokwoon95/go-structured-query/postgres"
	"github.com/bokwoon95/nusskylabx/app/skylab"
	"github.com/bokwoon95/nusskylabx/tables"

	"github.com/bokwoon95/nusskylabx/helpers/erro"
	"github.com/bokwoon95/nusskylabx/helpers/flash"
	"github.com/bokwoon95/nusskylabx/helpers/formutil"
	"github.com/bokwoon95/nusskylabx/helpers/headers"
)

func (adv Advisers) EvaluateeEvaluators(w http.ResponseWriter, r *http.Request) {
	adv.skylb.Log.TraceRequest(r)
	r = adv.skylb.SetRoleSection(w, r, skylab.RoleAdviser, skylab.AdviserEvaluateeEvaluators)
	headers.DoNotCache(w)
	evaluateeEvaluators := make(map[int]map[int]bool)
	user, _ := r.Context().Value(skylab.ContextUser).(skylab.User)

	// get list of teamIDs under adviser
	t := tables.TEAMS()
	var adviserTeamIDs []int64
	err := sq.WithDefaultLog(sq.Lstats).
		From(t).
		Where(t.ADVISER_USER_ROLE_ID.EqInt(user.Roles[skylab.RoleAdviser])).
		GroupBy(t.ADVISER_USER_ROLE_ID).
		SelectRowx(func(row *sq.Row) {
			row.ScanArray(&adviserTeamIDs, sq.Fieldf("array_agg(?)", t.TEAM_ID))
		}).
		Fetch(adv.skylb.DB)
	if err != nil {
		adv.skylb.InternalServerError(w, r, err)
		return
	}

	// fill in evaluateeEvaluators accordingly
	tp := tables.TEAM_EVALUATION_PAIRS()
	var evaluateeTeamID, evaluatorTeamID int
	err = sq.WithDefaultLog(sq.Lverbose).
		From(tp).
		Join(t, t.TEAM_ID.Eq(tp.EVALUATEE_TEAM_ID)).
		Where(
			tp.EVALUATEE_TEAM_ID.In(adviserTeamIDs),
			tp.EVALUATOR_TEAM_ID.In(adviserTeamIDs),
		).
		OrderBy(
			sq.Fieldf("array_position(ARRAY[?], ?)", skylab.ProjectLevels(), t.PROJECT_LEVEL),
		).
		Selectx(func(row *sq.Row) {
			evaluateeTeamID = row.Int(tp.EVALUATEE_TEAM_ID)
			evaluatorTeamID = row.Int(tp.EVALUATOR_TEAM_ID)
		}, func() {
			if evaluateeEvaluators[evaluateeTeamID] == nil {
				evaluateeEvaluators[evaluateeTeamID] = make(map[int]bool)
			}
			// Check this pair of evaluatee and evaluator
			evaluateeEvaluators[evaluateeTeamID][evaluatorTeamID] = true
		}).
		Fetch(adv.skylb.DB)
	if err != nil {
		adv.skylb.InternalServerError(w, r, err)
		return
	}

	for i := range adviserTeamIDs {
		evaluateeTeamID := int(adviserTeamIDs[i])
		evaluatorTeamIDs := append(append([]int64{}, adviserTeamIDs[:i]...), adviserTeamIDs[i+1:]...)
		for j := range evaluatorTeamIDs {
			evaluatorTeamID := int(evaluatorTeamIDs[j])
			if evaluateeEvaluators[evaluateeTeamID] == nil {
				evaluateeEvaluators[evaluateeTeamID] = make(map[int]bool)
			}
			if !evaluateeEvaluators[evaluateeTeamID][evaluatorTeamID] {
				// Uncheck this pair of evaluatee and evaluator
				evaluateeEvaluators[evaluateeTeamID][evaluatorTeamID] = false
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
	data := map[string]interface{}{
		"EvaluateeEvaluators": evaluateeEvaluators,
		"Teams":               teams,
	}
	adv.skylb.Wender(w, r, data, "app/advisers/evaluatee_evaluators.html")
}

func (adv Advisers) EvaluateeEvaluatorsUpdate(w http.ResponseWriter, r *http.Request) {
	adv.skylb.Log.TraceRequest(r)
	user, _ := r.Context().Value(skylab.ContextUser).(skylab.User)
	userRoleID := user.Roles[skylab.RoleAdviser]
	if userRoleID == 0 {
		adv.skylb.NotAuthorized(w, r)
		return
	}

	// get list of teamIDs under adviser
	adviserTeamIDs, err := adv.getTeamIDs(user)
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
		evaluateeTeamID, err := strconv.Atoi(key)
		if err != nil || !validTeamID[evaluateeTeamID] {
			continue
		}
		if pairings[evaluateeTeamID] == nil {
			pairings[evaluateeTeamID] = make(map[int]bool)
		}
		for _, value := range values {
			evaluatorTeamID, err := strconv.Atoi(value)
			if err != nil || !validTeamID[evaluatorTeamID] {
				continue
			}
			pairings[evaluateeTeamID][evaluatorTeamID] = true
		}
	}

	var fields sq.Fields
	var values sq.RowValues
	for i := range adviserTeamIDs {
		evaluateeTeamID := adviserTeamIDs[i]
		evaluatorTeamIDs := append(append([]int{}, adviserTeamIDs[:i]...), adviserTeamIDs[i+1:]...)
		for j := range evaluatorTeamIDs {
			evaluatorTeamID := evaluatorTeamIDs[j]
			if pairings[evaluateeTeamID][evaluatorTeamID] {
				// pending insertion
				values = append(values, []interface{}{evaluateeTeamID, evaluatorTeamID})
			} else {
				// pending deletion
				fields = append(fields, sq.Fieldf("(?, ?)", evaluateeTeamID, evaluatorTeamID))
			}
		}
	}
	tp := tables.TEAM_EVALUATION_PAIRS()
	_, err = sq.InsertQuery{
		Log:           adv.skylb.Log,
		LogFlag:       sq.Lstats,
		IntoTable:     tp,
		InsertColumns: []sq.Field{tp.EVALUATEE_TEAM_ID, tp.EVALUATOR_TEAM_ID},
		RowValues:     values,
	}.OnConflict().DoNothing().Exec(adv.skylb.DB, sq.ErowsAffected)
	if err != nil {
		adv.skylb.InternalServerError(w, r, err)
		return
	}
	_, err = sq.WithDefaultLog(sq.Lstats).
		DeleteFrom(tp).
		Where(sq.RowValue{tp.EVALUATEE_TEAM_ID, tp.EVALUATOR_TEAM_ID}.In(fields)).
		Exec(adv.skylb.DB, sq.ErowsAffected)
	if err != nil {
		adv.skylb.InternalServerError(w, r, err)
		return
	}
	r, _ = adv.skylb.SetFlashMsgs(w, r, map[string][]string{flash.Success: {"pairings updated!"}})
	http.Redirect(w, r, skylab.AdviserEvaluateeEvaluators, http.StatusMovedPermanently)
}

func (adv Advisers) getTidsV2(userRoleID int) (tids []int, err error) {
	query := `SELECT tid FROM teams WHERE adviser = $1 ORDER BY tid`
	rows, err := adv.skylb.DB.Queryx(query, userRoleID)
	if err != nil {
		return tids, erro.Wrap(err)
	}
	defer rows.Close()
	for rows.Next() {
		var tid int
		err = rows.Scan(&tid)
		if err != nil {
			return tids, erro.Wrap(err)
		}
		tids = append(tids, tid)
	}
	return tids, nil
}

type Team struct {
	Tid             int            `db:"tid"`
	TeamName        sql.NullString `db:"team_name"`
	ProjectLevel    string         `db:"project_level"`
	Stu1Displayname sql.NullString `db:"stu1_displayname"`
	Stu2Displayname sql.NullString `db:"stu2_displayname"`
}

func (adv Advisers) getTeams(userRoleID int) (teams map[int]Team, err error) {
	teams = make(map[int]Team)
	query := `
	SELECT
		tid
		,team_name
		,project_level
		,stu1_displayname
		,stu2_displayname
	FROM
		app.v_teams_and_students
	WHERE
		adviser = $1
	`
	rows, err := adv.skylb.DB.Queryx(query, userRoleID)
	if err != nil {
		return teams, erro.Wrap(err)
	}
	defer rows.Close()
	for rows.Next() {
		var team Team
		err = rows.StructScan(&team)
		if err != nil {
			return teams, erro.Wrap(err)
		}
		teams[team.Tid] = team
	}
	return teams, err
}
