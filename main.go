package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"

	// "github.com/gobwas/glob/util/strings"
	// _ "github.com/mattn/go-sqlite3"
	_ "modernc.org/sqlite"

	"gopkg.in/cas.v2"
)

const (
	port   = ":8080"
	layout = "2006-01-02 15:04:05"
)

// structure for receiving the user modifiable information
type config_vars struct {
	Term     string `json:"term"`
	Year     int    `json:"year"`
	RepoPath string `json:"repopath"`
}

var currentTerm string
var currentYear int

var lastTerm string
var lastYear int

var pathToBinary string
var pathToRepo string

var term_start string
var term_end string

// set path to binary
func setPathToBinary() {
	e, _ := os.Executable()          // retrieves the path to the binary including the filename itself
	pathToBinary = path.Dir(e) + "/" // strips the filename only
}

// this sets/resets the term, year, and path variables
func setTermsYearsRepoPath() {
	config := config_vars{}
	configPath := pathToBinary + "config.json"
	jsonFile, _ := os.Open(configPath)
	file, _ := ioutil.ReadAll(jsonFile)
	_ = json.Unmarshal(file, &config)
	currentTerm = config.Term
	currentYear = config.Year
	pathToRepo = config.RepoPath
	setLastTermandYear()
	setTermStart()
	setTermEnd()
}

func setLastTermandYear() {
	if currentTerm == "Fall" {
		lastTerm = "Spring"
		lastYear = currentYear
	} else if currentTerm == "Spring" {
		lastTerm = "Fall"
		lastYear = currentYear - 1
	}
}

// converting term and year into layouts for time.Time
func setTermStart() {
	if currentTerm == "Fall" {
		term_start = strconv.Itoa(currentYear) + "-08-01 00:00:00"
	} else if currentTerm == "Spring" {
		term_start = strconv.Itoa(currentYear) + "-01-01 00:00:00"
	}
}
func setTermEnd() {
	if currentTerm == "Fall" {
		term_end = strconv.Itoa(currentYear) + "-12-31 23:59:59"
	} else if currentTerm == "Spring" {
		term_end = strconv.Itoa(currentYear) + "-05-30 00:00:00"
	}
}

// struct for housing the relevant data of a talk
type Talk struct {
	Id            int
	Date          time.Time
	Upcoming      bool
	Year          string
	Month         string
	Month_name    string
	Day           string
	Date_string   string
	Time_string   string
	Speaker_first string
	Speaker_last  string
	Speaker_url   string
	Affiliation   string
	Title         string
	Abstract      string
	Vid_conf_url  string
	Vid_conf_pw   string
	Recording_url string
	Host          string
	Location      string
}

// opens the datebase
func dbOpen(file string) (db *sql.DB) {
	dbDriver := "sqlite"
	db, err := sql.Open(dbDriver, pathToBinary+file+".db")
	if err != nil {
		log.Fatal(err)
	}
	return db
}

// parse template files
var tmpl = template.Must(template.ParseGlob(pathToBinary + "forms/*"))

// trash talk to provide some protection against undesired deletion
// var trash = Talk{}

func PrependHTTP(url string) string {
	if len(url) == 0 {
		return url
	} else if (url[0:7] != "http://") && (url[0:8] != "https://") {
		return "http://" + url
	} else {
		return url
	}
}

// helper function that reads a sql query result into a slice of talks
func rowsToTalkSlice(rows *sql.Rows) (talksslice []Talk) {
	talk := Talk{}
	talks := []Talk{}
	for rows.Next() {
		var id int
		var event_date, event_time, speaker_first, speaker_last, speaker_url, speaker_affiliation, title, abstract, vid_conf_url, vid_conf_pw, recording_url, host, location string
		err := rows.Scan(&id, &event_date, &event_time, &speaker_first, &speaker_last, &speaker_url, &speaker_affiliation, &title, &abstract, &vid_conf_url, &vid_conf_pw, &recording_url, &host, &location)
		if err != nil {
			log.Fatal(err)
		}
		talk.Id = id
		s := strings.Split(event_time, "-")
		if len(s[0]) == 4 {
			s[0] = "0" + s[0]
		}
		event_time = s[0]
		string_time := event_date + "T" + event_time + ":00.000Z"
		converted_time, err := time.Parse("2006-01-02T15:04:05.000Z", string_time)
		if err != nil {
			log.Fatal(err)
		}
		talk.Date = converted_time
		talk.Upcoming = converted_time.After(time.Now())
		talk.Date_string = talk.Date.Format("January 02 2006")
		talk.Year = strconv.Itoa(converted_time.Year())
		month_num := strconv.Itoa(int(converted_time.Month()))
		if len(month_num) == 1 {
			talk.Month = "0" + month_num
		} else {
			talk.Month = strconv.Itoa(int(converted_time.Month()))
		}
		talk.Month_name = talk.Date.Format("January")
		talk.Day = strconv.Itoa(converted_time.Day())
		talk.Time_string = converted_time.Format("15:04")
		talk.Speaker_first = speaker_first
		talk.Speaker_last = speaker_last
		talk.Speaker_url = speaker_url
		talk.Affiliation = speaker_affiliation
		talk.Title = title
		talk.Abstract = abstract
		talk.Vid_conf_url = vid_conf_url
		talk.Vid_conf_pw = vid_conf_pw
		talk.Recording_url = recording_url
		talk.Host = host
		talk.Location = location
		talks = append(talks, talk)
	}
	return talks
}

// reads the database into a slice of talks, converts to data for seminar page template, writes the resulting html to both the main index page and the relevant folder in the git repo, and finally calls the function to update the git repo
func WriteToHTML() {
	setTermsYearsRepoPath() // update term and year
	db := dbOpen("talks")
	rows, err := db.Query("SELECT * FROM scagnt ORDER BY id DESC")
	if err != nil {
		log.Fatal(err)
	}
	talks := rowsToTalkSlice(rows)
	sort.Slice(talks, func(i, j int) bool {
		return talks[j].Date.After(talks[i].Date)
	})
	start_date, _ := time.Parse(layout, term_start)
	end_date, _ := time.Parse(layout, term_end)
	current_talks := []Talk{}
	for _, talk := range talks {
		if talk.Date.After(start_date) && talk.Date.Before(end_date) {
			current_talks = append(current_talks, talk)
		}
	}
	type page_talk_data struct {
		Date         string
		Time         string
		URL          string
		Speaker      string
		Speaker_last string
		Uni          string
		Title        string
		Host         string
		Abstract     string
		Location     string
	}
	type page_data struct {
		Term         string
		Year         string
		LastTermPath string
		Talks        []page_talk_data
		Prefix       string
	}
	page_talks := []page_talk_data{}
	for _, talk := range current_talks {
		talk_data := page_talk_data{}
		talk_data.Date = talk.Date.Format("Monday, Jan 2")
		talk_data.Time = talk.Date.Format("3:04PM")
		talk_data.URL = talk.Speaker_url
		talk_data.Speaker = talk.Speaker_first + " " + talk.Speaker_last
		talk_data.Speaker_last = talk.Speaker_last
		talk_data.Uni = talk.Affiliation
		talk_data.Title = talk.Title
		talk_data.Host = talk.Host
		talk_data.Abstract = talk.Abstract
		talk_data.Location = talk.Location
		page_talks = append(page_talks, talk_data)
	}

	data := page_data{
		Term:         currentTerm,
		Year:         strconv.Itoa(currentYear),
		LastTermPath: strconv.Itoa(lastYear) + "/" + strings.ToLower(lastTerm),
		Talks:        page_talks,
	}
	pathToTemplate := pathToBinary + "html/Seminar_Page.html"
	t, err := template.ParseFiles(pathToTemplate)
	if err != nil {
		log.Println("Parsing index page template:", err)
		return
	}

	index_path := pathToRepo + "index.html"
	yearPath := pathToRepo + strconv.Itoa(currentYear)
	termPath := yearPath + "/" + strings.ToLower(currentTerm)
	termIndexPath := termPath + "/index.html"

	writeIndex := func(path string) {
		f, err := os.Create(path)
		if err != nil {
			log.Println("create file: ", err)
			return
		}
		if path != index_path {
			data.Prefix = "../../"
		}
		err = t.Execute(f, data)
		if err != nil {
			log.Println("execute: ", err)
			return
		}
		f.Close()
	}

	writeIndex(index_path)

	createDir := func(path string) {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			err := os.Mkdir(path, 0700)
			if err != nil {
				log.Println("creating directory: ", err)
			}
		}
	}

	createDir(yearPath) // need to create directories one level at a time apparently
	createDir(termPath)

	writeIndex(termIndexPath)

	go updateRepo()
}

func updateRepo() {
	gitPull := exec.Command("git", "pull")
	gitPull.Dir = pathToRepo
	output, err := gitPull.Output()
	log.Printf("%s", output)
	if err != nil {
		log.Println("git pull: ", err)
	}

	gitAdd := exec.Command("git", "add", "-A")
	gitAdd.Dir = pathToRepo
	_, err = gitAdd.Output()
	if err != nil {
		log.Println("git add: ", err)
	}

	gitCommit := exec.Command("git", "commit", "-m", "'Auto update to repo'")
	gitCommit.Dir = pathToRepo
	_, err = gitCommit.Output()
	if err != nil {
		log.Println("git commit: ", err)
	}

	gitPush := exec.Command("git", "push")
	gitPush.Dir = pathToRepo
	output, err = gitPush.Output()
	log.Printf("%s", output)
	if err != nil {
		log.Println("git push: ", err)
	}
}

func Index(w http.ResponseWriter, r *http.Request) {
	if !cas.IsAuthenticated(r) {
		cas.RedirectToLogin(w, r)
		return
	}
	db := dbOpen("talks")
	rows, err := db.Query("SELECT * FROM scagnt ORDER BY id DESC")
	if err != nil {
		log.Fatal(err)
	}
	talks := rowsToTalkSlice(rows)
	sort.Slice(talks, func(i, j int) bool {
		return talks[j].Date.Before(talks[i].Date)
	})
	tmpl.ExecuteTemplate(w, "Index", talks)
	defer db.Close()
}

func Show(w http.ResponseWriter, r *http.Request) {
	db := dbOpen("talks")
	nId := r.URL.Query().Get("id")
	rows, err := db.Query("SELECT * FROM scagnt WHERE id=?", nId)
	if err != nil {
		log.Fatal(err)
	}
	talk := rowsToTalkSlice(rows)[0]
	tmpl.ExecuteTemplate(w, "Show", talk)
	defer db.Close()
}

func New(w http.ResponseWriter, r *http.Request) {
	WriteToHTML()
	tmpl.ExecuteTemplate(w, "New", nil)
}

func Edit(w http.ResponseWriter, r *http.Request) {
	db := dbOpen("talks")
	nId := r.URL.Query().Get("id")
	rows, err := db.Query("SELECT * FROM scagnt WHERE id=?", nId)
	if err != nil {
		log.Fatal(err)
	}
	talk := rowsToTalkSlice(rows)[0]
	tmpl.ExecuteTemplate(w, "Edit", talk)
	defer db.Close()
}

func formToTalk(r *http.Request) (talk Talk) {
	talk = Talk{}
	talk.Abstract = r.FormValue("abstract")
	talk.Affiliation = r.FormValue("speaker_affiliation")
	month := r.FormValue("month")
	day := r.FormValue("day")
	if len(day) == 1 {
		day = "0" + day
	}
	year := r.FormValue("year")
	talk.Date_string = year + "-" + month + "-" + day
	talk.Location = r.FormValue("location")
	talk.Recording_url = PrependHTTP(r.FormValue("recording_url"))
	talk.Host = r.FormValue("host")
	talk.Title = r.FormValue("title")
	if talk.Title == "" {
		talk.Title = "TBA"
	}
	talk.Abstract = r.FormValue("abstract")
	talk.Speaker_first = r.FormValue("speaker_first")
	if talk.Speaker_first == "" {
		talk.Speaker_first = "Reserved"
	}
	talk.Speaker_last = r.FormValue("speaker_last")
	talk.Time_string = r.FormValue("time")
	talk.Speaker_url = PrependHTTP(r.FormValue("speaker_url"))
	talk.Vid_conf_url = PrependHTTP(r.FormValue("vid_conf_url"))
	talk.Vid_conf_pw = r.FormValue("vid_conf_pw")
	return talk
}

func Insert(w http.ResponseWriter, r *http.Request) {
	db := dbOpen("talks")
	if r.Method == "POST" {
		talk := formToTalk(r)
		insForm, err := db.Prepare("INSERT INTO scagnt (event_date, time, speaker_first, speaker_last, speaker_url, speaker_affiliation, title, abstract, vid_conf_url, vid_conf_pw, recording_url, host, location) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?)")
		if err != nil {
			log.Fatal(err)
		}
		insForm.Exec(talk.Date_string, talk.Time_string, talk.Speaker_first, talk.Speaker_last, talk.Speaker_url, talk.Affiliation, talk.Title, talk.Abstract, talk.Vid_conf_url, talk.Vid_conf_pw, talk.Recording_url, talk.Host, talk.Location)
		log.Println("INSERT: Name: " + talk.Speaker_first + " " + talk.Speaker_last + " | Title: " + talk.Title)
	}
	defer db.Close()
	WriteToHTML()
	tmpl.ExecuteTemplate(w, "Insert", nil)
}

func Update(w http.ResponseWriter, r *http.Request) {
	db := dbOpen("talks")
	if r.Method == "POST" {
		id := r.FormValue("uid")
		talk := formToTalk(r)
		insForm, err := db.Prepare("UPDATE scagnt SET event_date=?, time=?, speaker_first=? ,speaker_last=?, speaker_url=?, speaker_affiliation=?, title=?, abstract=?, vid_conf_url=?, vid_conf_pw=?, recording_url=?, host=?, location=? WHERE id=?")
		if err != nil {
			log.Fatal(err)
		}
		insForm.Exec(talk.Date_string, talk.Time_string, talk.Speaker_first, talk.Speaker_last, talk.Speaker_url, talk.Affiliation, talk.Title, talk.Abstract, talk.Vid_conf_url, talk.Vid_conf_pw, talk.Recording_url, talk.Host, talk.Location, id)
		log.Println("UPDATE: Name: " + talk.Speaker_first + " " + talk.Speaker_last + " | Title: " + talk.Title)
	}
	defer db.Close()
	WriteToHTML()
	tmpl.ExecuteTemplate(w, "Update", nil)
}

func Delete(w http.ResponseWriter, r *http.Request) {
	db := dbOpen("talks")
	nId := r.URL.Query().Get("id")
	rows, err := db.Query("SELECT * FROM scagnt WHERE id=?", nId)
	if err != nil {
		log.Fatal(err)
	}
	trash := rowsToTalkSlice(rows)[0]
	log.Println("Pending deletion: " + trash.Speaker_first + " " + trash.Speaker_last + " | " + trash.Title)
	tmpl.ExecuteTemplate(w, "Delete", trash)
	defer db.Close()
}

func ConfirmDelete(w http.ResponseWriter, r *http.Request) {
	// bin := dbOpen("trash")
	// statement, err := bin.Prepare("CREATE TABLE scagnt (id INTEGER PRIMARY KEY, event_date TEXT, time TEXT, speaker_first TEXT, speaker_last TEXT, speaker_url TEXT, speaker_affiliation TEXT, title TEXT, abstract TEXT, vid_conf_url TEXT, vid_conf_pw TEXT, recording_url TEXT)")
	// if err != nil {
	//     log.Fatal(err)
	// }
	// statement.Exec()
	// // insForm, err := bin.Prepare("INSERT INTO talks(month,day,year,hour,minute,speaker,title) VALUES(?,?,?,?,?,?,?)")
	// // if err != nil {
	// //     log.Fatal(err)
	// // }
	// // insForm.Exec(trash.Month, trash.Day, trash.year, 0, 0, trash.speaker, trash.title)
	// // log.Println("INSERT: Name: " + trash.speaker + " | Title: " + trash.title)

	db := dbOpen("talks")
	nId := r.URL.Query().Get("id")
	delForm, err := db.Prepare("DELETE FROM scagnt WHERE id=?")
	if err != nil {
		log.Fatal(err)
	}
	delForm.Exec(nId)
	log.Println("DELETE")
	defer db.Close()
	WriteToHTML()
	tmpl.ExecuteTemplate(w, "ConfirmDelete", nil)
}

func convert_date(exp_date string) time.Time {
	dt_typed, err := time.Parse(layout, exp_date)
	if err != nil {
		log.Fatal(err)
	}
	return dt_typed
}

func check_token(token string) bool {
	var result bool
	if len(token) != 24 {
		result = false
	} else {
		var token2, exp_date string
		db := dbOpen("tokens")
		defer db.Close()
		err := db.QueryRow("SELECT * FROM tokens WHERE token=?", token).Scan(&token2, &exp_date)
		if err == sql.ErrNoRows {
			result = false
		} else if err != nil {
			log.Fatal(err)
		} else {
			exp_time := convert_date(exp_date)
			if exp_time.After(time.Now()) {
				result = false
			}
			result = true
		}
	}
	return result
}

func auth(handler http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// user_token := r.Header.Get("scagnt-authorization")
		var user_token string
		cookie, err := r.Cookie("scagnt")
		if err == nil {
			user_token = cookie.Value
		} else {
			user_token = ""
		}
		if check_token(user_token) {
			handler.ServeHTTP(w, r)
		} else {
			http.Redirect(w, r, "/login", http.StatusUnauthorized)
		}
	})
}

const charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

func StringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func String(length int) string {
	return StringWithCharset(length, charset)
}

func Login(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "Login", nil)
}

func Attempt(w http.ResponseWriter, r *http.Request) {
	// user_token := r.Header.Get("scagnt-authorization")
	var user_token string
	cookie, err := r.Cookie("scagnt")
	if err == nil {
		user_token = cookie.Value
	} else {
		user_token = ""
	}
	if check_token(user_token) {
		tmpl.ExecuteTemplate(w, "Success", nil)
	} else {
		var authenticate bool
		users := dbOpen("users")
		var user, pw string
		defer users.Close()
		if r.Method == "POST" {
			current_user := r.FormValue("user")
			current_pw := r.FormValue("password")
			log.Println("Authentication attempt:" + current_user + " " + current_pw)
			err := users.QueryRow("SELECT * FROM users WHERE user=?", current_user).Scan(&user, &pw)
			if err == sql.ErrNoRows {
				fmt.Println("No user here")
				authenticate = false
			} else if err != nil {
				log.Fatal(err)
			} else {
				if current_pw == pw {
					authenticate = true
				} else {
					log.Println(current_pw)
					log.Println(pw)
					log.Println("Passwords don't match")
					authenticate = false
				}
			}
		}
		if authenticate {
			tokens := dbOpen("tokens")
			defer tokens.Close()
			token := String(24)
			expire := time.Now().AddDate(0, 0, 1)
			date := expire.Format(layout)
			insForm, err := tokens.Prepare("INSERT INTO tokens(token,exp_date) VALUES(?,?)")
			if err != nil {
				log.Fatal(err)
			}
			insForm.Exec(token, date)
			cookie := &http.Cookie{
				Name:  "scagnt",
				Value: token,
			}
			http.SetCookie(w, cookie)
			tmpl.ExecuteTemplate(w, "Success", nil)
		} else {
			http.Redirect(w, r, "/", http.StatusSeeOther)
		}
	}
}

func setLogFile() {
	file, err := os.OpenFile("logs", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(file)
}

var casURL = "https://casserver.herokuapp.com/cas"

// var casURL = "https://cas-qa.auth.sc.edu/cas"

func main() {

	setLogFile()
	pathToBinary = "/Users/matt/GitHub/server/"
	setTermsYearsRepoPath()

	url, _ := url.Parse(casURL)
	client := cas.NewClient(&cas.Options{
		URL: url,
	})

	fmt.Println("Server started")
	log.Println("Server started on: http://localhost" + port)
	http.Handle("/", client.HandleFunc(Index))
	http.HandleFunc("/login", Login)
	http.HandleFunc("/attempt", Attempt)
	http.HandleFunc("/show", auth(Show))
	http.HandleFunc("/new", auth(New))
	http.HandleFunc("/edit", auth(Edit))
	http.HandleFunc("/insert", auth(Insert))
	http.HandleFunc("/update", auth(Update))
	http.HandleFunc("/delete", auth(Delete))
	http.HandleFunc("/confirmdeletionbesure", ConfirmDelete)

	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatal(err)
	}
}
