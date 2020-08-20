package app

import (
	"net/http"
	"strings"
	"testing"

	"github.com/bokwoon95/nusskylabx/app/skylab"
	"github.com/bokwoon95/nusskylabx/helpers/erro"
	"github.com/go-chi/chi"
)

type App struct{ skylb skylab.Skylab }

func NewSkylab(config skylab.Config) (skylab.Skylab, error) {
	skylb, err := skylab.New(config)
	if err != nil {
		return skylb, erro.Wrap(err)
	}
	AllRoutes(skylb)
	ServeDirectory(skylb.Mux, http.Dir(skylab.ProjectRootDir+"static"), "/static")
	return skylb, nil
}

func NewTestSkylab(t *testing.T) skylab.Skylab {
	skylab.LoadDotenv()
	skylb := skylab.NewTestDefault(t)
	AllRoutes(skylb)
	ServeDirectory(skylb.Mux, http.Dir(skylab.ProjectRootDir+"static"), "/static")
	return skylb
}

// Serve directory will serve a directory on a given url (for a *chi.Mux). You
// can convert a string path into a http.FileSystem by casting it to a
// http.Dir(). For example, this will serve a directory called 'assets' on a
// url '/my-assets':
//
//	ServeDirectory(mux, http.Dir("./assets"), "/my-assets")
//
// Taken from https://github.com/go-chi/chi/issues/35#issuecomment-465747073
func ServeDirectory(mux *chi.Mux, directory http.FileSystem, url string) {
	if strings.ContainsAny(url, "{}*") {
		panic("FileServer does not permit URL parameters.")
	}
	fs := http.StripPrefix(url, http.FileServer(directory))
	if url != "/" && url[len(url)-1] != '/' {
		mux.Get(url, http.RedirectHandler(url+"/", 301).ServeHTTP)
		url += "/"
	}
	url += "*"
	mux.Get(url, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	}))
}
