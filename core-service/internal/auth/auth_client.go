package auth

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mrityunjay-vashisth/medusa-proto/authpb"
)

type AuthClient struct {
	client authpb.AuthServiceClient
}

func NewAuthClient(client authpb.AuthServiceClient) *AuthClient {
	return &AuthClient{client: client}
}

func (a *AuthClient) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	authReq := &authpb.LoginRequest{
		Username: req.Username,
		Password: req.Password,
	}
	authResp, err := a.client.Login(context.Background(), authReq)
	if err != nil {
		log.Printf("Login failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to authenticate"})
		return
	}

	if authResp.Token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":   authResp.Token,
		"message": authResp.Message,
		"email":   authResp.Email,
	})
}
