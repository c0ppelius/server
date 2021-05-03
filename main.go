package main

import (
    "database/sql"
    "fmt"
    "log"
    "net/http"
    "text/template"
    "time"
    "math/rand"

    _ "github.com/mattn/go-sqlite3"
)

type Talk struct {
    Id      int
    Date    string 
    Speaker string
    Title   string
}

func dbOpen(file string) (db *sql.DB) {
    dbDriver := "sqlite3"
    db, err := sql.Open(dbDriver,file+".db")
    if err != nil {
        log.Fatal(err)
    }
    return db
}

const layout = "2006-01-02 15:04:05"

var tmpl = template.Must(template.ParseGlob("forms/*"))

var trash = Talk{}

func Index(w http.ResponseWriter, r *http.Request) {
    db := dbOpen("talks")
    rows, err := db.Query("SELECT * FROM talks ORDER BY id DESC")
    if err != nil {
        log.Fatal(err)
    }
    talk := Talk{}
    talks := []Talk{}
    for rows.Next() {
        var id int
        var month, day, year, hour, minute int 
        var speaker, title string
        err = rows.Scan(&id, &month, &day, &year, &hour, &minute, &speaker, &title)
        if err != nil {
            log.Fatal(err)
        }
        talk.Id = id
        talk.Date = time.Date(year, time.Month(month), day, hour, minute, 0, 0, time.UTC).Format("Jan 02 2006")
        talk.Speaker = speaker
        talk.Title = title
        talks = append(talks, talk)
    }
    tmpl.ExecuteTemplate(w, "Index", talks)
    defer db.Close()
}

func Show(w http.ResponseWriter, r *http.Request) {
    db := dbOpen("talks")
    nId := r.URL.Query().Get("id")
    rows, err := db.Query("SELECT * FROM talks WHERE id=?", nId)
    if err != nil {
        log.Fatal(err)
    }
    talk := Talk{}
    for rows.Next() {
        var id int
        var month, day, year, hour, minute int 
        var speaker, title string
        err = rows.Scan(&id, &month, &day, &year, &hour, &minute, &speaker, &title)
        if err != nil {
            log.Fatal(err)
        }
        talk.Id = id
        talk.Date = time.Date(year, time.Month(month), day, hour, minute, 0, 0, time.UTC).Format("Jan 02 2006")
        talk.Speaker = speaker
        talk.Title = title
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
    rows, err := db.Query("SELECT * FROM talks WHERE id=?", nId)
    if err != nil {
        log.Fatal(err)
    }
    talk := Talk{}
    for rows.Next() {
        var id int
        var month, day, year, hour, minute int 
        var speaker, title string
        err = rows.Scan(&id, &month, &day, &year, &hour, &minute, &speaker, &title)
        if err != nil {
            log.Fatal(err)
        }
        talk.Id = id
        talk.Date = time.Date(year, time.Month(month), day, hour, minute, 0, 0, time.UTC).Format("Jan 02 2006")
        talk.Speaker = speaker
        talk.Title = title
    }
    tmpl.ExecuteTemplate(w, "Edit", talk)
    defer db.Close()
}

func Insert(w http.ResponseWriter, r *http.Request) {
    db := dbOpen("talks")
    if r.Method == "POST" {
        speaker := r.FormValue("speaker")
        if speaker == "" {
            speaker = "Reserved"
        }
        title := r.FormValue("title")
        if title == "" {
            title = "TBA"
        }
        month := r.FormValue("month")
        day := r.FormValue("day")
        year := r.FormValue("year")
        insForm, err := db.Prepare("INSERT INTO talks(month,day,year,hour,minute,speaker,title) VALUES(?,?,?,?,?,?,?)")
        if err != nil {
            log.Fatal(err)
        }
        insForm.Exec(month, day, year, 0, 0, speaker, title)
        log.Println("INSERT: Name: " + speaker + " | Title: " + title)
    }
    defer db.Close()
    tmpl.ExecuteTemplate(w,"Insert",nil)
}

func Update(w http.ResponseWriter, r *http.Request) {
    db := dbOpen("talks")
    if r.Method == "POST" {
        speaker := r.FormValue("speaker")
        title := r.FormValue("title")
        month := r.FormValue("month")
        day := r.FormValue("day")
        year := r.FormValue("year")
        id := r.FormValue("uid")
        insForm, err := db.Prepare("UPDATE talks SET month=?, day=?, year=?, hour=?, minute=?, speaker=?, title=? WHERE id=?")
        if err != nil {
            log.Fatal(err)
        }
        insForm.Exec(month, day, year, 0, 0, speaker, title, id)
        log.Println("UPDATE: Name: " + speaker + " | Title: " + title)
    }
    defer db.Close()
    tmpl.ExecuteTemplate(w,"Update",nil)
}

func Delete(w http.ResponseWriter, r *http.Request) {
    db := dbOpen("talks")
    nId := r.URL.Query().Get("id")
    rows, err := db.Query("SELECT * FROM talks WHERE id=?", nId)
    if err != nil {
        log.Fatal(err)
    }
    for rows.Next() {
        var id int
        var month, day, year, hour, minute int 
        var speaker, title string
        err = rows.Scan(&id, &month, &day, &year, &hour, &minute, &speaker, &title)
        if err != nil {
            log.Fatal(err)
        }
        trash.Id = id
        trash.Date = time.Date(year, time.Month(month), day, hour, minute, 0, 0, time.UTC).Format("Jan 02 2006")
        trash.Speaker = speaker
        trash.Title = title
    }
    tmpl.ExecuteTemplate(w, "Delete", trash)
    defer db.Close()
}

func ConfirmDelete(w http.ResponseWriter, r *http.Request) {
    bin := dbOpen("trash")
    statement, err := bin.Prepare("CREATE TABLE IF NOT EXISTS talks (id INTEGER PRIMARY KEY, month INTEGER, day INTEGER, year INTEGER, hour INTEGER, minute INTEGER, speaker TEXT, title TEXT)")
    if err != nil {
        log.Fatal(err)
    }
    statement.Exec()
    // insForm, err := bin.Prepare("INSERT INTO talks(month,day,year,hour,minute,speaker,title) VALUES(?,?,?,?,?,?,?)")
    // if err != nil {
    //     log.Fatal(err)
    // }
    // insForm.Exec(trash.Month, trash.Day, trash.year, 0, 0, trash.speaker, trash.title)
    // log.Println("INSERT: Name: " + trash.speaker + " | Title: " + trash.title)

    db := dbOpen("talks")
    nId := r.URL.Query().Get("id")
    delForm, err := db.Prepare("DELETE FROM talks WHERE id=?")
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
        user_token := r.Header.Get("scagnt-authorization")
        fmt.Println(user_token)
        fmt.Println(r)
        if check_token(user_token) {
            handler.ServeHTTP(w,r)
        } else {
            fmt.Println("Uh oh")
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
    user_token := r.Header.Get("scagnt-authorization")
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
                    fmt.Println(current_pw)
                    fmt.Println(pw)
                    fmt.Println("Passwords don't match")
                    authenticate = false
                }
            }
        }
        if authenticate {
            fmt.Println("Boom.")
            tokens := dbOpen("tokens")
            defer tokens.Close()
            token := String(24)
            date := (time.Now().AddDate(0,0,1)).Format(layout)
            insForm, err := tokens.Prepare("INSERT INTO tokens(token,exp_date) VALUES(?,?)")
            if err != nil {
                log.Fatal(err)
            }
            insForm.Exec(token,date)
            fmt.Println(token)
            r.Header.Add("scagnt-authorization",token)
            fmt.Println(r.Header.Get("scagnt-authorization"))
            tmpl.ExecuteTemplate(w,"Success",nil)
        } else {
            http.Redirect(w, r, "/", http.StatusSeeOther)
        }
    }
}

func main() {
    port := ":2000"
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
    err := http.ListenAndServe(port, nil)
    if err != nil {
        fmt.Println(err)
    }
}