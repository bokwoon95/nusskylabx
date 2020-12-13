package admins

import (
	"fmt"
	"net/http"
	"runtime"
	"strings"

	"github.com/bokwoon95/nusskylabx/app/skylab"

	"github.com/bokwoon95/nusskylabx/helpers/cookies"
	"github.com/bokwoon95/nusskylabx/helpers/headers"
)

func (adm Admins) DumpJson(w http.ResponseWriter, r *http.Request) {
	adm.skylb.Log.TraceRequest(r)
	r = adm.skylb.SetRoleSection(w, r, skylab.RoleAdmin, skylab.AdminDumpJson)
	headers.DoNotCache(w)
	_, filename, linenr, _ := runtime.Caller(0)
	templatename := strings.ReplaceAll(filename, ".go", ".html")
	data := map[string]interface{}{
		"Title": `Enter a url to reveal the pages's underlying data (in JSON)`,
		"Description": fmt.Sprintf(`
Alternatively, add the query parameter '<pre class="di">dumpjson=true</pre>' to any url to achieve the same effect.
Try it with this page! Enter '<pre class="di">%s%s</pre>' into the text box below or append '<pre class="di">?dumpjson=true</pre>' to the url in the url bar and enter. <br>
`, adm.skylb.BaseURLWithProtocol(), skylab.AdminDumpJson),
		"DemoLink":    fmt.Sprintf("%s%s?dumpjson=true", adm.skylb.BaseURLWithProtocol(), skylab.AdminDumpJson),
		"PageData":    fmt.Sprintf(`This page is <pre class="di">%s</pre> and the data comes from <pre class="di">%s</pre>.`, templatename, filename),
		"Meta":        fmt.Sprintf("To verify that the data you see here is accurate, look in the file %s:%d to locate where the data is set. Slices, maps and structs translate quite comfortably into JSON arrays and objects", filename, linenr+2),
		"Plugins":     "To view this data comfortably in a browser, you will need a browser JSON formatter. Firefox has one by default, for Chrome I recommend the DJSON plugin, for Safari you can download the JSON Peep extension.",
		"SampleArray": []int{1, 2, 3, 4, 5},
		"SampleMap": map[string]string{
			"Chrome":  "https://chrome.google.com/webstore/detail/djson-json-viewer-formatt/chaeijjekipecdajnijdldjjipaegdjc?hl=en",
			"Safari":  "https://apps.apple.com/sg/app/json-peep-for-safari/id1458969831",
			"Firefox": "already built in",
		},
	}
	adm.skylb.Render(w, r, data, "app/admins/dump_json.html")
}

func (adm Admins) DumpJsonPost(w http.ResponseWriter, r *http.Request) {
	adm.skylb.Log.TraceRequest(r)
	url := r.FormValue("url")
	if url != "" {
		cookies.SetCookieOneMinute(w, string(skylab.ContextDumpJson), "true")
		http.Redirect(w, r, url, http.StatusMovedPermanently)
		return
	}
	http.Redirect(w, r, skylab.AdminDumpJson, http.StatusMovedPermanently)
}
