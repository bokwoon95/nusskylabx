package admins

import (
	"net/http"

	sq "github.com/bokwoon95/go-structured-query/postgres"
	"github.com/bokwoon95/nusskylabx/helpers/urlparams"
	"github.com/bokwoon95/nusskylabx/tables"

	"github.com/bokwoon95/nusskylabx/app/skylab"

	"github.com/bokwoon95/nusskylabx/helpers/headers"
)

func (adm Admins) ListUsers(w http.ResponseWriter, r *http.Request) {
	adm.skylb.Log.TraceRequest(r)
	r = adm.skylb.SetRoleSection(w, r, skylab.RoleAdmin, skylab.AdminListUsers)
	headers.DoNotCache(w)

	// Get the last valid cohort
	cohort, _ := urlparams.PersistentString(w, r, "cohort", "_admin_list_users_cohort")
	if cohort == "" || !skylab.Contains(adm.skylb.Cohorts(), cohort) {
		http.Redirect(w, r, skylab.AdminListUsers+"/"+adm.skylb.CurrentCohort(), http.StatusMovedPermanently)
		return
	}

	// Get the last valid role
	role, _ := urlparams.PersistentString(w, r, "role", "_admin_list_users_role")
	if !skylab.Contains(skylab.Roles(), role) {
		http.Redirect(w, r, skylab.AdminListUsers+"/"+cohort+"/"+skylab.RoleStudent, http.StatusMovedPermanently)
		return
	}

	var user skylab.User
	var users []skylab.User
	u, ur := tables.USERS(), tables.USER_ROLES()
	err := sq.WithDefaultLog(sq.Lverbose).
		From(u).
		Join(ur, ur.USER_ID.Eq(u.USER_ID)).
		Where(
			ur.COHORT.EqString(cohort),
			ur.ROLE.EqString(role),
		).
		OrderBy(u.USER_ID).
		Selectx(func(row *sq.Row) {
			user.Valid = row.IntValid(u.USER_ID)
			user.UserID = row.Int(u.USER_ID)
			user.Displayname = row.String(u.DISPLAYNAME)
			user.Email = row.String(u.EMAIL)
			user.Roles = map[string]int{
				row.String(ur.ROLE): row.Int(ur.USER_ROLE_ID),
			}
		}, func() {
			users = append(users, user)
		}).
		Fetch(adm.skylb.DB)
	if err != nil {
		adm.skylb.InternalServerError(w, r, err)
		return
	}
	data := map[string]interface{}{
		"Users":  users,
		"Cohort": cohort,
		"Role":   role,
	}
	switch role {
	case skylab.RoleStudent:
		t := tables.V_TEAMS()
		var team skylab.Team
		var teams []skylab.Team
		var student1UserID, student2UserID int
		var userIDToTeamIndex = make(map[int]int)
		err = sq.WithDefaultLog(sq.Lverbose).
			From(t).
			Where(t.COHORT.EqString(cohort)).
			Selectx(func(row *sq.Row) {
				// Team
				team.Valid = row.IntValid(t.TEAM_ID)
				team.TeamID = row.Int(t.TEAM_ID)
				team.Cohort = row.String(t.COHORT)
				team.TeamName = row.String(t.TEAM_NAME)
				team.ProjectLevel = row.String(t.PROJECT_LEVEL)
				team.Status = row.String(t.STATUS)
				// Adviser
				team.Adviser.Valid = row.IntValid(t.ADVISER_USER_ID)
				team.Adviser.UserID = row.Int(t.ADVISER_USER_ID)
				team.Adviser.Displayname = row.String(t.ADVISER_DISPLAYNAME)
				team.Adviser.Email = row.String(t.ADVISER_EMAIL)
				team.Adviser.Roles = map[string]int{
					skylab.RoleAdviser: row.Int(t.ADVISER_USER_ROLE_ID),
				}
				// Mentor
				team.Mentor.Valid = row.IntValid(t.MENTOR_USER_ID)
				team.Mentor.UserID = row.Int(t.MENTOR_USER_ID)
				team.Mentor.Displayname = row.String(t.MENTOR_DISPLAYNAME)
				team.Mentor.Email = row.String(t.MENTOR_EMAIL)
				team.Mentor.Roles = map[string]int{
					skylab.RoleMentor: row.Int(t.MENTOR_USER_ROLE_ID),
				}
				// Student1 and Student2
				student1UserID = row.Int(t.STUDENT1_USER_ID)
				student2UserID = row.Int(t.STUDENT2_USER_ID)
			}, func() {
				teams = append(teams, team)
				userIDToTeamIndex[student1UserID] = len(teams) - 1
				userIDToTeamIndex[student2UserID] = len(teams) - 1
			}).
			Fetch(adm.skylb.DB)
		if err != nil {
			adm.skylb.InternalServerError(w, r, err)
			return
		}
		data["Teams"] = teams
		data["UserIDToTeamIndex"] = userIDToTeamIndex
		adm.skylb.Render(w, r, data, "app/admins/list_students.html")
	default:
		adm.skylb.Render(w, r, data, "app/admins/list_users.html")
	}
}
