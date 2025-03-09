package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mrityunjay-vashisth/core-service/internal/db"
)

type UserHandlerInterface interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
	CreateUser(w http.ResponseWriter, r *http.Request)
	GetUser(w http.ResponseWriter, r *http.Request)
}

// UserHandler handles user APIs
type UserHandler struct {
	db db.DBClientInterface
}

// NewUserHandler initializes the handler
func NewUserHandler(db db.DBClientInterface) *UserHandler {
	return &UserHandler{db: db}
}

// ServeHTTP routes requests
func (h *UserHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	subPath := vars["subpath"]

	switch subPath {
	case "register":
		h.CreateUser(w, r)
	case "getuser":
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
