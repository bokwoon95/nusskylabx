package templateloader

import (
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bokwoon95/nusskylabx/helpers/erro"
	"github.com/oxtoacart/bpool"
)

// TODO: The mechanism for plugins to dump their data into the global map[string]interface{} and accessing it at a known namespace is quite simple.
// Basically their template namespace can be configured beforehand, at Parse() time. That way the user and plugin author always knows what namespace they can access their variables at. The user can also choose to tweak the namespace accordingly to avoid potential template namespace conflicts.
// Then, users are encouraged to pass the dot to the template as-is without narrowing down anything, thus allowing plugin templates to access their own data at the predetermined namespace.
// This may let templates potentially snoop at user data but essentially plugins must be trusted first before they can be used.
// That way we can sidestep any trust issues by saying it's up to the user to screen for security issues before trusting the plugin.

type Templates struct {
	bufpool *bpool.BufferPool
	funcs   map[string]interface{}
	opts    []string
	common  *template.Template
	lib     map[string]*template.Template
	cache   map[string]*template.Template
}

func addParseTree(parent *template.Template, child *template.Template) error {
	var err error
	for _, t := range child.Templates() {
		_, err = parent.AddParseTree(t.Name(), t.Tree)
		if err != nil {
			return erro.Wrap(err)
		}
	}
	return nil
}

func executeTemplate(t *template.Template, w io.Writer, bufpool *bpool.BufferPool, name string, data map[string]interface{}) error {
	tempbuf := bufpool.Get()
	defer bufpool.Put(tempbuf)
	err := t.ExecuteTemplate(tempbuf, name, data)
	if err != nil {
		return erro.Wrap(err)
	}
	_, err = tempbuf.WriteTo(w)
	if err != nil {
		return erro.Wrap(err)
	}
	return nil
}

type Opt func(t *Templates)

func Funcs(funcs map[string]interface{}) func(*Templates) {
	return func(t *Templates) {
		t.funcs = funcs
	}
}

func Option(opts ...string) func(*Templates) {
	return func(t *Templates) {
		t.opts = opts
	}
}

func stripSpaces(s string) string {
	return s
}

func Parse(common []string, templates []string, opts ...Opt) (*Templates, error) {
	main := &Templates{
		bufpool: bpool.NewBufferPool(64),
		common:  template.New(""),
		lib:     make(map[string]*template.Template),
		cache:   make(map[string]*template.Template),
	}
	for _, opt := range opts {
		opt(main)
	}
	if len(main.funcs) > 0 {
		main.common = main.common.Funcs(main.funcs)
	}
	if len(main.opts) > 0 {
		main.common = main.common.Option(main.opts...)
	}
	for _, name := range common {
		files, err := filepath.Glob(name)
		if err != nil {
			return main, erro.Wrap(err)
		}
		for _, file := range files {
			b, err := ioutil.ReadFile(file)
			if err != nil {
				return main, erro.Wrap(err)
			}
			var t *template.Template
			re, err := regexp.Compile(`{{\s*define\s+["` + "`" + `]` + file + `["` + "`" + `]\s*}}`)
			if err != nil {
				return main, erro.Wrap(err)
			}
			if re.MatchString(string(b)) {
				t = template.New("")
			} else {
				t = template.New(file)
			}
			t, err = t.Funcs(main.funcs).Option(main.opts...).Parse(string(b))
			if err != nil {
				return main, erro.Wrap(err)
			}
			main.lib[file] = t
			// TODO: is addParseTree equivalent to doing a Parse? Will the Funcs and Options take effect?
			err = addParseTree(main.common, t)
			if err != nil {
				return main, erro.Wrap(err)
			}
		}
	}
	for _, name := range templates {
		files, err := filepath.Glob(name)
		if err != nil {
			return main, erro.Wrap(err)
		}
		for _, file := range files {
			b, err := ioutil.ReadFile(file)
			if err != nil {
				return main, erro.Wrap(err)
			}
			var t *template.Template
			re, err := regexp.Compile(`{{\s*define\s+["` + "`" + `]` + file + `["` + "`" + `]\s*}}`)
			if err != nil {
				return main, erro.Wrap(err)
			}
			if re.MatchString(string(b)) {
				t = template.New("")
			} else {
				t = template.New(file)
			}
			t, err = t.Funcs(main.funcs).Option(main.opts...).Parse(string(b))
			if err != nil {
				return main, erro.Wrap(err)
			}
			main.lib[file] = t
			cacheEntry, err := main.common.Clone()
			if err != nil {
				return main, erro.Wrap(err)
			}
			err = addParseTree(cacheEntry, t)
			if err != nil {
				return main, erro.Wrap(err)
			}
			main.cache[file] = cacheEntry
		}
	}
	return main, nil
}

func (main *Templates) Render(w http.ResponseWriter, r *http.Request, data map[string]interface{}, name string, names ...string) error {
	// "app/students/milestone_team_evaluation.html"
	mainTemplate, ok := main.lib[name]
	if !ok {
		return erro.Wrap(fmt.Errorf("No such template '%s'\n", name))
	}
	fullname := strings.Join(append([]string{name}, names...), "\n")
	// used cached version if exists...
	if t, ok := main.cache[fullname]; ok {
		err := executeTemplate(t, w, main.bufpool, name, data)
		if err != nil {
			return erro.Wrap(err)
		}
		return nil
	}
	// ...otherwise generate ad-hoc template and cache it
	cacheEntry, err := main.common.Clone()
	if err != nil {
		return erro.Wrap(err)
	}
	err = addParseTree(cacheEntry, mainTemplate)
	if err != nil {
		return erro.Wrap(err)
	}
	for _, nm := range names {
		t, ok := main.lib[nm]
		if !ok {
			return erro.Wrap(fmt.Errorf("No such template '%s'\n", nm))
		}
		err := addParseTree(cacheEntry, t)
		if err != nil {
			return erro.Wrap(err)
		}
	}
	main.cache[fullname] = cacheEntry
	err = executeTemplate(cacheEntry, w, main.bufpool, name, data)
	if err != nil {
		return erro.Wrap(err)
	}
	return nil
}

func (main *Templates) DefinedTemplates() string {
	buf := &strings.Builder{}
	buf.WriteString("; The defined templates are: ")
	i := 0
	for name := range main.lib {
		if i > 0 {
			buf.WriteString(", ")
		}
		i++
		buf.WriteString(`"`)
		buf.WriteString(name)
		buf.WriteString(`"`)
	}
	return buf.String()
}

func (main *Templates) Templates() []*template.Template {
	templates := make([]*template.Template, len(main.lib))
	i := 0
	for _, t := range main.lib {
		templates[i] = t
	}
	return templates
}
