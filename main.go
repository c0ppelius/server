package main

import (
    "database/sql"
    "fmt"
    "log"
    "net/http"
    "text/template"
    "time"

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
    statement, err := db.Prepare("CREATE TABLE IF NOT EXISTS talks (id INTEGER PRIMARY KEY, month INTEGER, day INTEGER, year INTEGER, hour INTEGER, minute INTEGER, speaker TEXT, title TEXT)")
		if err != nil {
			log.Fatal(err)
		}
    statement.Exec()
    return db
}

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
    http.Redirect(w, r, "/", 301  )
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
    http.Redirect(w, r, "/", 301)
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
    http.Redirect(w, r, "/", 301 )
}

func main() {
    port := ":8081"
    log.Println("Server started on: http://localhost"+port)
    http.HandleFunc("/", Index)
    http.HandleFunc("/show", Show)
    http.HandleFunc("/new", New)
    http.HandleFunc("/edit", Edit)
    http.HandleFunc("/insert", Insert)
    http.HandleFunc("/update", Update)
    http.HandleFunc("/delete", Delete)
    http.HandleFunc("/confirmdeletionbesure", ConfirmDelete)
    err := http.ListenAndServe(port, nil)
    if err != nil {
        fmt.Println(err)
    }
}