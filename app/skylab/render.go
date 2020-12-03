package skylab

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/bokwoon95/nusskylabx/helpers/cookies"
	"github.com/bokwoon95/nusskylabx/helpers/erro"
	"github.com/bokwoon95/nusskylabx/helpers/flash"
	"github.com/bokwoon95/nusskylabx/helpers/headers"
	"github.com/davecgh/go-spew/spew"
	"github.com/gorilla/csrf"
	"github.com/microcosm-cc/bluemonday"
)

// Render will render one or more html templates together with the given data
// and funcs. Crucially, this function is where all the global template
// functions (prefixed by "Skylab") and templates (such as the global navbar
// template "app/skylab/navbar.html") are injected. If you want to add any
// globally available template functions or templates files, this is the place
// to do it
func (skylb Skylab) Render(w http.ResponseWriter, r *http.Request, data interface{}, funcs map[string]interface{}, filename string, filenames ...string) {
	// if requested, render JSON instead of HTML and return
	if shouldJSONify(w, r) {
		skylb.renderJSON(w, r, data)
		return
	}
	// Add global template files
	filenames = append(filenames,
		filename,
		"app/skylab/head.html",
		"app/skylab/navbar.html",
		"app/skylab/sidebar.html",
		"helpers/flash/flash.html",
	)
	// Convert all relative filepaths to absolute filepaths so that it doesn't
	// matter where tests are run from (tests will cd into whatever directory
	// the test package is in, potentially screwing up a relative file lookup)
	for i := range filenames {
		if !filepath.IsAbs(filenames[i]) {
			filenames[i] = ProjectRootDir + filenames[i]
		}
	}
	// Add global template functions
	if funcs == nil {
		funcs = map[string]interface{}{}
	}
	funcs = skylb.addConsts(funcs)
	funcs = skylb.NavbarFuncs(funcs, w, r)
	funcs = skylb.AddInputSelects(funcs)
	funcs = AddSections(funcs)
	funcs = flash.Funcs(funcs, w, r, skylb.SecretKey)
	funcs = headers.Funcs(funcs, r)
	// funcs["SkylabParentTemplateFilename"] = func() string { return filename } // needed for head.html, do not remove
	// funcs["SkylabSidebarItem"] = sidebarItem
	funcs["SkylabBaseURL"] = func() string { return skylb.BaseURLWithProtocol() }
	funcs["SkylabMilestoneName"] = MilestoneName
	funcs["SkylabMilestoneNameAbbrev"] = MilestoneNameAbbrev
	funcs["SkylabSanitizeHTML"] = SanitizeHTML(skylb.Policy)
	funcs["SkylabSGTime"] = SGTime
	// Parse template
	var t *template.Template
	var err error
	t, err = template.
		New(filepath.Base(filename)).
		Funcs(funcs).
		Option("missingkey=zero").
		ParseFiles(filenames...)
	if err != nil {
		_, sourcefile, linenr, _ := runtime.Caller(1)
		skylb.InternalServerError(w, r,
			fmt.Errorf("%s:%d tried to render %s and failed: %w", sourcefile, linenr, filename, err),
		)
		return
	}
	// Execute template into temporary buffer so that we can check for any
	// errors before writing out to the user.
	// https://blog.questionable.services/article/approximating-html-template-inheritance/#error-handling
	buf := skylb.Bufpool.Get()
	defer skylb.Bufpool.Put(buf)
	err = t.Execute(buf, data)
	if err != nil {
		_, sourcefile, linenr, _ := runtime.Caller(1)
		skylb.InternalServerError(w, r,
			fmt.Errorf("%s:%d tried to render %s and failed: %w", sourcefile, linenr, filename, err),
		)
		return
	}
	// If no error, set headers and write temporary buffer into w http.ResponseWriter
	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("X-CSRF-Token", csrf.Token(r))
	_, _ = buf.WriteTo(w)
}

func (skylb Skylab) Wender(w http.ResponseWriter, r *http.Request, data map[string]interface{}, name string, names ...string) {
	currentRole, _ := r.Context().Value(ContextCurrentRole).(string)
	currentSection, _ := r.Context().Value(ContextCurrentSection).(string)
	user, _ := r.Context().Value(ContextUser).(User)
	admin, _ := r.Context().Value(ContextAdmin).(User)
	mainData := make(map[string]interface{})
	mainData["skylab"] = map[string]interface{}{
		"ParentTemplateFilename": name,
		"CSRFToken":              csrf.TemplateField(r),
		"IsProd":                 skylb.IsProd,
		"CurrentRole":            currentRole,
		"CurrentSection":         currentSection,
		"User":                   user,
		"Admin":                  admin,
	}
	mainData["flash"] = flash.GetData(w, r, skylb.SecretKey, SanitizeHTML(skylb.Policy))
	mainData["headers"] = headers.GetData(r)
	for k, v := range data {
		mainData[k] = v
	}
	if shouldJSONify(w, r) {
		skylb.renderJSON(w, r, mainData)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("X-CSRF-Token", csrf.Token(r))
	err := skylb.Templates.Render(w, r, mainData, name, names...)
	if err != nil {
		_, sourcefile, linenr, _ := runtime.Caller(1)
		skylb.InternalServerError(w, r,
			fmt.Errorf("%s:%d tried to render %s, %+v and failed: %w", sourcefile, linenr, name, names, err),
		)
	}
}

func (skylb Skylab) getTemplates() (*template.Template, error) {
	funcs := map[string]interface{}{
		"UserIsRole":          userIsRole,
		"UserIsApplicantOnly": userIsApplicantOnly,
		"SkylabBaseURL":       skylb.BaseURLWithProtocol(),
		"MilestoneName":       MilestoneName,
		"MilestoneNameAbbrev": MilestoneNameAbbrev,
		"SanitizeHTML":        SanitizeHTML(skylb.Policy),
		"SGTime":              SGTime,
		"Map":                 Map,
	}
	funcs = skylb.addConsts(funcs)
	funcs = skylb.AddInputSelects(funcs)
	funcs = AddSections(funcs)
	t := template.New("").Funcs(funcs).Option("missingkey=zero")
	globs := []string{
		"app/skylab/head.html",
		"app/skylab/navbar.html",
		"app/skylab/sidebar.html",
		"app/*.html",
		"app/admins/*.html",
		"app/advisers/*.html",
		"app/applicants/*.html",
		"app/mentors/*.html",
		"app/students/*.html",
	}
	for _, glob := range globs {
		files, err := filepath.Glob(ProjectRootDir + glob)
		if err != nil {
			return t, erro.Wrap(err)
		}
		for _, file := range files {
			name := strings.TrimPrefix(file, ProjectRootDir)
			b, err := ioutil.ReadFile(file)
			if err != nil {
				return t, erro.Wrap(err)
			}
			t, err = t.New(name).Parse(string(b))
			if err != nil {
				return t, erro.Wrap(err)
			}
		}
	}
	// flash
	flashTemplate, err := flash.GetTemplate()
	if err != nil {
		return t, erro.Wrap(err)
	}
	t, err = t.AddParseTree(flashTemplate.Name(), flashTemplate.Tree)
	if err != nil {
		return t, erro.Wrap(err)
	}
	return t, nil
}

func Map(args ...interface{}) map[string]interface{} {
	data := make(map[string]interface{})
	for i := 0; i < len(args); i += 2 {
		if i+1 >= len(args) {
			break
		}
		key := fmt.Sprint(args[i])
		value := args[i+1]
		data[key] = value
	}
	return data
}

func SanitizeHTML(policy *bluemonday.Policy) func(string) template.HTML {
	return func(input string) template.HTML {
		input = strings.ReplaceAll(input, "\\n", "<br>")
		input = policy.Sanitize(input)
		return template.HTML(input)
	}
}

// JSONifyResponse will set the value of ContextDumpJson to true, effectively
// signalling to (*Skylab).Render that it should render its output as JSON, not
// as HTML
func JSONifyResponse(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r = r.WithContext(context.WithValue(r.Context(), ContextDumpJson, true))
		next(w, r)
	}
}

// check context, cookie and request for whether the response should be JSONified
func shouldJSONify(w http.ResponseWriter, r *http.Request) bool {
	doJSONify, _ := r.Context().Value(ContextDumpJson).(bool)
	if doJSONify {
		return true
	}
	cookie, _ := r.Cookie(string(ContextDumpJson))
	if cookie != nil {
		cookies.DeleteCookie(w, string(ContextDumpJson))
		if cookie.Value != "" {
			return true
		}
	}
	if r.FormValue("dumpjson") == "true" {
		return true
	}
	return false
}

// renderJSON will marshal the data struct into a json string and write it out
// to w http.ResponseWriter. Only works if the current user is an admin.
func (skylb Skylab) renderJSON(w http.ResponseWriter, r *http.Request, data interface{}) {
	// Turn off ContextDumpJson avoid infinite loop when calling skylb.InternalServerError
	r = r.WithContext(context.WithValue(r.Context(), ContextDumpJson, false))
	headers.DoNotCache(w)
	admin, err := skylb.GetUserFromCookie(r, AdminSessionCookieName)
	if err != nil {
		skylb.InternalServerError(w, r, err)
		return
	}
	if !admin.Valid || admin.Roles[RoleAdmin] == 0 {
		skylb.NotAnAdmin(w, r)
		return
	}
	// Marshal data into JSON. If it fails, fall back on spew.Fdump (more
	// verbose but always works)
	b, err := json.Marshal(data)
	if err != nil {
		w.Header().Set("Content-Type", "text/plain")
		spew.Fdump(w, data)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(b)
}

// addConsts adds all Orbital/Skylab related consts to FuncMap
func (skylb Skylab) addConsts(funcs map[string]interface{}) map[string]interface{} {
	if funcs == nil {
		funcs = map[string]interface{}{}
	}
	funcs = skylb.addConstCohort(funcs)
	funcs = addConstProjectLevel(funcs)
	funcs = addConstRole(funcs)
	funcs = addConstApplicationStatus(funcs)
	funcs = addConstTeamStatus(funcs)
	funcs = addConstStage(funcs)
	funcs = addConstMilestone(funcs)
	return funcs
}

func (skylb Skylab) addConstCohort(funcs map[string]interface{}) map[string]interface{} {
	// cohorts := skylb.Cohorts()
	latest := skylb.LatestCohort()
	current := skylb.CurrentCohort()
	for _, cohort := range skylb.Cohorts() {
		cohort := cohort
		funcs["Cohort"+cohort] = func() string { return cohort }
	}
	funcs["SkylabCohorts"] = func() []string { return skylb.Cohorts() }
	funcs["CohortCurrent"] = func() string { return current }
	funcs["CohortLatest"] = func() string { return latest }
	return funcs
}

// csrfToken is a Template Function that returns a HTML input element with the
// CSRF token as the input value, e.g.
// <input type="hidden" name="gorilla.csrf.Token" value="<token>">
func csrfToken(r *http.Request) func() template.HTML {
	csrfHTML := csrf.TemplateField(r)
	return func() template.HTML { return csrfHTML }
}

// MilestoneName pretty prints the name of the milestone
func MilestoneName(milestone string) string {
	switch milestone {
	case Milestone1:
		return "Milestone 1"
	case Milestone2:
		return "Milestone 2"
	case Milestone3:
		return "Milestone 3"
	default:
		return "<Invalid Milestone>"
	}
}

// MilestoneNameAbbrev
func MilestoneNameAbbrev(milestone string) string {
	switch milestone {
	case Milestone1:
		return "M1"
	case Milestone2:
		return "M2"
	case Milestone3:
		return "M3"
	default:
		return "<Invalid Milestone>"
	}
}

func SGTime(t sql.NullTime) string {
	if !t.Valid {
		return "<nil>"
	}
	singapore, err := time.LoadLocation("Asia/Singapore")
	if err != nil {
		panic(err)
	}
	t.Time = t.Time.In(singapore)
	return t.Time.Format("2006-Jan-02 15:04")
}
