package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"
)

// --- Domain Types ---

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type ErrorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// --- HTTP Server with proper configuration ---

func main() {
	mux := http.NewServeMux()

	// Register routes
	mux.HandleFunc("GET /health", healthHandler)
	mux.HandleFunc("GET /users", listUsersHandler)
	mux.HandleFunc("GET /users/{id}", getUserHandler)
	mux.HandleFunc("POST /users", createUserHandler)

	// Server with proper timeouts
	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,

		// Prevent slowloris attack
		ReadTimeout: 5 * time.Second,

		// Prevent stuck handlers
		WriteTimeout: 10 * time.Second,

		// Close idle connections
		IdleTimeout: 120 * time.Second,

		// Limit header size
		MaxHeaderBytes: 1 << 20, // 1MB

		// Custom error log
		ErrorLog: log.Default(),

		// Connection state tracking (optional)
		ConnState: func(conn net.Conn, state http.ConnState) {
			// Can track active connections here
			_ = conn
			_ = state
		},
	}

	fmt.Println("Server starting on :8080")
	fmt.Println("Endpoints:")
	fmt.Println("  GET  /health")
	fmt.Println("  GET  /users")
	fmt.Println("  GET  /users/{id}")
	fmt.Println("  POST /users")

	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

// --- Handlers ---

func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "healthy",
		"time":   time.Now().Format(time.RFC3339),
	})
}

func listUsersHandler(w http.ResponseWriter, r *http.Request) {
	// Parse query params
	page := r.URL.Query().Get("page")
	if page == "" {
		page = "1"
	}

	users := []User{
		{ID: 1, Name: "Alice", Email: "alice@example.com"},
		{ID: 2, Name: "Bob", Email: "bob@example.com"},
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data": users,
		"pagination": map[string]interface{}{
			"page":      page,
			"page_size": 20,
			"total":     2,
		},
	})
}

func getUserHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Missing user ID")
		return
	}

	// Simulate user lookup
	user := User{ID: 1, Name: "Alice", Email: "alice@example.com"}
	writeJSON(w, http.StatusOK, user)
}

func createUserHandler(w http.ResponseWriter, r *http.Request) {
	// Limit body size
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1MB max

	var input struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	// Validate
	if input.Name == "" {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "name is required")
		return
	}
	if input.Email == "" {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "email is required")
		return
	}

	// Create user (simulated)
	user := User{ID: 3, Name: input.Name, Email: input.Email}

	writeJSON(w, http.StatusCreated, user)
}

// --- Helpers ---

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	resp := ErrorResponse{}
	resp.Error.Code = code
	resp.Error.Message = message
	writeJSON(w, status, resp)
}
