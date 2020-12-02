package main

import (
	"io/ioutil"
	"log"
	"os"
	"text/template"
)

func toString(file string) string {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatalln(err)
	}
	return string(b)
}

func main() {
	mybatis := func() string { return "mybatis" }
	mybatis2 := func() string { return "mybatis2" }
	_, _ = mybatis, mybatis2

	t3, err := template.New("z/t3.html").Funcs(template.FuncMap{"mybatis": mybatis2}).Parse(toString("z/t3.html"))
	if err != nil {
		log.Fatalln(err)
	}
	t1, err := template.New("z/t1.html").Funcs(template.FuncMap{"mybatis": mybatis}).Parse(toString("z/t1.html"))
	if err != nil {
		log.Fatalln(err)
	}
	t2, err := template.New("z/t2.html").Parse(toString("z/t2.html"))
	if err != nil {
		log.Fatalln(err)
	}
	for _, t := range t1.Templates() {
		_, err = t3.AddParseTree(t.Name(), t.Tree)
		if err != nil {
			log.Fatalln(err)
		}
	}
	_, err = t3.AddParseTree(t2.Name(), t2.Tree)
	if err != nil {
		log.Fatalln(err)
	}
	_, _, _ = t1, t2, t3
	log.SetFlags(log.Llongfile)
	log.SetOutput(os.Stdout)
	log.Println(t1.DefinedTemplates())
	log.Println(t2.DefinedTemplates())
	log.Println(t3.DefinedTemplates())
	err = t3.ExecuteTemplate(os.Stdout, "z/t3.html", nil)
	if err != nil {
		log.Fatalln(err)
	}
}
