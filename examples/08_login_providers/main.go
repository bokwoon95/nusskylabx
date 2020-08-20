package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/bokwoon95/nusskylabx/app/skylab"
	"github.com/bokwoon95/nusskylabx/helpers/auth"
	"github.com/bokwoon95/nusskylabx/helpers/cookies"
	"github.com/bokwoon95/nusskylabx/helpers/headers"
	"github.com/bokwoon95/nusskylabx/helpers/logutil"
	"github.com/bokwoon95/nusskylabx/helpers/random"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

var (
	_, sourcefile, _, _ = runtime.Caller(0)                                   // sourcefile is the path to this file
	rootdir             = filepath.Dir(sourcefile) + string(os.PathSeparator) // rootdir is the directory containing this file
)

type User struct {
	Username, Displayname, Email string
}

var Log = logutil.NewLogger(os.Stdin)
var Users = make(map[string]User)

const SessionCookieName = "07_login_providers_session_cookie"

func main() {
	skylab.LoadDotenv()
	PORT := os.Getenv("PORT")

	mux := chi.NewRouter()
	mux.Use(middleware.RequestID) // Insert a unique ID in every request
	mux.Use(middleware.Recoverer) // pretty print any panicked errors
	mux.Use(middleware.Logger)    // Log all paths hit

	callbackURL := fmt.Sprintf("http://localhost:%s/login/callback", PORT)
	mux.Get("/", Home)
	mux.Get("/login", func(w http.ResponseWriter, r *http.Request) {
		provider := r.FormValue("provider")
		if provider == "" {
			panic("provider cannot be blank")
		}
		auth.Redirect(w, r, provider, callbackURL, errorHandler)
	})
	mux.With(auth.Authenticate(callbackURL, errorHandler)).Get("/login/callback", RedirectHome)
	mux.Post("/logout", Logout)
	fmt.Println("Listening on localhost:" + PORT)
	err := http.ListenAndServe(":"+PORT, mux)
	if strings.Contains(err.Error(), "address already in use") {
		log.Fatalf("Application already listening on port :%s, please terminate that application first", PORT)
	}
	log.Fatalln(err)
}

func errorHandler(w http.ResponseWriter, r *http.Request, err error) {
	Log.TraceRequest(r)
	panic(err)
}

func Home(w http.ResponseWriter, r *http.Request) {
	Log.TraceRequest(r)
	headers.DoNotCache(w)
	type Data struct {
		IsLoggedIn bool
		User       User
	}
	var data Data
	sessionID := cookies.GetCookieValue(r, SessionCookieName)
	if user, ok := Users[sessionID]; ok {
		Log.Printf("Found a valid sessionID '%s' from cookie '%s'", sessionID, SessionCookieName)
		data.User = user
		data.IsLoggedIn = true
	} else {
		data.IsLoggedIn = false
	}

	t, err := template.ParseFiles(rootdir + "main.html")
	if err != nil {
		panic(err)
	}
	err = t.Execute(w, data)
	if err != nil {
		panic(err)
	}
}

func RedirectHome(w http.ResponseWriter, r *http.Request) {
	Log.TraceRequest(r)
	// The auth.Authenticate middleware is called before RedirectHome, which is
	// why username, displayname and email are available in the request context
	username, ok1 := r.Context().Value("username").(string)
	displayname, ok2 := r.Context().Value("displayname").(string)
	email, ok3 := r.Context().Value("email").(string)
	if !ok1 || !ok2 || !ok3 {
		panic(fmt.Sprintf(
			"[One or more of the following was not found] username:%s, displayname: %s, email: %s",
			username, displayname, email,
		))
	}

	sessionID := random.SecretKey()
	user := User{Username: username, Displayname: displayname, Email: email}
	Users[sessionID] = user
	cookies.SetCookie(w, SessionCookieName, sessionID)
	Log.Printf("Setting sessionID '%s' into cookie '%s' for user %+v", sessionID, SessionCookieName, user)
	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

func Logout(w http.ResponseWriter, r *http.Request) {
	Log.TraceRequest(r)
	sessionID := cookies.GetCookieValue(r, SessionCookieName)
	delete(Users, sessionID)
	cookies.DeleteCookie(w, SessionCookieName)
	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}
