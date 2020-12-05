package skylab

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	sq "github.com/bokwoon95/go-structured-query/postgres"
	"github.com/bokwoon95/nusskylabx/helpers/auth"
	"github.com/bokwoon95/nusskylabx/helpers/cookies"
	"github.com/bokwoon95/nusskylabx/helpers/erro"
	"github.com/bokwoon95/nusskylabx/tables"
	"github.com/lib/pq"
)

const (
	SessionCookieName      = "_skylab_session"       // The name of the cookie that stores the user's session
	AdminSessionCookieName = "_skylab_session_admin" // The name of the cookie that stores the admin's session
)

// EnsureIsUser ensures that the email from auth.Authenticate is a valid user
// in the system
func (skylb Skylab) EnsureIsUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		skylb.Log.TraceRequest(r)
		email, ok := r.Context().Value("email").(string)
		if !ok {
			skylb.BadRequest(w, r, "email not found from context")
			return
		}
		query := `SELECT EXISTS( SELECT 1 FROM users WHERE email = $1 )`
		var exists bool
		err := skylb.DB.QueryRowx(query, email).Scan(&exists)
		if err != nil {
			skylb.InternalServerError(w, r, err)
			return
		}
		if !exists {
			skylb.NotAUser(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// SetSession sets the session for the user
func (skylb Skylab) SetSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		skylb.Log.TraceRequest(r)
		displayname, ok2 := r.Context().Value("displayname").(string)
		email, ok3 := r.Context().Value("email").(string)
		user := User{Displayname: displayname, Email: email}
		if !ok2 || !ok3 {
			skylb.BadRequest(w, r, fmt.Sprintf("Incomplete user retrieved from context: %+v", user))
			return
		}
		user.Valid = true
		sessionID, err := auth.GenerateRandomString()
		if err != nil {
			skylb.InternalServerError(w, r, err)
			return
		}
		sessionHash := skylb.Hash([]byte(sessionID))
		query := "SELECT app.set_session($1, $2)"
		_, err = skylb.DB.Exec(query, sessionHash, user.Email)
		if err != nil {
			if pqerr, ok := erro.AsPqError(err); ok {
				switch pqerr.Code {
				case ErrUserNotExist.PqCode():
					skylb.NotLoggedIn(w, r)
					return
				}
			}
			skylb.InternalServerError(w, r, err)
			return
		}
		cookies.SetCookie(w, SessionCookieName, sessionID)
		ctx := r.Context()
		ctx = context.WithValue(ctx, ContextUser, user)
		isAdmin, err := skylb.SessionIdIsValidRole(sessionID, RoleAdmin)
		if err != nil {
			skylb.InternalServerError(w, r.WithContext(ctx), err)
			return
		}
		if isAdmin {
			cookies.SetCookie(w, AdminSessionCookieName, sessionID)
			ctx = context.WithValue(ctx, ContextAdmin, user)
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserFromCookie gets a User from the database using a cookie's session ID
func (skylb Skylab) GetUserFromCookie(r *http.Request, cookieName string) (user User, err error) {
	cookie, _ := r.Cookie(cookieName)
	if cookie != nil {
		sessionID := cookie.Value
		user, err = skylb.GetUserFromSessionID(sessionID)
	}
	return user, erro.Wrap(err)
}

// GetSession gets a User and an Admin from the database using their
// corresponding cookie session IDs, and injects them into the current context
func (skylb Skylab) GetSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// skylb.Log.StartRequest(r)
		// skylb.Log.TraceRequest(r)
		user, err := skylb.GetUserFromCookie(r, SessionCookieName)
		if err != nil {
			skylb.InternalServerError(w, r, err)
			return
		}
		admin, err := skylb.GetUserFromCookie(r, AdminSessionCookieName)
		if err != nil {
			skylb.InternalServerError(w, r, err)
			return
		}
		admin.Valid = admin.Roles[RoleAdmin] != 0 // Ensure that admin is valid only if it is a RoleAdmin
		// skylb.Log.RequestPrintf(r, "user: %+v", user)
		// skylb.Log.RequestPrintf(r, "admin: %+v", admin)
		r = r.WithContext(context.WithValue(r.Context(), ContextUser, user))
		r = r.WithContext(context.WithValue(r.Context(), ContextAdmin, admin))
		next.ServeHTTP(w, r)
	})
}

// GetSession gets a User and an Admin from the database using their
// corresponding cookie session IDs, and injects them into the current context.
// If the User or Admin does not have the required roles, the user will
// be redirected to "app/skylab/403.html" instead
//
// Calling EnsureRole(RoleNull) is equivalent to calling
// GetSession() directly
func (skylb Skylab) EnsureRole(role string) func(http.Handler) http.Handler {
	if !Contains(Roles(), role) {
		panic("invalid role: " + role)
	}
	return func(next http.Handler) http.Handler {
		fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch role {
			case RoleAdmin:
				admin, _ := r.Context().Value(ContextAdmin).(User)
				if !admin.Valid || admin.Roles[RoleAdmin] == 0 {
					skylb.Log.Printf("admin is not a valid admin %+v", admin)
					skylb.NotAnAdmin(w, r)
					return
				}
			default:
				user, _ := r.Context().Value(ContextUser).(User)
				if role != RoleNull && (!user.Valid || user.Roles[role] == 0) {
					skylb.Log.Printf("user is not a valid %s %+v", role, user)
					skylb.NotARole(role)(w, r)
					return
				}
			}
			next.ServeHTTP(w, r)
		})
		return skylb.GetSession(skylb.HasValidRole(fn))
	}
}

// RevokeSessionCookie will revoke a cookie's session ID from the database,
// followed by deleting the cookie
func (skylb Skylab) RevokeSessionCookie(w http.ResponseWriter, r *http.Request, cookieName string) (err error) {
	cookie, _ := r.Cookie(cookieName)
	if cookie != nil {
		sessionID := cookie.Value
		sessionHash := skylb.Hash([]byte(sessionID))
		query := "DELETE FROM sessions WHERE hash = $1"
		_, err = skylb.DB.Exec(query, sessionHash)
		if err != nil {
			return erro.Wrap(err)
		}
		cookies.DeleteCookie(w, cookieName)
	}
	return nil
}

// RevokeSession will revoke the user's or admin's (or both) sessions depending
// on whether the 'user' and 'admin' query params were provided. If neither was
// provided, both sessions will be revoked.
//
// If only the user's session is revoked and the admin's session is not, the
// admin's existing session ID will be copied over as the new user session
// cookie. This ensures that the user session cookie is always present if
// someone is logged in.
func (skylb Skylab) RevokeSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		skylb.Log.TraceRequest(r)
		userLogout := r.FormValue("user") != ""
		adminLogout := r.FormValue("admin") != ""
		if !userLogout && !adminLogout {
			userLogout = true
			adminLogout = true
		}
		skylb.Log.Printf("userLogout: %t, adminLogout: %t", userLogout, adminLogout)
		if userLogout {
			err := skylb.RevokeSessionCookie(w, r, SessionCookieName)
			if err != nil {
				skylb.InternalServerError(w, r, err)
				return
			}
		}
		if adminLogout {
			err := skylb.RevokeSessionCookie(w, r, AdminSessionCookieName)
			if err != nil {
				skylb.InternalServerError(w, r, err)
				return
			}
			cookies.DeleteCookie(w, LastRoleCookieName)
		}
		if userLogout && !adminLogout {
			adminSessionID := cookies.GetCookieValue(r, AdminSessionCookieName)
			isAdmin, err := skylb.SessionIdIsValidRole(adminSessionID, RoleAdmin)
			if err != nil {
				skylb.InternalServerError(w, r, err)
				return
			}
			// If admin session cookie is valid, replace the just-deleted user
			// session cookie with the admin session cookie. Also set the
			// LastRoleCookie to admin.
			if isAdmin {
				skylb.Log.Printf("admin still logged in, making admin the current user")
				cookies.SetCookie(w, SessionCookieName, adminSessionID)
				cookies.SetCookie(w, LastRoleCookieName, RoleAdmin)
			}
		}
		next.ServeHTTP(w, r)
	})
}

func (skylb Skylab) RedirectAfterLogout(w http.ResponseWriter, r *http.Request) {
	skylb.Log.TraceRequest(r)
	redirectHome := skylb.Redirect("/")
	admin, _ := skylb.GetUserFromCookie(r, AdminSessionCookieName)
	admin.Valid = admin.Roles[RoleAdmin] != 0
	if admin.Valid {
		skylb.Log.Printf("admin still logged in, redirecting to last admin section")
		skylb.RedirectToLastSection(RoleAdmin)(redirectHome).ServeHTTP(w, r)
		return
	}
	redirectHome(w, r)
}

// SessionIdIsValidRole checks if the given sessionID (hashed into a
// sessionHash) exists in the database together with the given role
func (skylb Skylab) SessionIdIsValidRole(sessionID string, role string) (valid bool, err error) {
	sessionHash := skylb.Hash([]byte(sessionID))
	ss, ur := tables.SESSIONS(), tables.USER_ROLES()
	rowsAffected, err := sq.WithDefaultLog(sq.Lverbose).
		From(ss).
		Join(ur, ur.USER_ID.Eq(ss.USER_ID)).
		Where(
			ss.HASH.EqString(sessionHash),
			ur.ROLE.EqString(role),
		).
		SelectOne().
		Exec(skylb.DB, sq.ErowsAffected)
	valid = rowsAffected != 0
	return valid, erro.Wrap(err)
}

// SetSessionForUserID sets the session for the given userID and returns the
// sessionID as well as the sessionHash. Prefer using SetSession over
// SetSessionForUserID if possible, as userID and sessionHash are considered
// more low level implementation details that are only needed in specific
// situations
func (skylb Skylab) SetSessionForUserID(userID int) (sessionID string, sessionHash string, err error) {
	sessionID, err = auth.GenerateRandomString()
	if err != nil {
		return sessionID, sessionHash, erro.Wrap(err)
	}
	sessionHash = skylb.Hash([]byte(sessionID))
	query := "INSERT INTO sessions (hash, user_id) VALUES ($1, $2)"
	_, err = skylb.DB.Exec(query, sessionHash, userID)
	var e *pq.Error
	if errors.As(err, &e) {
		switch e.Code {
		case erro.PqForeignKeyViolation:
			return sessionID, sessionHash, erro.Wrap(fmt.Errorf(
				"Tried to set a session for a nonexistent user userID[%d]: %w", userID, err,
			))
		}
	}
	return sessionID, sessionHash, erro.Wrap(err)
}

// GetUserFromSessionID retrieves a User by sessionID.
//
// Even if the user cannot be found in the database, this function will return
// without error. It is up to you to check the if the returned user's 'Valid'
// field is set to true. This makes it easier to distinguish between whether a
// user is valid (the user's Valid field is false) or whether an actual error
// occurred while querying the database (the returned error is non-nil).
//
// Alternatively you can also check if the returned user has a particular role
// that is required to view the resource.
func (skylb Skylab) GetUserFromSessionID(sessionID string) (user User, err error) {
	sessionHash := skylb.Hash([]byte(sessionID))
	// Get the user
	u, ss := tables.USERS(), tables.SESSIONS()
	err = sq.
		From(ss).
		Join(u, u.USER_ID.Eq(ss.USER_ID)).
		Where(ss.HASH.EqString(sessionHash)).
		SelectRowx(func(row *sq.Row) {
			user.Valid = row.IntValid(u.USER_ID)
			user.UserID = row.Int(u.USER_ID)
			user.Displayname = row.String(u.DISPLAYNAME)
			user.Email = row.String(u.EMAIL)
		}).
		Fetch(skylb.DB)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return user, nil
		}
		return user, erro.Wrap(err)
	}
	// Get the user roles
	user.Roles = make(map[string]int)
	ur := tables.USER_ROLES()
	var userRoleID int
	var role string
	err = sq.From(ur).Where(ur.USER_ID.EqInt(user.UserID)).Selectx(func(row *sq.Row) {
		userRoleID = row.Int(ur.USER_ROLE_ID)
		role = row.String(ur.ROLE)
	}, func() {
		user.Roles[role] = userRoleID
	}).Fetch(skylb.DB)
	if err != nil {
		return user, erro.Wrap(err)
	}
	return user, nil
}
