package main 

import (
    "database/sql"
    "log"
    // "os"
    "fmt"
    // "strconv"
    "html/template"
    "net/http"

    _ "github.com/mattn/go-sqlite3"
)

type Talk struct {
    Speaker  string
    Title    string 
}

func main() {
	tmpl := template.Must(template.ParseFiles("form.html"))

	database, err := sql.Open("sqlite3","./talks.db")
    if err != nil {
        log.Fatal(err)
    }

    defer database.Close() 

	rows, err := database.Query("SELECT id, speaker, title FROM talk ")
	if err != nil {
		log.Fatal(err)
	}
	var id int
	var speaker string
	var title string
	for rows.Next() {
		rows.Scan(&id, &speaker, &title)
		fmt.Println( speaker + ": " + title)
	}


	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost {
			tmpl.Execute(w, nil)
			return
		}
	
		talk := Talk{
			Speaker:	r.FormValue("speaker"),
			Title:	r.FormValue("title"),
		}
	
		database, err := sql.Open("sqlite3","./talks.db")
		if err != nil {
			log.Fatal(err)
		}
	
		defer database.Close()
	
		statement, err := database.Prepare("CREATE TABLE IF NOT EXISTS talk (id INTEGER PRIMARY KEY, speaker TEXT, title TEXT)")
		if err != nil {
			log.Fatal(err)
		}
		statement.Exec()
	
		statement, err = database.Prepare("INSERT INTO talk (speaker, title) VALUES (?, ?)")
		if err != nil {
			log.Fatal(err)
		}
		statement.Exec(talk.Speaker, talk.Title)
		
		tmpl.Execute(w, struct{ Success bool }{true})
    })

	fmt.Println("Listening at localhost:2000")
	http.ListenAndServe(":2000", nil)
}