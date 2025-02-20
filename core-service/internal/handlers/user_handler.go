package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
)

// UserHandler handles user APIs
type UserHandler struct {
	db *mongo.Client
}

// NewUserHandler initializes the handler
func NewUserHandler(db *mongo.Client) *UserHandler {
	return &UserHandler{db: db}
}

// ServeHTTP routes requests
func (h *UserHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	subPath := vars["subpath"]

	switch subPath {
	case "":
		h.CreateUser(w, r)
	case "{id}":
		h.GetUser(w, r)
	default:
		http.Error(w, "Invalid User API", http.StatusNotFound)
	}
}

// CreateUser handles user creation
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{"message": "User Created"})
}

// GetUser handles fetching a user
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{"message": "User Retrieved"})
}
