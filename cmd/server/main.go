package main

import (
	"gtasks/internal/database"
	"gtasks/internal/routes"
	"net/http"
)

func main() {
	database.InitDB()
	router := routes.InitRoutes()

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.Handle("/", router)
	http.ListenAndServe(":8080", nil)
}
