package main

import (
	"fmt"
	"log"
	"os"
	"text/template"

	"github.com/davecgh/go-spew/spew"
)

func Map(args ...interface{}) map[string]interface{} {
	data := make(map[string]interface{})
	for i := 0; i < len(args); i += 2 {
		if i+1 >= len(args) {
			break
		}
		key := fmt.Sprint(args[i])
		value := args[i+1]
		data[key] = value
		fmt.Println(key+":", spew.Sdump(value))
	}
	return data
}

var str = `
[OG] {{SpewDump .}}
{{template "sidebar" Map "skylab" $.skylab "Section" "/student/dashboard" "Icon" "dashboard_svg" "Display" "Dashboard"}}
{{define "sidebar"}}
[sidebar] {{SpewDump .}}
{{$.skylab.One}}, {{$.Display}}
{{end}}
`

func main() {
	funcs := map[string]interface{}{
		"Map":      Map,
		"SpewDump": spew.Sdump,
	}
	t, err := template.New("").Funcs(funcs).Parse(str)
	if err != nil {
		log.Fatalln(err)
	}
	data := map[string]interface{}{
		"skylab": map[string]interface{}{
			"One":   1,
			"Two":   2,
			"Three": 3,
		},
	}
	err = t.Execute(os.Stdout, data)
	if err != nil {
		log.Fatalln(err)
	}
}
