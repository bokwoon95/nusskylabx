package skylab

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"regexp"

	"github.com/bokwoon95/nusskylabx/helpers/erro"
	"github.com/bokwoon95/nusskylabx/helpers/logutil"
)

// InternalServerError is a catch-all handler that will direct the user to a generic
// error page indicating 500 Internal Server Error.
//
// InternalServerError must not depend on any function that calls InternalServerError on error
// (such as (*Skylab).Render), otherwise it will spin into an infinite loop
func (skylb Skylab) InternalServerError(w http.ResponseWriter, r *http.Request, err error) {
	skylb.Log.TraceRequest(r)
	isProd, _ := r.Context().Value(ContextIsProd).(bool)
	var diagnosis Diagnosis
	type Data struct {
		IsProd        bool
		RequestID     string
		Error         string
		OnelinerError string
		Diagnosis     Diagnosis
		URL           *url.URL
	}
	requestID := logutil.GetReqID(r.Context())
	prettifiedError := erro.Sdump(err)
	onelinerError := erro.S1dump(err)
	diagnosis = skylb.diagnoseError(err, r)
	data := Data{
		IsProd:        isProd,
		RequestID:     requestID,
		Error:         prettifiedError,
		OnelinerError: onelinerError,
		Diagnosis:     diagnosis,
		URL:           r.URL,
	}
	log.Printf("RequestID:%s URL:%s %s\n\n", requestID, r.URL, onelinerError)
	w.WriteHeader(http.StatusInternalServerError)
	funcs := skylb.NavbarFuncs(nil, w, r)
	t, err := template.New("500.html").Funcs(funcs).ParseFiles("app/skylab/500.html", "app/skylab/navbar.html")
	if err != nil {
		fmt.Fprintf(w, "%s\n\n%s\n\nError parsing 500.html: %s\n", onelinerError, string(data.Diagnosis.HTML), err.Error())
		return
	}
	err = t.Execute(w, data)
	if err != nil {
		log.Printf("%s\n\nError executing 500.html: %s\n", onelinerError, err.Error())
		return
	}
}

type Diagnosis struct {
	Valid bool
	HTML  template.HTML
}

type diagnostic func(error, *http.Request) Diagnosis

func (skylb Skylab) diagnoseError(err error, r *http.Request) (diagnosis Diagnosis) {
	diagnostics := []diagnostic{
		isUndefinedTemplateFunction,
		obtainPostgresErrCode,
	}
	for _, diagnose := range diagnostics {
		diagnosis = diagnose(err, r)
		if diagnosis.Valid {
			break
		}
	}
	return diagnosis
}

func ahref(link string) string {
	return `<a href="` + link + `">` + link + `</a>`
}

func isUndefinedTemplateFunction(err error, r *http.Request) (diagnosis Diagnosis) {
	matches := regexp.MustCompile(`template: .+ function "(\w+)" not defined`).FindStringSubmatch(err.Error())
	if len(matches) < 2 {
		return diagnosis
	}
	templateFunction := matches[1]
	diagnosis.HTML = template.HTML(fmt.Sprintf(`
Check if the template function "%s" is defined in one of these places:
<ul>
	<li>Global template functions are declared in app.Skylab.Render() in "app/skylab/render.go"</li>
	<li>Handler specific template functions may also be declared in the handler(s) of the url <code>%s</code> and passed into app.skylab.Render()</li>
	<li>If %s is a section you are trying to use, make sure it has been added into the sectionSymbols map</li>
</ul>
For more information on template functions, check out %s or the official docs at %s.
`,
		templateFunction,
		r.URL.Path,
		templateFunction,
		ahref("https://forum.golangbridge.org/t/how-to-call-a-custom-function-in-golang-template/4934"),
		ahref("https://golang.org/pkg/text/template/#FuncMap"),
	))
	diagnosis.Valid = true
	return diagnosis
}

func obtainPostgresErrCode(err error, r *http.Request) (diagnosis Diagnosis) {
	if pqerr, ok := erro.AsPqError(err); ok {
		diagnosis.HTML = template.HTML(fmt.Sprintf(`
The postgres error code is <code>%s</code>. Check the postgres error code reference at %s.
If it is a custom error (starts with the capital "O"), check app/skylab/errors.go instead.
`,
			string(pqerr.Code),
			ahref("https://www.postgresql.org/docs/12/errcodes-appendix.html"),
		))
		diagnosis.Valid = true
	}
	return diagnosis
}
