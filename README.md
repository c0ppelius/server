# A backend seminar server

A Go server allowing multiple organizers to schedule talks. 

Features:
- implements basic CRUD 
- very, very, ..., very basic authentication 
- generates HTML and automates committing and pushing it to a git repository serving the static files (where your actual seminar webpage lives)
- config file allows for updating the semester and year without restarting the server
- data is stored SQLite files which must be supplied for deployment 

Beyond the standard library, the server uses the [SQLite driver from ModernC](https://pkg.go.dev/modernc.org/sqlite]).

To do:
- integrate with OAuth2 for better authentication 
- clean up and separate the code into some sub-packages 