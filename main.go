package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"text/template"
	"time"
    "strconv"
    "strings"
    "sort"
    "os"

	// "github.com/gobwas/glob/util/strings"
	_ "github.com/mattn/go-sqlite3"
)

type Talk struct {
    Id              int
    Date            time.Time 
    Upcoming        bool
    Year            string
    Month           string
    Month_name      string
    Day             string
    Date_string     string
    Time_string     string
    Speaker_first   string
    Speaker_last    string
    Speaker_url     string
    Affiliation     string
    Title           string
    Abstract        string
    Vid_conf_url    string
    Vid_conf_pw     string
    Recording_url   string
}

func dbOpen(file string) (db *sql.DB) {
    dbDriver := "sqlite3"
    db, err := sql.Open(dbDriver,prefix+file+".db")
    if err != nil {
        log.Fatal(err)
    }
    return db
}

// const prefix = "/home/pi/"
const prefix = ""

const port = ":2000"

const layout = "2006-01-02 15:04:05"

var tmpl = template.Must(template.ParseGlob(prefix+"forms/*"))

var trash = Talk{}

func PrependHTTP (url string) string {
    if len(url) == 0 {
        return url 
    } else if (url[0:7] != "http://") &&ca (url[0:8] != "https://") {
        return "http://"+url 
    } else {
        return url 
    }
}

func Index(w http.ResponseWriter, r *http.Request) {
    db := dbOpen("talks")
    rows, err := db.Query("SELECT * FROM scagnt ORDER BY id DESC")
    if err != nil {
        log.Fatal(err)
    }
    talk := Talk{}
    talks := []Talk{}
    for rows.Next() {
        var id int
        var event_date, event_time, speaker_first, speaker_last, speaker_url, speaker_affiliation, title, abstract, vid_conf_url, vid_conf_pw, recording_url string
        err = rows.Scan(&id, &event_date, &event_time, &speaker_first, &speaker_last, &speaker_url, &speaker_affiliation, &title, &abstract, &vid_conf_url, &vid_conf_pw, &recording_url)
        if err != nil {
            log.Fatal(err)
        }
        talk.Id = id
        s := strings.Split(event_time,"-")
        if len(s[0]) == 4 {
            s[0] = "0"+s[0]
        }
        event_time = s[0]
        string_time := event_date+"T"+event_time+":00.000Z"
        converted_time, err := time.Parse("2006-01-02T15:04:05.000Z",string_time)
        if err != nil {
            log.Fatal(err)
        }
        talk.Date = converted_time
        talk.Upcoming = converted_time.After(time.Now())
        talk.Date_string = talk.Date.Format("January 02 2006") 
        talk.Year = strconv.Itoa(converted_time.Year())
        month_num := strconv.Itoa(int(converted_time.Month()))
        if len(month_num) == 1 {
            talk.Month = "0"+month_num
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
        talks = append(talks, talk)
    }
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
    talk := Talk{}
    for rows.Next() {
        var id int
        var event_date, event_time, speaker_first, speaker_last, speaker_url, speaker_affiliation, title, abstract, vid_conf_url, vid_conf_pw, recording_url string
        err = rows.Scan(&id, &event_date, &event_time, &speaker_first, &speaker_last, &speaker_url, &speaker_affiliation, &title, &abstract, &vid_conf_url, &vid_conf_pw, &recording_url)
        if err != nil {
            log.Fatal(err)
        }
        talk.Id = id
        s := strings.Split(event_time,"-")
        if len(s[0]) == 4 {
            s[0] = "0"+s[0]
        }
        event_time = s[0]
        string_time := event_date+"T"+event_time+":00.000Z"
        converted_time, err := time.Parse("2006-01-02T15:04:05.000Z",string_time)
        if err != nil {
            log.Fatal(err)
        }
        talk.Date = converted_time
        talk.Date_string = talk.Date.Format("January 02 2006")
        talk.Year = strconv.Itoa(converted_time.Year())
        month_num := strconv.Itoa(int(converted_time.Month()))
        if len(month_num) == 1 {
            talk.Month = "0"+month_num
        } else {
            talk.Month = strconv.Itoa(int(converted_time.Month()))
        }
        talk.Month_name = talk.Date.Format("January")
        talk.Day = strconv.Itoa(converted_time.Day())
        talk.Time_string = converted_time.Format("15:04")
        talk.Speaker_first = speaker_first
        talk.Speaker_last = speaker_last
        talk.Affiliation = speaker_affiliation
        talk.Title = title
        talk.Abstract = abstract
        talk.Speaker_url = speaker_url
        talk.Vid_conf_url = vid_conf_url
        talk.Vid_conf_pw = vid_conf_pw
        talk.Recording_url = recording_url
    }
    tmpl.ExecuteTemplate(w, "Show", talk)
    defer db.Close()
}

func New(w http.ResponseWriter, r *http.Request) {
    tmpl.ExecuteTemplate(w, "New", nil)
}

func Edit(w http.ResponseWriter, r *http.Request) {
    db := dbOpen("talks")
    nId := r.URL.Query().Get("id")
    rows, err := db.Query("SELECT * FROM scagnt WHERE id=?", nId)
    if err != nil {
        log.Fatal(err)
    }
    talk := Talk{}
    for rows.Next() {
        var id int
        var event_date, event_time, speaker_first, speaker_last, speaker_url, speaker_affiliation, title, abstract, vid_conf_url, vid_conf_pw, recording_url string
        err = rows.Scan(&id, &event_date, &event_time, &speaker_first, &speaker_last, &speaker_url, &speaker_affiliation, &title, &abstract, &vid_conf_url, &vid_conf_pw, &recording_url)
        if err != nil {
            log.Fatal(err)
        }
        talk.Id = id
        s := strings.Split(event_time,"-")
        if len(s[0]) == 4 {
            s[0] = "0"+s[0]
        }
        event_time = s[0]
        string_time := event_date+"T"+event_time+":00.000Z"
        converted_time, err := time.Parse("2006-01-02T15:04:05.000Z",string_time)
        if err != nil {
            log.Fatal(err)
        }
        talk.Date = converted_time
        talk.Date_string = talk.Date.Format("January 02 2006")
        talk.Year = strconv.Itoa(converted_time.Year())
        month_num := strconv.Itoa(int(converted_time.Month()))
        if len(month_num) == 1 {
            talk.Month = "0"+month_num
        } else {
            talk.Month = strconv.Itoa(int(converted_time.Month()))
        }
        talk.Month_name = talk.Date.Format("January")
        talk.Day = strconv.Itoa(converted_time.Day())
        talk.Time_string = converted_time.Format("15:04")
        talk.Speaker_first = speaker_first
        talk.Speaker_last = speaker_last
        talk.Affiliation = speaker_affiliation
        talk.Title = title
        talk.Abstract = abstract
        talk.Speaker_url = speaker_url
        talk.Vid_conf_url = vid_conf_url
        talk.Vid_conf_pw = vid_conf_pw
        talk.Recording_url = recording_url
    }
    tmpl.ExecuteTemplate(w, "Edit", talk)
    defer db.Close()
}

func Insert(w http.ResponseWriter, r *http.Request) {
    db := dbOpen("talks")
    if r.Method == "POST" {
        speaker_first := r.FormValue("speaker_first")
        if speaker_first == "" {
            speaker_first = "Reserved"
        }
        speaker_last := r.FormValue("speaker_last")
        event_time := r.FormValue("time")
        speaker_url := PrependHTTP(r.FormValue("speaker_url"))
        speaker_affiliation := r.FormValue("speaker_affiliation")
        vid_conf_url := PrependHTTP(r.FormValue("vid_conf_url"))
        vid_conf_pw := r.FormValue("vid_conf_pw")
        recording_url := PrependHTTP(r.FormValue("recording_url"))
        title := r.FormValue("title")
        abstract := r.FormValue("abstract")
        if title == "" {
            title = "TBA"
        }
        month := r.FormValue("month")
        day := r.FormValue("day")
        if len(day) == 1 {
            day = "0"+day
        }
        year := r.FormValue("year")
        event_date := year+"-"+month+"-"+day
        insForm, err := db.Prepare("INSERT INTO scagnt (event_date, time, speaker_first, speaker_last, speaker_url, speaker_affiliation, title, abstract, vid_conf_url, vid_conf_pw, recording_url) VALUES(?,?,?,?,?,?,?,?,?,?,?)")
        if err != nil {
            log.Fatal(err)
        }
        insForm.Exec(event_date,event_time,speaker_first,speaker_last,speaker_url,speaker_affiliation,title,abstract,vid_conf_url, vid_conf_pw, recording_url)
        log.Println("INSERT: Name: " + speaker_first + " " + speaker_last + " | Title: " + title)
    }
    defer db.Close()
    tmpl.ExecuteTemplate(w,"Insert",nil)
}

func Update(w http.ResponseWriter, r *http.Request) {
    db := dbOpen("talks")
    if r.Method == "POST" {
        id := r.FormValue("uid")
        speaker_first := r.FormValue("speaker_first")
        if speaker_first == "" {
            speaker_first = "Reserved"
        }
        speaker_last := r.FormValue("speaker_last")
        event_time := r.FormValue("time")
        speaker_url := PrependHTTP(r.FormValue("speaker_url"))
        speaker_affiliation := r.FormValue("speaker_affiliation")
        vid_conf_url := PrependHTTP(r.FormValue("vid_conf_url"))
        vid_conf_pw := r.FormValue("vid_conf_pw")
        recording_url := PrependHTTP(r.FormValue("recording_url"))
        title := r.FormValue("title")
        abstract := r.FormValue("abstract")
        if title == "" {
            title = "TBA"
        }
        month := r.FormValue("month")
        day := r.FormValue("day")
        if len(day) == 1 {
            day = "0"+day
        }
        year := r.FormValue("year")
        event_date := year+"-"+month+"-"+day
        insForm, err := db.Prepare("UPDATE scagnt SET event_date=?, time=?, speaker_first=? ,speaker_last=?, speaker_url=?, speaker_affiliation=?, title=?, abstract=?, vid_conf_url=?, vid_conf_pw=?, recording_url=? WHERE id=?")
        if err != nil {
            log.Fatal(err)
        }
        insForm.Exec(event_date, event_time, speaker_first, speaker_last, speaker_url, speaker_affiliation, title, abstract, vid_conf_url, vid_conf_pw, recording_url, id)
        log.Println("UPDATE: Name: " + speaker_first + " " + speaker_last + " | Title: " + title)
    }
    defer db.Close()
    tmpl.ExecuteTemplate(w,"Update",nil)
}

func Delete(w http.ResponseWriter, r *http.Request) {
    db := dbOpen("talks")
    nId := r.URL.Query().Get("id")
    rows, err := db.Query("SELECT * FROM scagnt WHERE id=?", nId)
    if err != nil {
        log.Fatal(err)
    }
    for rows.Next() {
        var id int
        var event_date, event_time, speaker_first, speaker_last, speaker_url, speaker_affiliation, title, abstract, vid_conf_url, vid_conf_pw, recording_url string
        err = rows.Scan(&id, &event_date, &event_time, &speaker_first, &speaker_last, &speaker_url, &speaker_affiliation, &title, &abstract, &vid_conf_url, &vid_conf_pw, &recording_url)
        if err != nil {
            log.Fatal(err)
        }
        trash.Id = id
        s := strings.Split(event_time,"-")
        if len(s[0]) == 4 {
            s[0] = "0"+s[0]
        }
        event_time = s[0]
        string_time := event_date+"T"+event_time+":00.000Z"
        converted_time, err := time.Parse("2006-01-02T15:04:05.000Z",string_time)
        if err != nil {
            log.Fatal(err)
        }
        trash.Date = converted_time
        trash.Date_string = trash.Date.Format("January 02 2006")
        trash.Year = strconv.Itoa(converted_time.Year())
        month_num := strconv.Itoa(int(converted_time.Month()))
        if len(month_num) == 1 {
            trash.Month = "0"+month_num
        } else {
            trash.Month = strconv.Itoa(int(converted_time.Month()))
        }
        trash.Month_name = trash.Date.Format("January")
        trash.Day = strconv.Itoa(converted_time.Day())
        trash.Time_string = converted_time.Format("15:04")
        trash.Speaker_first = speaker_first
        trash.Speaker_last = speaker_last
        trash.Affiliation = speaker_affiliation
        trash.Title = title
        trash.Abstract = abstract
        trash.Speaker_url = speaker_url
        trash.Vid_conf_url = vid_conf_url
        trash.Vid_conf_pw = vid_conf_pw
        trash.Recording_url = recording_url
    }
    log.Println("Pending deletion: "+trash.Speaker_first+" "+trash.Speaker_last+" | "+trash.Title)
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
    tmpl.ExecuteTemplate(w,"ConfirmDelete",nil)
}

func convert_date(exp_date string) time.Time {
	dt_typed, err := time.Parse(layout, exp_date)
    if err != nil{
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
        err := db.QueryRow("SELECT * FROM tokens WHERE token=?", token).Scan(&token2,&exp_date)
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
            handler.ServeHTTP(w,r)
        } else {
            http.Redirect(w,r,"/login", http.StatusUnauthorized)
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
    tmpl.ExecuteTemplate(w,"Login",nil)
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
        tmpl.ExecuteTemplate(w,"Success",nil)
    } else {
        var authenticate bool
        users := dbOpen("users")
        var user, pw string 
        defer users.Close()
        if r.Method == "POST" {
            current_user := r.FormValue("user")
            current_pw := r.FormValue("password")
            log.Println("Authentication attempt:"+current_user+" "+current_pw)
            err := users.QueryRow("SELECT * FROM users WHERE user=?", current_user).Scan(&user,&pw)
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
            expire := time.Now().AddDate(0,0,1) 
            date := expire.Format(layout)
            insForm, err := tokens.Prepare("INSERT INTO tokens(token,exp_date) VALUES(?,?)")
            if err != nil {
                log.Fatal(err)
            }
            insForm.Exec(token,date)
            cookie := &http.Cookie {
                Name: "scagnt",
                Value: token,
            }
            http.SetCookie(w,cookie)
            tmpl.ExecuteTemplate(w,"Success",nil)
        } else {
            http.Redirect(w, r, "/", http.StatusSeeOther)
        }
    }
}

func main() {
    file, err := os.OpenFile("logs", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
    if err != nil {
        log.Fatal(err)
    }
    log.SetOutput(file)
    fmt.Println("Server started")
    log.Println("Server started on: http://localhost"+port)
    http.HandleFunc("/", auth(Index))
    http.HandleFunc("/login", Login)
    http.HandleFunc("/attempt", Attempt)
    http.HandleFunc("/show", auth(Show))
    http.HandleFunc("/new", auth(New))
    http.HandleFunc("/edit", auth(Edit))
    http.HandleFunc("/insert", auth(Insert))
    http.HandleFunc("/update", auth(Update))
    http.HandleFunc("/delete", auth(Delete))
    http.HandleFunc("/confirmdeletionbesure", ConfirmDelete)
    err = http.ListenAndServe(port, nil)
    if err != nil {
        log.Fatal(err)
    }
}