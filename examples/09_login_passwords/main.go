package main

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/bokwoon95/nusskylabx/helpers/auth"
	"github.com/bokwoon95/nusskylabx/helpers/cookies"
	"github.com/bokwoon95/nusskylabx/helpers/headers"
	"github.com/bokwoon95/nusskylabx/helpers/logutil"
	"github.com/bokwoon95/nusskylabx/helpers/random"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

type User struct {
	Username string

	// Password is stored here only for illustrative purposes. In a normal
	// system you DO NOT store the user's passwords at all! You only store the
	// password hash, and whenever you need to check if a user-supplied
	// password is correct you simply hash it and check if it matches the
	// password hash.
	Password, PasswordHash string
}

var (
	_, sourcefile, _, _ = runtime.Caller(0)                                   // sourcefile is the path to this file
	rootdir             = filepath.Dir(sourcefile) + string(os.PathSeparator) // rootdir is the directory containing this file
)

// Log is a global logger variable
var Log = logutil.NewLogger(os.Stdin)

// Users is a global store that maps usernames to Users. It stores the users
// that have signed up.
var Users = make(map[string]User)

// Sessions is a global store that maps sessionIDs to usernames. It stores the
// signed up users that are logged in.
var Sessions = make(map[string]string)

const SessionCookieName = "09_login_passwords_session_cookie"

func main() {
	mux := chi.NewRouter()
	mux.Use(middleware.RequestID) // Insert a unique ID in every request
	mux.Use(middleware.Recoverer) // pretty print any panicked errors
	mux.Use(middleware.Logger)    // Log all paths hit

	mux.Get("/", Home)
	mux.Post("/signup", Signup)
	mux.Post("/login", Login)
	mux.Post("/logout", Logout)

	fmt.Println("Listening on localhost:8009")
	http.ListenAndServe(":8009", mux)
}

func Home(w http.ResponseWriter, r *http.Request) {
	Log.TraceRequest(r)
	headers.DoNotCache(w)
	type Data struct {
		IsLoggedIn bool
		User       User
		Users      []User
	}
	var data Data

	// Grab the list of all the signed up users
	for _, user := range Users {
		data.Users = append(data.Users, user)
	}
	// Checks to cookies to see if there is a valid user session
	sessionID := cookies.GetCookieValue(r, SessionCookieName)
	if username, ok := Sessions[sessionID]; ok {
		user := Users[username]
		data.User = user
		data.IsLoggedIn = true
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

func Signup(w http.ResponseWriter, r *http.Request) {
	Log.TraceRequest(r)
	username := r.FormValue("username")
	password := r.FormValue("password")
	passwordHash, err := auth.HashPassword(password)
	if err != nil {
		panic(err)
	}
	user := User{
		Username:     username,
		Password:     password,
		PasswordHash: passwordHash,
	}
	Users[username] = user // Add user to global store
	SetSession(w, user)    // Set session for user
	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

func Login(w http.ResponseWriter, r *http.Request) {
	Log.TraceRequest(r)
	username := r.FormValue("username")
	password := r.FormValue("password")
	passwordHash := Users[username].PasswordHash
	err := auth.CompareHashAndPassword(passwordHash, password)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintf(w, "Username '%s' or password '%s' is incorrect\n", username, password)
		return
	}
	user := User{
		Username:     username,
		Password:     password,
		PasswordHash: passwordHash,
	}
	SetSession(w, user) // Set session for user
	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

func Logout(w http.ResponseWriter, r *http.Request) {
	Log.TraceRequest(r)
	sessionID := cookies.GetCookieValue(r, SessionCookieName)
	delete(Users, sessionID)
	cookies.DeleteCookie(w, SessionCookieName)
	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

func SetSession(w http.ResponseWriter, user User) {
	Log.TraceFunc()
	sessionID := random.SecretKey()
	Sessions[sessionID] = user.Username
	cookies.SetCookie(w, SessionCookieName, sessionID)
	Log.Printf("Setting sessionID '%s' into cookie '%s' for user %+v", sessionID, SessionCookieName, user)
}
