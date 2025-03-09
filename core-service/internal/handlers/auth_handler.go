package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mrityunjay-vashisth/core-service/internal/services"
	"go.uber.org/zap"
)

type AuthHandler struct {
	Service services.AuthServices
	Logger  *zap.Logger
}

func NewAuthHandler(service services.AuthServices, logger *zap.Logger) *AuthHandler {
	return &AuthHandler{Service: service, Logger: logger}
}

// ServeHTTP routes requests
func (h *AuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	h.Logger.Info("Auth API", zap.String("subpath", vars["subpath"]))
	subPath := vars["subpath"]

	switch subPath {
	case "login":
		h.Login(w, r)
	case "register":
		h.Register(w, r)
	default:
		respondWithError(w, http.StatusNotFound, "Invalid Onboarding API")
	}
}

func (a *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		// http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	authResp, err := a.Service.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	if authResp.Token == "" {
		respondWithError(w, http.StatusUnauthorized, authResp.Message)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{
		"token":   authResp.Token,
		"message": authResp.Message,
		"email":   authResp.Email,
	})
}

func (a *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	// handle reg request
}
