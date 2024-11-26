package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"regexp"
	"sbilis-todo-app/db"
	"sbilis-todo-app/models"
	"strconv"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

// Define a type for the key we'll use in the context
type contextKey string

const usernameKey contextKey = "username"

// Helper function to get the logged-in user from the request context
func getLoggedInUsername(r *http.Request) (string, error) {
	cookie, err := r.Cookie("session")
	if err != nil || cookie.Value == "" {
		log.Println("getLoggedInUsername: No session cookie found")
		return "", errors.New("no session")
	}

	var username string
	err = db.DB.QueryRow("SELECT username FROM users WHERE id = ?", cookie.Value).Scan(&username)
	if err != nil {
		log.Printf("getLoggedInUsername: User not found for ID %s", cookie.Value)
		return "", fmt.Errorf("user not found")
	}
	log.Printf("getLoggedInUsername: User %s found", username)
	return username, nil
}

// AuthMiddleware checks if the user is authenticated and extracts the username
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username, err := getLoggedInUsername(r)
		if err != nil {
			log.Println("AuthMiddleware: User not logged in, redirecting to /login")
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		log.Printf("AuthMiddleware: User %s is logged in", username)
		next.ServeHTTP(w, r)
	}
}

// Function to validate password strength
func isStrongPassword(password string) error {
	var hasUpper, hasLower, hasDigit, hasSpecial bool
	specialCharRegex := regexp.MustCompile(`[!@#~$%^&*()_+\-=\[\]{};':"\\|,.<>/?]+`)

	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case specialCharRegex.MatchString(string(char)):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}
	if !hasLower {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}
	if !hasDigit {
		return fmt.Errorf("password must contain at least one digit")
	}
	if !hasSpecial {
		return fmt.Errorf("password must contain at least one special character")
	}

	return nil
}

// ErrorResponse represents the structure of error messages
type ErrorResponse struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// SuccessResponse represents the structure of success messages
type SuccessResponse struct {
	Message string `json:"message"`
}

// Function to send a JSON response
func sendJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// Set a session cookie
func setSessionCookie(w http.ResponseWriter, userID int) {
	cookie := &http.Cookie{
		Name:     "session",
		Value:    fmt.Sprintf("%d", userID),
		Path:     "/",
		HttpOnly: true,
	}
	http.SetCookie(w, cookie)
	log.Printf("setSessionCookie: Cookie set for user ID %d", userID)
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	// Check if the user is already logged in
	_, err := getLoggedInUsername(r)
	if err == nil {
		// If logged in, redirect to the tasks page
		http.Redirect(w, r, "/tasks", http.StatusFound)
		return
	}

	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		password := r.FormValue("password")
		confirmPassword := r.FormValue("confirmPassword")

		// Check if username already exists
		var existingUser string
		err := db.DB.QueryRow("SELECT username FROM users WHERE username = ?", username).Scan(&existingUser)
		if err != nil && err != sql.ErrNoRows {
			sendJSONResponse(w, http.StatusInternalServerError, ErrorResponse{"", "Could not check username availability"})
			return
		}
		if existingUser != "" {
			sendJSONResponse(w, http.StatusConflict, ErrorResponse{"username", "Username already taken"})
			return
		}

		// Validate password strength
		if err := isStrongPassword(password); err != nil {
			sendJSONResponse(w, http.StatusBadRequest, ErrorResponse{"password", err.Error()})
			return
		}

		// Check if passwords match
		if password != confirmPassword {
			sendJSONResponse(w, http.StatusBadRequest, ErrorResponse{"confirmPassword", "Passwords do not match"})
			return
		}

		// Hash the password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			sendJSONResponse(w, http.StatusInternalServerError, ErrorResponse{"", "Could not hash password"})
			return
		}

		// Insert the new user into the database
		_, err = db.DB.Exec("INSERT INTO users (username, password) VALUES (?, ?)", username, hashedPassword)
		if err != nil {
			sendJSONResponse(w, http.StatusInternalServerError, ErrorResponse{"", "Could not create user"})
			return
		}

		// Registration successful
		sendJSONResponse(w, http.StatusOK, SuccessResponse{"User registered successfully"})
	} else {
		http.ServeFile(w, r, "templates/register.html")
	}
}

// Handler for user login
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		password := r.FormValue("password")

		var userID int
		var hashedPassword string

		err := db.DB.QueryRow("SELECT id, password FROM users WHERE username = ?", username).Scan(&userID, &hashedPassword)
		if err != nil {
			sendJSONResponse(w, http.StatusInternalServerError, ErrorResponse{"username", "User does not exist"})
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
		if err != nil {
			sendJSONResponse(w, http.StatusInternalServerError, ErrorResponse{"password", "Incorrect password"})
			return
		}

		// Set the session cookie
		setSessionCookie(w, userID)
		log.Printf("HandleLogin: Set session for user %s (ID: %d)", username, userID)

		// Send a success response with the redirect URL
		sendJSONResponse(w, http.StatusOK, map[string]string{
			"success":  "true",
			"redirect": "/tasks",
		})
		return // Make sure to return after a redirect
	}

	// Serve the login page for GET requests
	http.ServeFile(w, r, "templates/login.html")
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie := &http.Cookie{
		Name:   "session",
		Value:  "",
		Path:   "/",
		MaxAge: -1, // delete the cookie
	}
	http.SetCookie(w, cookie)
	http.Redirect(w, r, "/home", http.StatusFound)
}

// Home page handler
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	// Check if the user is logged in
	_, err := getLoggedInUsername(r)

	if err == nil {
		// Redirect logged-in users to the tasks page
		http.Redirect(w, r, "/tasks", http.StatusFound)
		return
	}

	// Serve the generic home page for non-logged-in users
	http.ServeFile(w, r, "templates/home.html")
}

// TasksHandler renders the list of todos
func TasksHandler(w http.ResponseWriter, r *http.Request) {
	username, err := getLoggedInUsername(r)
	if err != nil {
		log.Println("IndexHandler: User not logged in")
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	log.Printf("IndexHandler: Fetching todos for user %s", username)

	// Fetch the todos from the database
	rows, err := db.DB.Query(`
		SELECT todos.id, todos.title, todos.status
		FROM todos
		INNER JOIN users ON todos.user_id = users.id
		WHERE users.username = ?`, username)
	if err != nil {
		log.Printf("IndexHandler: Failed to fetch todos: %v", err)
		http.Error(w, "Failed to fetch todos", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Parse the rows into a slice of Todo structs
	todos := []models.Todo{}
	for rows.Next() {
		var todo models.Todo
		if err := rows.Scan(&todo.ID, &todo.Title, &todo.Status); err != nil {
			log.Println("Error scanning todo:", err)
			http.Error(w, "Failed to scan todos", http.StatusInternalServerError)
			return
		}
		todos = append(todos, todo)
	}
	log.Printf("IndexHandler: Found %d todos for user %s", len(todos), username)

	// Prepare the data to pass to the template
	data := struct {
		Username string
		Todos    []models.Todo
	}{
		Username: username,
		Todos:    todos,
	}

	// Parse and execute the template with the fetched data
	tmpl, err := template.ParseFiles("templates/tasks.html")
	if err != nil {
		log.Println("Error parsing template:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		log.Println("Error executing template:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// AddHandler handles adding a new todo item
func AddHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// get the user_id from the session cookie
		username, err := getLoggedInUsername(r)
		if err != nil {
			http.Error(w, "Not authenticated", http.StatusUnauthorized)
			return
		}

		// Get the task title from the form input
		title := r.FormValue("title")
		if title == "" {
			http.Error(w, "Task title cannot be empty", http.StatusBadRequest)
			return
		}

		// Fetch the user ID from the database using the username
		var userID string
		err = db.DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
		if err != nil {
			http.Error(w, "Failed to get user ID", http.StatusInternalServerError)
			return
		}

		// insert the new task into the database
		result, err := db.DB.Exec("INSERT INTO todos (user_id, title, status) VALUES (?, ?, ?)", userID, title, false)
		if err != nil {
			http.Error(w, "Failed to create task", http.StatusInternalServerError)
			return
		}

		// check if a row was actually inserted
		rowsAffected, err := result.RowsAffected()
		if err != nil || rowsAffected == 0 {
			http.Error(w, "No task was created", http.StatusInternalServerError)
			return
		}

		// redirect to tasks page after successful creation
		http.Redirect(w, r, "/tasks", http.StatusFound)
	}
}

// DeleteHandler handles deleting a todo item
func DeleteHandler(w http.ResponseWriter, r *http.Request) {
	// Check if the user is authenticated and get the username
	username, err := getLoggedInUsername(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse the task ID from the query string
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	// Fetch the user ID from the database using the username
	var userID int
	err = db.DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
	if err != nil {
		http.Error(w, "Failed to get user ID", http.StatusInternalServerError)
		return
	}

	// Delete the task only if it belongs to the logged-in user
	result, err := db.DB.Exec("DELETE FROM todos WHERE id = ? AND user_id = ?", id, userID)
	if err != nil {
		http.Error(w, "Failed to delete task", http.StatusInternalServerError)
		return
	}

	// Check if any rows were affected
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Task not found or you do not have permission to delete it", http.StatusForbidden)
		return
	}

	// Redirect to the home page after deletion
	http.Redirect(w, r, "/tasks", http.StatusSeeOther)
}

// EditHandler handles editing a todo item
func EditHandler(w http.ResponseWriter, r *http.Request) {
	// Ensure the user is logged in
	username, err := getLoggedInUsername(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse the task ID from the form data
	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	// Get the updated title from the form
	title := r.FormValue("title")
	if title == "" {
		http.Error(w, "Title cannot be empty", http.StatusBadRequest)
		return
	}

	// Determine the status based on the checkbox value
	status := "Pending" // Default to "Pending"
	if r.FormValue("status") == "Completed" {
		status = "Completed"
	}

	// Fetch the user ID using the username
	var userID int
	err = db.DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
	if err != nil {
		http.Error(w, "Failed to get user ID", http.StatusInternalServerError)
		return
	}

	// Check if the task belongs to the logged-in user before updating
	var taskOwnerID int
	err = db.DB.QueryRow("SELECT user_id FROM todos WHERE id = ?", id).Scan(&taskOwnerID)
	if err != nil {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}
	if taskOwnerID != userID {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Update the task in the database
	_, err = db.DB.Exec("UPDATE todos SET title = ?, status = ? WHERE id = ? AND user_id = ?", title, status, id, userID)
	if err != nil {
		http.Error(w, "Failed to update task", http.StatusInternalServerError)
		return
	}

	// Redirect to the home page after updating
	http.Redirect(w, r, "/tasks", http.StatusSeeOther)
}
