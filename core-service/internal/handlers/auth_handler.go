package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/mrityunjay-vashisth/core-service/internal/db"
	"github.com/mrityunjay-vashisth/medusa-proto/authpb"
)

type AuthHandler struct {
	client authpb.AuthServiceClient
	db     db.DBClientInterface
}

func NewAuthHandler(db db.DBClientInterface, client authpb.AuthServiceClient) *AuthHandler {
	return &AuthHandler{
		db:     db,
		client: client,
	}
}

// ServeHTTP routes requests
func (h *AuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Login(w, r)
}

func (a *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	authReq := &authpb.LoginRequest{
		Username: req.Username,
		Password: req.Password,
	}
	authResp, err := a.client.Login(context.Background(), authReq)
	if err != nil {
		log.Printf("Login failed: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if authResp.Token == "" {
		http.Error(w, authResp.Message, http.StatusUnauthorized)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{
		"token":   authResp.Token,
		"message": authResp.Message,
		"email":   authResp.Email,
	})
}
