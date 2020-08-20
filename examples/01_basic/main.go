package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"text/template"
	"time"

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

	mux.Get("/", HelloWorld)
	mux.Get("/number", RandomNumber)
	fmt.Println("Listening on localhost:8001")
	log.Fatalln(http.ListenAndServe(":8001", mux))
}

func HelloWorld(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles(rootdir + "hello_world.html")
	if err != nil {
		panic(err)
	}
	err = t.Execute(w, nil)
	if err != nil {
		panic(err)
	}
}

func RandomNumber(w http.ResponseWriter, r *http.Request) {
	rand.Seed(time.Now().UnixNano())
	type Data struct {
		Number int
	}
	var data Data
	data.Number = rand.Intn(math.MaxInt32)

	t, err := template.ParseFiles(rootdir + "random_number.html")
	if err != nil {
		log.Fatalln(err)
	}
	err = t.Execute(w, data)
	if err != nil {
		panic(err)
	}
}
