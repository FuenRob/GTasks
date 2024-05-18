package models

type User struct {
	ID       int
	Name     string
	Email    string
	Password string
	Role     string
}

type Project struct {
	ID          int
	Name        string
	Description string
	OwnerID     int
}

type Task struct {
	ID        int
	Name      string
	ProjectID int
	StateID   int
}

type State struct {
	ID   int
	Name string
}
