package skylab

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"runtime"

	"github.com/bokwoon95/nusskylabx/helpers/formutil"
	"github.com/bokwoon95/nusskylabx/helpers/logutil"
	"github.com/bokwoon95/nusskylabx/helpers/templateutil"
)

func (skylb Skylab) BadRequest(w http.ResponseWriter, r *http.Request, msg string) {
	skylb.Log.TraceRequest(r)
	type Data struct {
		Msg      string
		Location string
	}
	var data Data
	pc, filename, linenr, _ := runtime.Caller(1)
	data.Msg = msg
	data.Location = fmt.Sprintf("%s:%s:%d", runtime.FuncForPC(pc).Name(), filename, linenr)
	w.WriteHeader(http.StatusBadRequest)
	skylb.Render(w, r, data, nil, "app/skylab/400.html")
}

func (skylb Skylab) NotLoggedIn(w http.ResponseWriter, r *http.Request) {
	skylb.Log.TraceRequest(r)
	w.WriteHeader(http.StatusUnauthorized)
	skylb.Render(w, r, nil, templateutil.Txt(template.FuncMap{}), "app/skylab/401.html")
}

func (skylb Skylab) NotAUser(w http.ResponseWriter, r *http.Request) {
	skylb.Log.TraceRequest(r)
	w.WriteHeader(http.StatusUnauthorized)
	skylb.Render(w, r, nil, templateutil.Txt(template.FuncMap{}), "app/skylab/401_not_a_user.html")
}

type fourOhThree struct {
	Role string
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
	data := fourOhThree{Role: RoleNull}
	w.WriteHeader(http.StatusForbidden)
	skylb.Render(w, r, data, templateutil.Txt(template.FuncMap{}), "app/skylab/403.html")
}

func (skylb Skylab) NotARole(role string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		skylb.Log.TraceRequest(r)
		data := fourOhThree{Role: role}
		w.WriteHeader(http.StatusForbidden)
		skylb.Render(w, r, data, templateutil.Txt(template.FuncMap{}), "app/skylab/403.html")
	}
}

func (skylb Skylab) NotAnApplicant(w http.ResponseWriter, r *http.Request) {
	skylb.Log.TraceRequest(r)
	data := fourOhThree{Role: RoleApplicant}
	w.WriteHeader(http.StatusForbidden)
	skylb.Render(w, r, data, templateutil.Txt(template.FuncMap{}), "app/skylab/403.html")
}

func (skylb Skylab) NotAStudent(w http.ResponseWriter, r *http.Request) {
	skylb.Log.TraceRequest(r)
	data := fourOhThree{Role: RoleStudent}
	w.WriteHeader(http.StatusForbidden)
	skylb.Render(w, r, data, templateutil.Txt(template.FuncMap{}), "app/skylab/403.html")
}

func (skylb Skylab) NotAnAdviser(w http.ResponseWriter, r *http.Request) {
	skylb.Log.TraceRequest(r)
	data := fourOhThree{Role: RoleAdviser}
	w.WriteHeader(http.StatusForbidden)
	skylb.Render(w, r, data, templateutil.Txt(template.FuncMap{}), "app/skylab/403.html")
}

func (skylb Skylab) NotAMentor(w http.ResponseWriter, r *http.Request) {
	skylb.Log.TraceRequest(r)
	data := fourOhThree{Role: RoleMentor}
	w.WriteHeader(http.StatusForbidden)
	skylb.Render(w, r, data, templateutil.Txt(template.FuncMap{}), "app/skylab/403.html")
}

func (skylb Skylab) NotAnAdmin(w http.ResponseWriter, r *http.Request) {
	skylb.Log.TraceRequest(r)
	data := fourOhThree{Role: RoleAdmin}
	w.WriteHeader(http.StatusForbidden)
	skylb.Render(w, r, data, templateutil.Txt(template.FuncMap{}), "app/skylab/403.html")
}

func (skylb Skylab) CsrfTokenInvalid() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		skylb.Log.TraceRequest(r)
		r = skylb.SetRoleSection(w, r, RolePreserve, SectionPreserve)
		_ = formutil.ParseForm(r)
		type Data struct {
			RequestID string
			URL       *url.URL
			FormStr   string
		}
		var data Data
		data.RequestID = logutil.GetReqID(r.Context())
		data.URL = r.URL
		data.FormStr = fmt.Sprintf("%#v\n", r.Form)
		w.WriteHeader(http.StatusForbidden)
		skylb.Render(w, r, data, nil, "app/skylab/403_csrf.html")
	})
}

func (skylb Skylab) NotFound(w http.ResponseWriter, r *http.Request) {
	skylb.Log.TraceRequest(r)
	type Data struct{ URL string }
	data := Data{URL: r.URL.String()}
	w.WriteHeader(http.StatusNotFound)
	skylb.Render(w, r, data, nil, "app/skylab/404.html")
}

func (skylb Skylab) MethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	skylb.Log.TraceRequest(r)
	type Data struct {
		URL    string
		Method string
	}
	data := Data{URL: r.URL.String(), Method: r.Method}
	w.WriteHeader(http.StatusMethodNotAllowed)
	skylb.Render(w, r, data, nil, "app/skylab/405.html")
}
