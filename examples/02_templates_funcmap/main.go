package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"text/template"

	"github.com/bokwoon95/nusskylabx/helpers/random"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

var (
	_, sourcefile, _, _ = runtime.Caller(0)                                   // sourcefile is the path to this file
	rootdir             = filepath.Dir(sourcefile) + string(os.PathSeparator) // rootdir is the directory containing this file
)

func main() {
	mux := chi.NewRouter()
	mux.Use(middleware.Recoverer) // pretty print any panicked errors

	mux.Get("/", TemplatesFuncmap)
	fmt.Println("Listening on localhost:8002")
	log.Fatalln(http.ListenAndServe(":8002", mux))
}

func TemplatesFuncmap(w http.ResponseWriter, r *http.Request) {
	type Subsection struct {
		Filename   string
		RandomWord string
	}
	type Data struct {
		True               bool
		Five               int
		SubsectionOne      Subsection
		SubsectionTwo      Subsection
		TemperatureCelcius float64
	}
	var data = Data{
		True: true,
		Five: 5,
		SubsectionOne: Subsection{
			Filename:   rootdir + "main.html",
			RandomWord: random.Word(),
		},
		SubsectionTwo: Subsection{
			Filename:   rootdir + "subsection_two.html",
			RandomWord: random.Word(),
		},
		TemperatureCelcius: random.Float64(),
	}
	funcs := map[string]interface{}{
		"C_To_F":      CelciusToFarenheit,
		"hello_world": HelloWorld,
	}
	t, err := template.
		New("main.html").
		Funcs(funcs).
		ParseFiles(
			rootdir+"main.html",
			rootdir+"subsection_two.html",
			rootdir+"hello_world.html",
		)
	if err != nil {
		panic(err)
	}
	err = t.Execute(w, data)
	if err != nil {
		panic(err)
	}
}

func CelciusToFarenheit(celcius float64) float64 {
	return (celcius * 9 / 5) + 32
}

func HelloWorld() string {
	return "Hello World!"
}
