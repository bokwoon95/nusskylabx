package templateloader

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"

	"github.com/bokwoon95/nusskylabx/helpers/erro"
	"github.com/davecgh/go-spew/spew"
	"github.com/oxtoacart/bpool"
)

// TODO: The mechanism for plugins to dump their data into the global map[string]interface{} and accessing it at a known namespace is quite simple.
// Basically their template namespace can be configured beforehand, at Parse() time. That way the user and plugin author always knows what namespace they can access their variables at. The user can also choose to tweak the namespace accordingly to avoid potential template namespace conflicts.
// Then, users are encouraged to pass the dot to the template as-is without narrowing down anything, thus allowing plugin templates to access their own data at the predetermined namespace.
// This may let templates potentially snoop at user data but essentially plugins must be trusted first before they can be used.
// That way we can sidestep any trust issues by saying it's up to the user to screen for security issues before trusting the plugin.

type ctxkey int

const (
	RenderJSON ctxkey = iota
)

type Source struct {
	// equivalent html/template call:
	// t.New(src.Name).Funcs(src.Funcs).Option(src.Options...).Parse(src.Text)
	Name      string
	Filepaths []string
	Text      string
	Funcs     map[string]interface{}
	Options   []string
}

type Sources struct {
	Templates       []Source
	CommonTemplates []Source
	CommonFuncs     map[string]interface{}
	CommonOptions   []string
}

type Template struct {
	Name string
	HTML *template.Template
	CSS  []template.CSS
	JS   []template.JS
}

type Templates struct {
	bufpool *bpool.BufferPool
	common  *template.Template            // gets included in every template in the cache
	lib     map[string]*template.Template // never gets executed, main purpose for cloning
	cache   map[string]*template.Template // is what gets executed, should not changed after it is set
	funcs   map[string]interface{}
	opts    []string
}

type ParseOption func(*Sources) error

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

func Parse(opts ...ParseOption) (*Templates, error) {
	var err error
	ts := &Templates{
		bufpool: bpool.NewBufferPool(64),
		common:  template.New(""),
		lib:     make(map[string]*template.Template),
		cache:   make(map[string]*template.Template),
	}
	srcs := &Sources{
		CommonFuncs: make(map[string]interface{}),
	}
	for _, opt := range opts {
		err = opt(srcs)
		if err != nil {
			return ts, err
		}
	}
	ts.opts = srcs.CommonOptions // clone options
	if len(srcs.CommonFuncs) > 0 {
		ts.common = ts.common.Funcs(srcs.CommonFuncs)
	}
	if len(srcs.CommonOptions) > 0 {
		ts.common = ts.common.Option(srcs.CommonOptions...)
	}
	for _, src := range srcs.CommonTemplates {
		ts.common, err = ts.common.New(src.Name).Parse(src.Text)
		if err != nil {
			return ts, err
		}
	}
	for _, src := range srcs.Templates {
		var tmpl, cacheEntry *template.Template
		tmpl, err = template.New(src.Name).Funcs(srcs.CommonFuncs).Option(srcs.CommonOptions...).Parse(src.Text)
		if err != nil {
			return ts, err
		}
		ts.lib[src.Name] = tmpl
		cacheEntry, err = ts.common.Clone()
		if err != nil {
			return ts, err
		}
		cacheEntry = cacheEntry.Option(srcs.CommonOptions...)
		err = addParseTree(cacheEntry, tmpl)
		if err != nil {
			return ts, err
		}
		ts.cache[src.Name] = cacheEntry
	}
	return ts, nil
}

func AddParse(base *Templates, opts ...ParseOption) (*Templates, error) {
	var err error
	ts := &Templates{
		bufpool: bpool.NewBufferPool(64),
		lib:     make(map[string]*template.Template),
		cache:   make(map[string]*template.Template),
	}
	// Clone base.common
	ts.common, err = base.common.Clone()
	if err != nil {
		return ts, err
	}
	// Clone base.lib and regenerate base.cache
	for name, tmpl := range base.lib {
		libTmpl, err := tmpl.Clone()
		if err != nil {
			return ts, err
		}
		ts.lib[name] = libTmpl
		cacheEntry, err := ts.common.Clone()
		if err != nil {
			return ts, err
		}
		cacheEntry = cacheEntry.Option(base.opts...) // clone options
		err = addParseTree(cacheEntry, libTmpl)
		if err != nil {
			return ts, err
		}
		ts.cache[name] = cacheEntry
	}
	srcs := &Sources{
		CommonFuncs: make(map[string]interface{}),
	}
	for _, opt := range opts {
		err = opt(srcs)
		if err != nil {
			return ts, err
		}
	}
	if len(srcs.CommonFuncs) > 0 {
		ts.common = ts.common.Funcs(srcs.CommonFuncs)
	}
	if len(srcs.CommonOptions) > 0 {
		ts.common = ts.common.Option(srcs.CommonOptions...)
	}
	for _, src := range srcs.CommonTemplates {
		ts.common, err = ts.common.New(src.Name).Parse(src.Text)
		if err != nil {
			return ts, err
		}
	}
	for _, src := range srcs.Templates {
		var tmpl, cacheEntry *template.Template
		tmpl, err = template.New(src.Name).Funcs(srcs.CommonFuncs).Option(srcs.CommonOptions...).Parse(src.Text)
		if err != nil {
			return ts, err
		}
		ts.lib[src.Name] = tmpl
		cacheEntry, err = ts.common.Clone()
		if err != nil {
			return ts, err
		}
		cacheEntry = cacheEntry.Option(srcs.CommonOptions...)
		err = addParseTree(cacheEntry, tmpl)
		if err != nil {
			return ts, err
		}
		ts.cache[src.Name] = cacheEntry
	}
	return ts, nil
}
func AddCommonFiles(filepatterns ...string) ParseOption {
	return func(srcs *Sources) error {
		for _, filepattern := range filepatterns {
			filenames, err := filepath.Glob(filepattern)
			if err != nil {
				return err
			}
			for _, filename := range filenames {
				src := Source{}
				b, err := ioutil.ReadFile(filename)
				if err != nil {
					return err
				}
				src.Text = string(b)
				src.Filepaths = append(src.Filepaths, filename)
				// check if user already defined a template called `filename` inside the template itself
				re, err := regexp.Compile(`{{\s*define\s+["` + "`" + `]` + filename + `["` + "`" + `]\s*}}`)
				if err != nil {
					return err
				}
				if !re.MatchString(string(b)) {
					src.Name = filename
				}
				srcs.CommonTemplates = append(srcs.CommonTemplates, src)
			}
		}
		return nil
	}
}

func AddFiles(filepatterns ...string) ParseOption {
	return func(srcs *Sources) error {
		for _, filepattern := range filepatterns {
			filenames, err := filepath.Glob(filepattern)
			if err != nil {
				return err
			}
			for _, filename := range filenames {
				src := Source{}
				b, err := ioutil.ReadFile(filename)
				if err != nil {
					return err
				}
				src.Text = string(b)
				src.Filepaths = append(src.Filepaths, filename)
				// check if user already defined a template called `filename` inside the template itself
				re, err := regexp.Compile(`{{\s*define\s+["` + "`" + `]` + filename + `["` + "`" + `]\s*}}`)
				if err != nil {
					return err
				}
				if !re.MatchString(string(b)) {
					src.Name = filename
				}
				srcs.Templates = append(srcs.Templates, src)
			}
		}
		return nil
	}
}

func Funcs(funcs map[string]interface{}) ParseOption {
	return func(srcs *Sources) error {
		for name, fn := range funcs {
			srcs.CommonFuncs[name] = fn
		}
		return nil
	}
}

func Option(opts ...string) ParseOption {
	return func(srcs *Sources) error {
		srcs.CommonOptions = append(srcs.CommonOptions, opts...)
		return nil
	}
}

func lookup(ts *Templates, name string) (tmpl *template.Template, isCommon bool) {
	tmpl = ts.lib[name]
	if tmpl != nil {
		return tmpl, false
	}
	tmpl = ts.common.Lookup(name)
	if tmpl != nil {
		return tmpl, true
	}
	return nil, false
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

func (ts *Templates) Render(w http.ResponseWriter, r *http.Request, data map[string]interface{}, name string, names ...string) error {
	// check if render JSON
	renderJSON, _ := r.Context().Value(RenderJSON).(bool)
	if renderJSON {
		sanitizedData, err := sanitizeObject(data)
		if err != nil {
			return erro.Wrap(err)
		}
		b, err := json.Marshal(sanitizedData)
		if err != nil {
			s := spew.Sdump(data)
			b = []byte(s)
		}
		w.Write(b)
		return nil
	}
	// check if the template being rendered exists
	tmpl, isCommon := lookup(ts, name)
	if tmpl == nil {
		return fmt.Errorf("No such template '%s'\n", name)
	}
	if isCommon {
		err := executeTemplate(ts.common, w, ts.bufpool, name, data)
		if err != nil {
			return err
		}
		return nil
	}
	// used cached version if exists...
	fullname := strings.Join(append([]string{name}, names...), "\n")
	if tmpl, ok := ts.cache[fullname]; ok {
		err := executeTemplate(tmpl, w, ts.bufpool, name, data)
		if err != nil {
			return err
		}
		return nil
	}
	// ...otherwise generate ad-hoc template and cache it
	cacheEntry, err := ts.common.Clone()
	if err != nil {
		return err
	}
	cacheEntry = cacheEntry.Option(ts.opts...)
	for _, nm := range names {
		tmpl, _ := lookup(ts, nm)
		if tmpl == nil {
			return fmt.Errorf("No such template '%s'\n", nm)
		}
		err := addParseTree(cacheEntry, tmpl)
		if err != nil {
			return err
		}
	}
	ts.cache[fullname] = cacheEntry
	err = executeTemplate(cacheEntry, w, ts.bufpool, name, data)
	if err != nil {
		return err
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

func RenderJSONHandler(activationParam string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_ = r.ParseForm()
			if _, ok := r.Form[activationParam]; ok {
				ctx := r.Context()
				r = r.WithContext(context.WithValue(ctx, RenderJSON, true))
			}
			next.ServeHTTP(w, r)
		})
	}
}

func sanitizeObject(object map[string]interface{}) (map[string]interface{}, error) {
	output := make(map[string]interface{})
	for key, value := range object {
		switch value := value.(type) {
		case nil: // null
			output[key] = value
		case string: // string
			output[key] = value
		case int, int8, int16, int32, int64, uint, uint8,
			uint16, uint64, uintptr, float32, float64: // number
			output[key] = value
		case map[string]interface{}: // object
			v, err := sanitizeObject(value)
			if err != nil {
				return output, err
			}
			output[key] = v
		case []interface{}: // array
			v, err := sanitizeArray(value)
			if err != nil {
				return output, err
			}
			output[key] = v
		default:
			v, err := sanitizeInterface(value)
			if err != nil {
				return output, err
			}
			output[key] = v
		}
	}
	return output, nil
}

func sanitizeArray(array []interface{}) ([]interface{}, error) {
	var output []interface{}
	for _, item := range array {
		switch value := item.(type) {
		case nil: // null
			output = append(output, value)
		case string: // string
			output = append(output, value)
		case int, int8, int16, int32, int64, uint, uint8,
			uint16, uint64, uintptr, float32, float64: // number
			output = append(output, value)
		case map[string]interface{}: // object
			v, err := sanitizeObject(value)
			if err != nil {
				return output, err
			}
			output = append(output, v)
		case []interface{}: // array
			v, err := sanitizeArray(value)
			if err != nil {
				return output, err
			}
			output = append(output, v)
		default:
			v, err := sanitizeInterface(value)
			if err != nil {
				return output, err
			}
			output = append(output, v)
		}
	}
	return output, nil
}

func sanitizeInterface(v interface{}) (interface{}, error) {
	var output interface{}
	switch vv := reflect.ValueOf(v); vv.Kind() {
	case reflect.Array: // K
		output = v
	case reflect.Chan: // K
		output = v
	case reflect.Func: // ?
		return funcType(v), nil
	case reflect.Interface: // ?
		output = v
	case reflect.Map: // K,V
		output = v
	case reflect.Ptr: // K
		output = v
	case reflect.Slice: // K
		output = v
	case reflect.Struct: // K
		output = v
	case reflect.Complex64, reflect.Complex128: // unsupported
		return output, fmt.Errorf("unsupported type: complex number")
	case reflect.UnsafePointer: // unsupported
		return output, fmt.Errorf("unsupported type: unsafe.Pointer")
	case reflect.Invalid: // unsupported
		return output, fmt.Errorf("unsupported type: reflect.Invalid")
	default:
		output = v
	}
	return output, nil
}

func funcType(f interface{}) string {
	t := reflect.TypeOf(f)
	if t.Kind() != reflect.Func {
		return "<not a function>"
	}
	buf := strings.Builder{}
	buf.WriteString("func(")
	for i := 0; i < t.NumIn(); i++ {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(t.In(i).String())
	}
	buf.WriteString(")")
	if numOut := t.NumOut(); numOut > 0 {
		if numOut > 1 {
			buf.WriteString(" (")
		} else {
			buf.WriteString(" ")
		}
		for i := 0; i < t.NumOut(); i++ {
			if i > 0 {
				buf.WriteString(", ")
			}
			buf.WriteString(t.Out(i).String())
		}
		if numOut > 1 {
			buf.WriteString(")")
		}
	}
	return buf.String()
}
