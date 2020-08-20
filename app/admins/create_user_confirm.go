package admins

import (
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/bokwoon95/nusskylabx/app/skylab"

	"github.com/bokwoon95/nusskylabx/helpers/auth"
	"github.com/bokwoon95/nusskylabx/helpers/erro"
	"github.com/bokwoon95/nusskylabx/helpers/flash"
	"github.com/bokwoon95/nusskylabx/helpers/formutil"
	"github.com/bokwoon95/nusskylabx/helpers/headers"
	"github.com/bokwoon95/nusskylabx/helpers/templateutil"
)

// Represents all the available actions one can take in the
// create_user_confirm.html page
type createUserAction int

const (
	actionCreateUser createUserAction = 1 << iota
	actionCreateRole
	actionUpdateDisplayname
	actionDoNothing
	actionBadEntry // Used if user enters invalid data i.e. client side error
	actionError    // Used if an unknown error occurs i.e. server side error
)

func addCreateUserActions(funcs template.FuncMap) template.FuncMap {
	if funcs == nil {
		funcs = template.FuncMap{}
	}
	funcs["actionCreateUser"] = func() createUserAction { return actionCreateUser }
	funcs["actionCreateRole"] = func() createUserAction { return actionCreateRole }
	funcs["actionUpdateDisplayname"] = func() createUserAction { return actionUpdateDisplayname }
	funcs["actionDoNothing"] = func() createUserAction { return actionDoNothing }
	funcs["actionBadEntry"] = func() createUserAction { return actionBadEntry }
	funcs["actionError"] = func() createUserAction { return actionError }
	return funcs
}

// UserPendingCreation is a struct that contains the details of a user pending
// creation, as well as the actions that should be done on him (denoted by the
// Action field).
type UserPendingCreation struct {
	Cohort         string
	Role           string
	Displayname    string
	OldDisplayname string
	Email          string

	Action          createUserAction
	BadEntryDetails string
	ErrStr          string
}

func (user UserPendingCreation) String() string {
	return fmt.Sprintf("%s,%s,%s,%s", user.Cohort, user.Role, user.Displayname, user.Email)
}

func (adm Admins) CreateUserConfirm(w http.ResponseWriter, r *http.Request) {
	adm.skylb.Log.TraceRequest(r)
	r = adm.skylb.SetRoleSection(w, r, skylab.RoleAdmin, skylab.AdminCreateUser)
	headers.DoNotCache(w)
	type Data struct {
		Users []UserPendingCreation

		// Group the UsersPendingCreation by the action(s) that should be taken
		SortedUsers map[createUserAction][]UserPendingCreation
	}
	data := Data{
		SortedUsers: make(map[createUserAction][]UserPendingCreation),
	}
	csv := r.FormValue("csv") // Get the csv string from the front end
	rows := strings.Split(strings.Replace(csv, "\r\n", "\n", -1), "\n")
	rows = removeEmptyStrings(rows)
	for _, row := range rows {
		var user UserPendingCreation
		var columns []string
		if strings.Contains(row, "\t") {
			columns = strings.Split(row, "\t")
		} else {
			columns = strings.Split(row, ",")
		}
		// Deconstruct each csv row into its corresponding meaning
		for i, column := range columns {
			column = strings.TrimSpace(column)
			switch i {
			case 0:
				user.Cohort = column
			case 1:
				user.Role = column
			case 2:
				user.Displayname = column
			case 3:
				user.Email = column
			default:
				break
			}
		}
		// For the rest of this loop, we will decide what to do with the
		// UserPendingCreation for this row i.e. should we create a user and
		// role? Should we just create a role because the user already exists?
		// Is it a bad entry e.g. the role is left blank?
		if user.Cohort == "" {
			user.Cohort = adm.skylb.CurrentCohort()
		}
		if !skylab.Contains(adm.skylb.Cohorts(), user.Cohort) {
			user.Action = actionBadEntry
			user.BadEntryDetails = fmt.Sprintf(
				"Invalid cohort: %s<br>Available cohorts: %s",
				user.Cohort, strings.Join(adm.skylb.Cohorts(), ", "),
			)
			data.Users = append(data.Users, user)
			continue
		}
		if user.Role == "" {
			user.Action = actionBadEntry
			user.BadEntryDetails = fmt.Sprintf(
				"Role cannot be blank<br>Available roles: %s",
				strings.Join(skylab.Roles(), ", "),
			)
			data.Users = append(data.Users, user)
			continue
		}
		if !skylab.Contains(skylab.Roles(), user.Role) {
			user.Action = actionBadEntry
			user.BadEntryDetails = fmt.Sprintf(
				"Invalid role: %s<br>Available roles: %s",
				user.Role, strings.Join(skylab.Roles(), ", "),
			)
			data.Users = append(data.Users, user)
			continue
		}
		if user.Email == "" {
			user.Action = actionBadEntry
			user.BadEntryDetails = "Email cannot be blank"
			data.Users = append(data.Users, user)
			continue
		}
		var userID int
		var displayname string
		query := `SELECT user_id, displayname FROM users WHERE email = $1`
		err := adm.skylb.DB.QueryRowx(query, user.Email).Scan(&userID, &displayname)
		if err != nil {
			switch {
			case errors.Is(err, sql.ErrNoRows):
				// If user doesn't exist, create both user and role
				user.Action = actionCreateUser | actionCreateRole
			default:
				log.Println(err.Error())
				user.Action = actionError
				user.ErrStr = err.Error()
			}
			data.Users = append(data.Users, user)
			continue
		}
		if user.Displayname == "" && displayname != "" {
			// If user.Displayname is left blank and user has displayname, we
			// should reuse the user's current displayname
			user.Displayname = displayname
		}
		var exists bool
		query = `SELECT EXISTS( SELECT 1 FROM user_roles WHERE user_id = $1 AND role = $2 )`
		err = adm.skylb.DB.QueryRowx(query, userID, user.Role).Scan(&exists)
		if err != nil {
			log.Println(err.Error())
			user.Action = actionError
			user.ErrStr = err.Error()
			data.Users = append(data.Users, user)
			continue
		}
		if displayname == user.Displayname && exists {
			user.Action = actionDoNothing
			data.Users = append(data.Users, user)
			continue
		}
		if displayname != user.Displayname {
			user.Action = user.Action | actionUpdateDisplayname
			user.OldDisplayname = displayname
		}
		if !exists {
			user.Action = user.Action | actionCreateRole
		}
		data.Users = append(data.Users, user)
	}
	upcExists := make(map[UserPendingCreation]bool)
	for _, user := range data.Users {
		if !upcExists[user] {
			upcExists[user] = true
			data.SortedUsers[user.Action] = append(data.SortedUsers[user.Action], user)
		}
	}
	funcs := template.FuncMap{}
	funcs = addCreateUserActions(funcs)
	funcs = templateutil.Funcs(funcs)
	funcs["bitwiseOr"] = bitwiseOr
	funcs["hasBits"] = hasBits
	funcs["serialize"] = func(user UserPendingCreation) (output string, err error) {
		return auth.Serialize(adm.skylb.SecretKey, user)
	}
	funcs["deserialize"] = func(input string) (user UserPendingCreation, err error) {
		err = auth.Deserialize(adm.skylb.SecretKey, input, &user)
		return user, err
	}
	adm.skylb.Render(w, r, data, funcs, "app/admins/create_user_confirm.html")
}

func bitwiseOr(actions ...createUserAction) createUserAction {
	var a createUserAction
	for _, action := range actions {
		a = a | action
	}
	return a
}

// hasBits checks it the bit pattern b contains any of the given flags
func hasBits(b createUserAction, flag createUserAction, flags ...createUserAction) bool {
	if b&flag != 0 {
		return true
	}
	for _, f := range flags {
		if b&f != 0 {
			return true
		}
	}
	return false
}

const createUserRowsCookie = "_skylab_create_user_rows"

func (adm Admins) CreateUserConfirmPost(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		adm.skylb.Log.TraceRequest(r)
		_ = formutil.ParseForm(r)
		msgs := make(map[string][]string)
		var rows []string
		var nothingDone int  // keep track of number of users with nothing to do
		var usersCreated int // keep track of number of users created
		var usersUpdated int // keep track of number of users updated
		for _, values := range r.Form {
			var user UserPendingCreation
			var upcValid bool      // Represents whether the value was deserialized successfully into user without error
			var confirmAction bool // Represents whether the admin ticked the checkbox for that action
			for _, value := range values {
				if value == "checked" {
					// Every UserPendingCreation displayed to the admin has a
					// corresponding checkbox input with the value "checked".
					// If we see this "checked", it means the admin has checked
					// the confirmation checkbox and we should proceed with
					// whatever Action is in the UserPendingCreation.
					//
					// It also means if we never see this "checked" but we see
					// a valid UserPendingCreation, it means the admin is
					// choosing not to continue with the Action and we should
					// return the UserPendingCreation back to the textarea for
					// the admin to edit
					confirmAction = true
					continue
				}
				err := auth.Deserialize(adm.skylb.SecretKey, value, &user)
				if err != nil {
					switch {
					case errors.Is(err, auth.ErrDeserializeInputInvalid):
						// Do nothing
					default:
						msgs[flash.Error] = append(msgs[flash.Error], err.Error())
					}
				} else {
					upcValid = true
					continue
				}
			}
			if upcValid {
				notifyAdminOfError := func(row string, err error) {
					// notifyAdminOfError is called in the event of an unexpected
					// error. We want to add the error as a flash message to
					// display to the admin, and return the row back to the textarea
					// for the admin to inspect and edit
					msgs[flash.Error] = append(msgs[flash.Error], fmt.Sprintf("%s<br>%s", row, erro.Wrap(err)))
					rows = append(rows, row)
				}
				row := fmt.Sprintf("%s, %s, %s, %s", user.Cohort, user.Role, user.Displayname, user.Email)
				if hasBits(user.Action, actionCreateUser, actionCreateRole) && !confirmAction {
					// If the action involves either creating a user or
					// creating a role but the admin did not confirm the
					// action, we don't proceed with the action and will just
					// return the current row to the textarea for editing.
					rows = append(rows, row)
					continue
				}
				// Carry out the necessary actions depending on what user.Action
				// is, and update the usersCreated/usersUpdated/nothingDone
				// count accordingly
				switch user.Action {
				case actionCreateUser | actionCreateRole, actionCreateRole:
					// Since DB.CreateUser is idempotent, we can use it for
					// both creating the user role only (actionCreateRole) and
					// creating both the user and role (actionCreateUser |
					// actionCreateRole)
					_, err := adm.d.CreateUser(skylab.User{
						Displayname: user.Displayname,
						Email:       user.Email,
						Roles:       map[string]int{user.Role: 0},
					}, user.Cohort)
					if err != nil {
						notifyAdminOfError(row, err)
						continue
					}
					usersCreated++
				case actionUpdateDisplayname:
					query := `UPDATE users SET displayname = $1 WHERE email = $2`
					_, err := adm.skylb.DB.Exec(query, user.Displayname, user.Email)
					if err != nil {
						notifyAdminOfError(row, err)
						continue
					}
					usersUpdated++
				case actionUpdateDisplayname | actionCreateRole:
					query := `UPDATE users SET displayname = $1 WHERE email = $2`
					_, err := adm.skylb.DB.Exec(query, user.Displayname, user.Email)
					if err != nil {
						notifyAdminOfError(row, err)
						continue
					}
					_, err = adm.d.CreateUser(skylab.User{
						Displayname: user.Displayname,
						Email:       user.Email,
						Roles:       map[string]int{user.Role: 0},
					}, user.Cohort)
					if err != nil {
						notifyAdminOfError(row, err)
						continue
					}
					usersUpdated++
				case actionDoNothing:
					nothingDone++
				case actionBadEntry:
					msgs[flash.Error] = append(msgs[flash.Error], fmt.Sprintf("%s<br>%s", row, user.BadEntryDetails))
					rows = append(rows, row)
				case actionError:
					msgs[flash.Error] = append(msgs[flash.Error], fmt.Sprintf("%s<br>%s", row, user.ErrStr))
					rows = append(rows, row)
				}
			}
		}
		if usersCreated > 0 {
			msgs[flash.Success] = append(msgs[flash.Success], fmt.Sprintf("%d users created", usersCreated))
		}
		if usersUpdated > 0 {
			msgs[flash.Success] = append(msgs[flash.Success], fmt.Sprintf("%d users updated", usersUpdated))
		}
		if nothingDone > 0 {
			msgs[flash.Warning] = append(msgs[flash.Warning], fmt.Sprintf("Notice: %d user(s) left unchanged", nothingDone))
		}
		r, _ = adm.skylb.SetFlashMsgs(w, r, msgs)
		_ = adm.skylb.EncodeVariableInCookie(w, createUserRowsCookie, rows)
		next.ServeHTTP(w, r)
	})
}
