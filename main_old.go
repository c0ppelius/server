// package main

// import (
//     "database/sql"
//     "log"
//     // "os"
//     "fmt"
//     // "strconv"
//     "html/template"
//     "net/http"

//     _ "github.com/mattn/go-sqlite3"
// )

// type Talk struct {
//     Speaker  string
//     Title    string 
// }

// func main() {
//     tmpl := template.Must(template.ParseFiles("display.html"))

//     database, err := sql.Open("sqlite3","./nraboy.db")
//     if err != nil {
//         log.Fatal(err)
//     }

//     defer database.Close() 

//     rows, err := database.Query("SELECT id, firstname, lastname FROM people")
//     if err != nil {
//         log.Fatal(err)
//     }

//     Talks := []Talk{}

//     var id int
//     var firstname string 
//     var lastname string

//     for rows.Next() {
//         rows.Scan(&id, &firstname, &lastname)
//         talk := Talk {
//             Speaker: firstname,
//             Title: lastname,
//         }
//         Talks = append(Talks,talk)
//     }

//     http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
//         tmpl.Execute(w, Talks)
//     })

//     fmt.Println("Listening at localhost:2000")
//     http.ListenAndServe(":2000", nil)
// }
