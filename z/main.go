package main

import (
	"html/template"
	"log"
	"os"
)

var funcs0 = map[string]interface{}{
	"Map": func(args ...interface{}) string { return "Map0" },
}

var funcs1 = map[string]interface{}{
	"Map": func(args ...interface{}) string { return "Map1" },
}

var funcs2 = map[string]interface{}{
	"Map": func(args ...interface{}) string { return "Map2" },
}

var t1_data = `t1 {{Map}}`

var t2_data = `t2 {{Map}}`

var t3_data = `{{template "t1"}}, {template "t2"}}, t3`

func main() {
	log.SetOutput(os.Stdout)
	common := template.New("").Funcs(funcs0)
	cacheEntry := template.Must(common.Clone())

	t1, err := template.New("t1").Funcs(funcs1).Parse(t1_data)
	if err != nil {
		log.Fatalln(err)
	}
	err = addParseTree(cacheEntry, t1)
	if err != nil {
		log.Fatalln(err)
	}

	t2, err := template.New("t2").Funcs(funcs2).Parse(t2_data)
	if err != nil {
		log.Fatalln(err)
	}
	err = addParseTree(cacheEntry, t2)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println(cacheEntry.DefinedTemplates())
	err = cacheEntry.ExecuteTemplate(os.Stdout, "t1", nil)
	if err != nil {
		log.Fatalln(err)
	}

}

func addParseTree(parent *template.Template, child *template.Template) error {
	var err error
	for _, t := range child.Templates() {
		_, err = parent.AddParseTree(t.Name(), t.Tree)
		if err != nil {
			return err
		}
	}
	return nil
}
