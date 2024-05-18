package routes

import (
	"gtasks/internal/handlers"

	"github.com/gorilla/mux"
)

func InitRoutes() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/", handlers.IndexHandler).Methods("GET")
	router.HandleFunc("/login", handlers.LoginHandler).Methods("GET", "POST")
	router.HandleFunc("/projects", handlers.ProjectsHandler).Methods("GET", "POST")
	router.HandleFunc("/projects/{projectID}/users", handlers.UsersHandler).Methods("GET", "POST")
	router.HandleFunc("/tasks", handlers.TasksHandler).Methods("GET", "POST")
	return router
}
