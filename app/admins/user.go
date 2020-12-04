package admins

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	sq "github.com/bokwoon95/go-structured-query/postgres"
	"github.com/bokwoon95/nusskylabx/app/skylab"
	"github.com/bokwoon95/nusskylabx/helpers/cookies"
	"github.com/bokwoon95/nusskylabx/helpers/flash"
	"github.com/bokwoon95/nusskylabx/helpers/headers"
	"github.com/bokwoon95/nusskylabx/helpers/urlparams"
	"github.com/bokwoon95/nusskylabx/tables"
)

func (adm Admins) UserView(w http.ResponseWriter, r *http.Request) {
	adm.skylb.Log.TraceRequest(r)
	r = adm.skylb.SetRoleSection(w, r, skylab.RolePreserve, skylab.SectionPreserve)
	headers.DoNotCache(w)
	userID, err := urlparams.Int(r, "userID")
	if err != nil {
		adm.skylb.BadRequest(w, r, err.Error())
		return
	}
	// Get User
	u := tables.USERS()
	var user skylab.User
	err = sq.WithDefaultLog(sq.Lverbose).
		From(u).
		Where(u.USER_ID.EqInt(userID)).
		SelectRowx(func(row *sq.Row) {
			user.Valid = row.IntValid(u.USER_ID)
			user.UserID = row.Int(u.USER_ID)
			user.Displayname = row.String(u.DISPLAYNAME)
			user.Email = row.String(u.EMAIL)
		}).
		Fetch(adm.skylb.DB)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			adm.skylb.BadRequest(w, r, fmt.Sprintf("No such user found for userID %d", userID))
		default:
			adm.skylb.InternalServerError(w, r, err)
		}
		return
	}
	// Get Roles
	ur := tables.USER_ROLES()
	var userRoleID int
	var role string
	user.Roles = make(map[string]int)
	err = sq.WithDefaultLog(sq.Lverbose).
		From(ur).
		Where(ur.USER_ID.EqInt(userID)).
		Selectx(func(row *sq.Row) {
			userRoleID = row.Int(ur.USER_ROLE_ID)
			role = row.String(ur.ROLE)
		}, func() {
			user.Roles[role] = userRoleID
		}).
		Fetch(adm.skylb.DB)
	if err != nil {
		adm.skylb.InternalServerError(w, r, err)
		return
	}
	// Get Team
	t, urs := tables.TEAMS(), tables.USER_ROLES_STUDENTS()
	var userTeam skylab.Team
	err = sq.WithDefaultLog(sq.Lverbose).
		From(t).
		Join(urs, urs.TEAM_ID.Eq(t.TEAM_ID)).
		Where(urs.USER_ROLE_ID.EqInt(user.Roles[skylab.RoleStudent])).
		SelectRowx(func(row *sq.Row) {
			userTeam.Valid = row.IntValid(t.TEAM_ID)
			userTeam.TeamID = row.Int(t.TEAM_ID)
			userTeam.Cohort = row.String(t.COHORT)
			userTeam.TeamName = row.String(t.TEAM_NAME)
			userTeam.ProjectLevel = row.String(t.PROJECT_LEVEL)
			userTeam.Status = row.String(t.STATUS)
		}).
		Fetch(adm.skylb.DB)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			// do nothing, user might not be a student
		default:
			adm.skylb.InternalServerError(w, r, err)
			return
		}
	}
	// Get AdvisingTeams
	var team skylab.Team
	var advisingTeams []skylab.Team
	err = sq.WithDefaultLog(sq.Lverbose).
		From(t).
		Where(t.ADVISER_USER_ROLE_ID.EqInt(user.Roles[skylab.RoleAdviser])).
		Selectx(func(row *sq.Row) {
			team.Valid = row.IntValid(t.TEAM_ID)
			team.TeamID = row.Int(t.TEAM_ID)
			team.Cohort = row.String(t.COHORT)
			team.TeamName = row.String(t.TEAM_NAME)
			team.ProjectLevel = row.String(t.PROJECT_LEVEL)
			team.Status = row.String(t.STATUS)
		}, func() {
			advisingTeams = append(advisingTeams, team)
		}).
		Fetch(adm.skylb.DB)
	if err != nil {
		adm.skylb.InternalServerError(w, r, err)
		return
	}
	// Get MentoringTeams
	var mentoringTeams []skylab.Team
	err = sq.WithDefaultLog(sq.Lverbose).
		From(t).
		Where(t.MENTOR_USER_ROLE_ID.EqInt(user.Roles[skylab.RoleMentor])).
		Selectx(func(row *sq.Row) {
			team.Valid = row.IntValid(t.TEAM_ID)
			team.TeamID = row.Int(t.TEAM_ID)
			team.Cohort = row.String(t.COHORT)
			team.TeamName = row.String(t.TEAM_NAME)
			team.ProjectLevel = row.String(t.PROJECT_LEVEL)
			team.Status = row.String(t.STATUS)
		}, func() {
			mentoringTeams = append(mentoringTeams, team)
		}).
		Fetch(adm.skylb.DB)
	if err != nil {
		adm.skylb.InternalServerError(w, r, err)
		return
	}
	data := map[string]interface{}{
		"User":           user,
		"Team":           userTeam,
		"AdvisingTeams":  advisingTeams,
		"MentoringTeams": mentoringTeams,
		"TeamBaseURL":    skylab.AdminTeam,
		"UserBaseURL":    skylab.AdminUser,
	}
	adm.skylb.Wender(w, r, data, "app/skylab/user_view.html")
}

func (adm Admins) UserPreviewAs(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		adm.skylb.Log.TraceRequest(r)
		headers.DoNotCache(w)

		userID, err := urlparams.Int(r, "userID")
		if err != nil {
			adm.skylb.BadRequest(w, r, err.Error())
			return
		}
		currentUser, _ := r.Context().Value(skylab.ContextUser).(skylab.User)
		if userID == currentUser.UserID {
			http.Redirect(w, r, skylab.AdminUser+"/"+strconv.Itoa(userID), http.StatusMovedPermanently)
			return
		}
		currentAdmin, _ := r.Context().Value(skylab.ContextAdmin).(skylab.User)
		if currentAdmin.UserID != currentUser.UserID {
			err := adm.skylb.RevokeSessionCookie(w, r, skylab.SessionCookieName)
			if err != nil {
				adm.skylb.InternalServerError(w, r, err)
				return
			}
		}
		sessionID, sessionHash, err := adm.skylb.SetSessionForUserID(userID)
		if err != nil {
			adm.skylb.InternalServerError(w, r, err)
			return
		}
		newUser, err := adm.skylb.GetUserFromSessionID(sessionID)
		if err != nil {
			adm.skylb.InternalServerError(w, r, err)
			return
		}
		cookies.SetCookie(w, skylab.SessionCookieName, sessionID)
		msgs := make(map[string][]string)
		msgs[flash.Success] = []string{fmt.Sprintf(`Previewing as User
<div>UserID: %d</div>
<div>Displayname: %s</div>
<div>Email: %s</div>
<div>SessionID (Cookie): <code>%s</code></div>
<div>SessionHash (Database): <code>%s</code></div>
`, newUser.UserID, newUser.Displayname, newUser.Email, sessionID, sessionHash)}
		r, _ = adm.skylb.SetFlashMsgs(w, r, msgs)
		next.ServeHTTP(w, r)
	})
}
