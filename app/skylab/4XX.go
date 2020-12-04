package skylab

import (
	"fmt"
	"net/http"
	"runtime"

	"github.com/bokwoon95/nusskylabx/helpers/formutil"
	"github.com/bokwoon95/nusskylabx/helpers/logutil"
	"github.com/bokwoon95/nusskylabx/helpers/templateutil"
)

func (skylb Skylab) BadRequest(w http.ResponseWriter, r *http.Request, msg string) {
	skylb.Log.TraceRequest(r)
	pc, filename, linenr, _ := runtime.Caller(1)
	data := map[string]interface{}{
		"Msg":      msg,
		"Location": fmt.Sprintf("%s:%s:%d", runtime.FuncForPC(pc).Name(), filename, linenr),
	}
	w.WriteHeader(http.StatusBadRequest)
	skylb.Wender(w, r, data, "app/skylab/400.html")
}

func (skylb Skylab) NotLoggedIn(w http.ResponseWriter, r *http.Request) {
	skylb.Log.TraceRequest(r)
	w.WriteHeader(http.StatusUnauthorized)
	skylb.Wender(w, r, nil, "app/skylab/401.html")
}

func (skylb Skylab) NotAUser(w http.ResponseWriter, r *http.Request) {
	skylb.Log.TraceRequest(r)
	w.WriteHeader(http.StatusUnauthorized)
	skylb.Wender(w, r, nil, "app/skylab/401_not_a_user.html")
}

// Authentication is not Authorization. Not authenticated means the user
// is not logged in. Not authorized means user is logged in but not allowed
// to carry out the action e.g. student trying to access an admin page
//
// 401 Unauthorized == Authentication error;
// 403 Forbidden == Authorization error
//
// It seems contradictory that an *Authorization* error is actually 403 Forbidden
// and not 401 Unauthorized, but it's true. The people who designed the spec
// made a mistake and now we're all suffering for it.
// https://stackoverflow.com/a/6937030
func (skylb Skylab) NotAuthorized(w http.ResponseWriter, r *http.Request) {
	skylb.Log.TraceRequest(r)
	data := map[string]interface{}{
		"Role": RoleNull,
	}
	w.WriteHeader(http.StatusForbidden)
	skylb.Wender(w, r, data, "app/skylab/403.html")
}

func (skylb Skylab) NotARole(role string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		skylb.Log.TraceRequest(r)
		data := map[string]interface{}{
			"Role": role,
		}
		w.WriteHeader(http.StatusForbidden)
		skylb.Wender(w, r, data, "app/skylab/403.html")
	}
}

func (skylb Skylab) NotAnApplicant(w http.ResponseWriter, r *http.Request) {
	skylb.Log.TraceRequest(r)
	data := map[string]interface{}{
		"Role": RoleApplicant,
	}
	w.WriteHeader(http.StatusForbidden)
	skylb.Wender(w, r, data, "app/skylab/403.html")
}

func (skylb Skylab) NotAStudent(w http.ResponseWriter, r *http.Request) {
	skylb.Log.TraceRequest(r)
	data := map[string]interface{}{
		"Role": RoleStudent,
	}
	w.WriteHeader(http.StatusForbidden)
	skylb.Wender(w, r, data, "app/skylab/403.html")
}

func (skylb Skylab) NotAnAdviser(w http.ResponseWriter, r *http.Request) {
	skylb.Log.TraceRequest(r)
	data := map[string]interface{}{
		"Role": RoleAdviser,
	}
	w.WriteHeader(http.StatusForbidden)
	skylb.Wender(w, r, data, "app/skylab/403.html")
}

func (skylb Skylab) NotAMentor(w http.ResponseWriter, r *http.Request) {
	skylb.Log.TraceRequest(r)
	data := map[string]interface{}{
		"Role": RoleMentor,
	}
	w.WriteHeader(http.StatusForbidden)
	skylb.Wender(w, r, data, "app/skylab/403.html")
}

func (skylb Skylab) NotAnAdmin(w http.ResponseWriter, r *http.Request) {
	skylb.Log.TraceRequest(r)
	data := map[string]interface{}{
		"Role": RoleAdmin,
	}
	w.WriteHeader(http.StatusForbidden)
	skylb.Wender(w, r, data, "app/skylab/403.html")
}

func (skylb Skylab) CsrfTokenInvalid() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		skylb.Log.TraceRequest(r)
		r = skylb.SetRoleSection(w, r, RolePreserve, SectionPreserve)
		_ = formutil.ParseForm(r)
		data := map[string]interface{}{
			"RequestID": logutil.GetReqID(r.Context()),
			"URL":       r.URL,
			"FormStr":   fmt.Sprintf("%#v\n", r.Form),
		}
		w.WriteHeader(http.StatusForbidden)
		skylb.Wender(w, r, data, "app/skylab/403_csrf.html")
	})
}

func (skylb Skylab) NotFound(w http.ResponseWriter, r *http.Request) {
	skylb.Log.TraceRequest(r)
	data := map[string]interface{}{
		"URL": r.URL.String(),
	}
	w.WriteHeader(http.StatusNotFound)
	skylb.Wender(w, r, data, "app/skylab/404.html")
}

func (skylb Skylab) MethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	skylb.Log.TraceRequest(r)
	data := map[string]interface{}{
		"URL":    r.URL.String(),
		"Method": r.Method,
	}
	w.WriteHeader(http.StatusMethodNotAllowed)
	skylb.Wender(w, r, data, "app/skylab/405.html")
}
