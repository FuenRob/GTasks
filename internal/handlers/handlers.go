package handlers

import (
	"database/sql"
	"gtasks/internal/database"
	"gtasks/internal/models"
	"html/template"
	"net/http"

	"github.com/gorilla/mux"
)

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	files := []string{
		"internal/templates/" + tmpl + ".tmpl",
		"internal/templates/layouts/base.tmpl",
		"internal/templates/partials/header.tmpl",
	}

	t, err := template.ParseFiles(files...)
	if err != nil {
		http.Error(w, "Error parsing templates: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = t.ExecuteTemplate(w, "base.tmpl", data)
	if err != nil {
		http.Error(w, "Error executing template: "+err.Error(), http.StatusInternalServerError)
	}
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "pages/index", nil)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		email := r.FormValue("email")
		password := r.FormValue("password")

		// Authenticate user (simplified example)
		var user models.User
		err := database.DB.QueryRow("SELECT id, name, email, password, role FROM users WHERE email = ?", email).Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.Role)
		if err != nil {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		// Check password (this example assumes plain text, use a proper hash comparison in a real app)
		if user.Password != password {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		// Set session or token (simplified)
		http.SetCookie(w, &http.Cookie{
			Name:  "user_id",
			Value: string(rune(user.ID)),
		})

		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else {
		renderTemplate(w, "pages/login", nil)
	}
}

func ProjectsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		name := r.FormValue("name")
		description := r.FormValue("description")
		ownerID := 1 // In a real app, get this from session or context

		_, err := database.DB.Exec("INSERT INTO projects (name, description, owner_id) VALUES (?, ?, ?)", name, description, ownerID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	rows, err := database.DB.Query("SELECT id, name, description FROM projects WHERE owner_id = ?", 1) // Example: fetch projects for user 1
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var projects []models.Project
	for rows.Next() {
		var project models.Project
		err := rows.Scan(&project.ID, &project.Name, &project.Description)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		projects = append(projects, project)
	}

	data := struct {
		Projects []models.Project
	}{
		Projects: projects,
	}

	renderTemplate(w, "pages/project", data)
}

func TasksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		name := r.FormValue("name")
		projectID := r.FormValue("project_id")
		stateID := r.FormValue("state_id")

		_, err := database.DB.Exec("INSERT INTO tasks (name, project_id, state_id) VALUES (?, ?, ?)", name, projectID, stateID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	projects, err := fetchProjects()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	states, err := fetchStates()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tasks, err := fetchTasks()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Projects []models.Project
		States   []models.State
		Tasks    []struct {
			Name        string
			ProjectName string
			StateName   string
		}
	}{
		Projects: projects,
		States:   states,
		Tasks:    tasks,
	}

	renderTemplate(w, "pages/task", data)
}

func UsersHandler(w http.ResponseWriter, r *http.Request) {
	projectID := mux.Vars(r)["projectID"]

	if r.Method == http.MethodPost {
		email := r.FormValue("email")
		name := r.FormValue("name")
		role := r.FormValue("role")

		var user models.User
		err := database.DB.QueryRow("SELECT id, name, email, role FROM users WHERE email = ?", email).Scan(&user.ID, &user.Name, &user.Email, &user.Role)

		if err == sql.ErrNoRows {
			_, err := database.DB.Exec("INSERT INTO users (name, email, role) VALUES (?, ?, ?)", name, email, role)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			err = database.DB.QueryRow("SELECT id FROM users WHERE email = ?", email).Scan(&user.ID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		_, err = database.DB.Exec("INSERT INTO project_users (project_id, user_id) VALUES (?, ?)", projectID, user.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	rows, err := database.DB.Query("SELECT u.id, u.name, u.email, u.role FROM users u JOIN project_users pu ON u.id = pu.user_id WHERE pu.project_id = ?", projectID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.Role)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		users = append(users, user)
	}

	data := struct {
		ProjectID string
		Users     []models.User
	}{
		ProjectID: projectID,
		Users:     users,
	}

	renderTemplate(w, "pages/users", data)
}

func fetchProjects() ([]models.Project, error) {
	rows, err := database.DB.Query("SELECT id, name FROM projects")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []models.Project
	for rows.Next() {
		var project models.Project
		err := rows.Scan(&project.ID, &project.Name)
		if err != nil {
			return nil, err
		}
		projects = append(projects, project)
	}
	return projects, nil
}

func fetchStates() ([]models.State, error) {
	rows, err := database.DB.Query("SELECT id, name FROM states")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var states []models.State
	for rows.Next() {
		var state models.State
		err := rows.Scan(&state.ID, &state.Name)
		if err != nil {
			return nil, err
		}
		states = append(states, state)
	}
	return states, nil
}

func fetchTasks() ([]struct {
	Name        string
	ProjectName string
	StateName   string
}, error) {
	rows, err := database.DB.Query(`SELECT t.name, p.name AS project_name, s.name AS state_name 
                                     FROM tasks t 
                                     JOIN projects p ON t.project_id = p.id 
                                     JOIN states s ON t.state_id = s.id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []struct {
		Name        string
		ProjectName string
		StateName   string
	}
	for rows.Next() {
		var task struct {
			Name        string
			ProjectName string
			StateName   string
		}
		err := rows.Scan(&task.Name, &task.ProjectName, &task.StateName)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}
