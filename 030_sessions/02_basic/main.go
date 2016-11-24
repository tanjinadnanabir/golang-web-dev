package main

import (
	"github.com/satori/go.uuid"
	"html/template"
	"log"
	"net/http"
	"time"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"os"
	"encoding/json"
)

type user struct {
	First, Last, UserName string
	Password []byte
}

var tpl *template.Template
var dbUsers = map[string]user{}
var dbSession = map[string]string{}

func init() {
	dbUsersLoad()
	tpl = template.Must(template.ParseGlob("templates/*"))
}

func main() {
	defer dbUsersSave()
	http.HandleFunc("/", index)
	http.HandleFunc("/signup", signup)
	http.HandleFunc("/login", login)
	http.HandleFunc("/logout", logout)
	http.HandleFunc("/loggedin", loggedin)
	http.HandleFunc("/elapsed", timer) //fyi
	http.Handle("/favicon.ico", http.NotFoundHandler())
	http.ListenAndServe(":8080", nil)
}

func signup(w http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodPost {
		// signup
		f := req.FormValue("firstname")
		l := req.FormValue("lastname")
		u := req.FormValue("username")
		p1 := req.FormValue("password1")
		p2 := req.FormValue("password2")
		if dbUsers[u] {
			http.Error(w, "Username taken", http.StatusUnprocessableEntity)
			return
		}
		if p1 != p2 {
			http.Error(w, "Passwords do not match", http.StatusUnprocessableEntity)
			return
		}
		bs, err := bcrypt.GenerateFromPassword([]byte(p1), bcrypt.MinCost)
		if err != nil {
			log.Fatalln("bcrypt didn't work,", err)
			return
		}
		dbUsers[u] = user{f,l,u,bs}
		createSession(w, req, u)
		http.Redirect(w, req, "/", http.StatusSeeOther)
		return
	}
	tpl.ExecuteTemplate(w, "signup.gohtml", nil)
}

func createSession(w http.ResponseWriter, req *http.Request, u string) {
	sID := uuid.NewV4()
	c := &http.Cookie{
		Name: "session",
		Value: sID.String(),
		//Secure: true,
		HttpOnly: true,
	}
	http.SetCookie(w, c)

	// dbSession is a map
	// the key is the session ID
	// the session ID is stored in the cookie (it's the cookie's value)
	// the value of dbSession is the username
	// the username is the key for dbUsers
	// dbUsers is a map
	// the value in dbUsers is a user
	// with a session ID, we can access user information
	dbSession[c.Value] = u
}

func login(w http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodPost {

		// retrieve form values
		un := req.FormValue("username")
		p := req.FormValue("password")

		// check credentials - Does the username exist in the db?
		u, ok := dbUsers[un]
		if !ok {
			http.Error(w, "Username and/or password incorrect", http.StatusForbidden)
			return
		}

		// check credentials - Does the password entered match the password for that user?
		err := bcrypt.CompareHashAndPassword(u.Password, []byte(p))
		if err != nil {
			http.Error(w, "Username and/or password incorrect", http.StatusForbidden)
			return
		}

		createSession(w, req, un)
		http.Redirect(w, req, "/", http.StatusSeeOther)
		return
	}
	tpl.ExecuteTemplate(w, "login.gohtml", nil)
}

func logout(w http.ResponseWriter, req *http.Request) {
	sID := getSession(w, req)
	delete(dbSession, sID)

}

func getSession(w http.ResponseWriter, req *http.Request) string {
	c, err := req.Cookie("session")
	log.Printf("cookie received from browser - %v", c) //fyi
	if err != nil {
		http.Redirect(w, req, "/login", http.StatusSeeOther)
		return
	}
	log.Printf("cookie returned to browser - %v", c) //fyi
	return c.Value
}


func index(w http.ResponseWriter, req *http.Request) {
	sID := getSession(w, req)
	fname := db[sID]
	if req.Method == http.MethodPost {
		fname = req.FormValue("firstname")
		db[sID] = fname
	}
	log.Printf("DB in index - %v\n\n", db) //fyi
	tpl.ExecuteTemplate(w, "index.gohtml", fname)
}



func access(w http.ResponseWriter, req *http.Request) {
	sID := getSession(w, req)
	fname := db[sID]                        // *** accessing session data based upon session ID !!!***
	log.Printf("DB in access - %v\n\n", db) //fyi
	tpl.ExecuteTemplate(w, "access.gohtml", fname)
}


//fyi - time the life of the cookie
var sessionStartTime time.Time

//fyi
func startSessionTimer() {
	sessionStartTime = time.Now()
}

//fyi - this does not call getSession therefore the session's MaxAge is not reset
func timer(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "Session time elapsed in seconds %v", time.Now().Sub(sessionStartTime).Seconds()) //fyi
}

func dbUsersLoad() {
	f, err := os.Open("db/simulatedDB.json")
	if err != nil {
		log.Fatalln("error opening json file", err)
	}
	defer f.Close()

	err = json.NewDecoder(f).Decode(&dbUsers)
	if err != nil {
		log.Fatalln("error decoding json", err)
	}
}

func dbUsersSave() {
	f, err := os.Create("db/simulatedDB.json")
	if err != nil {
		log.Fatalln("error creating json file", err)
	}
	defer f.Close()

	err = json.NewEncoder(f).Encode(&dbUsers)
	if err != nil {
		log.Fatalln("error encoding json", err)
	}
}