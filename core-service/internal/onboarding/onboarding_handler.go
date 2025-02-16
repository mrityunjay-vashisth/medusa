package onboarding

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var jwtKey = []byte("my-secret-key")

type onboardingHandler struct {
	client *mongo.Client
}

type onboardingRequest struct {
	OrganizationName string `json:"organization_name" bson:"organization_name"`
	Email            string `json:"email" bson:"email"`
}

type onboardingPending struct {
	RequestID        string `bson:"request_id"`
	TenantID         string `bson:"tenant_id"`
	OrganizationName string `bson:"organization_name"`
	Email            string `bson:"email"`
	Username         string `bson:"username"`
	Status           string `bson:"status"`
	CreatedAt        string `bson:"created_at"`
}

type claims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	TenantID string `json:"tenant_id"`
	jwt.RegisteredClaims
}

func NewOnboardingHandler(client *mongo.Client) *onboardingHandler {
	return &onboardingHandler{client: client}
}

func (o *onboardingHandler) OnboardTenant(c *gin.Context) {
	log.Println("Received request")
	var req onboardingRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	if req.OrganizationName == "" || req.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing fields"})
		return
	}
	log.Printf("%v", req)

	requestId := fmt.Sprintf("req-%d", time.Now().UnixNano())
	tenantId := fmt.Sprintf("tenant-%d", time.Now().UnixNano())
	userId := fmt.Sprintf("user-%d", time.Now().UnixNano())

	collection := o.client.Database("coredb").Collection("onboarding_requests")
	pendingReq := &onboardingPending{
		RequestID:        requestId,
		TenantID:         tenantId,
		OrganizationName: req.OrganizationName,
		Email:            req.Email,
		Username:         userId,
		Status:           "pending",
		CreatedAt:        time.Now().String(),
	}
	_, err := collection.InsertOne(context.Background(), pendingReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	log.Printf("New Onboarding Request: %s (Request ID: %s)", req.OrganizationName, requestId)
	c.JSON(http.StatusOK, gin.H{"message": "Onboarding request submitted, pending approval", "request_id": requestId})
}

func (o *onboardingHandler) GetPendingRequests(c *gin.Context) {
	var req struct {
		OwnerToken string `json:"owner_token"`
		Username   string `json:"username"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if !validateServiceOwner(c, req.OwnerToken) {
		return
	}
	collection := o.client.Database("coredb").Collection("onboarding_requests")
	cursor, err := collection.Find(context.Background(), bson.M{"status": "pending"})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer cursor.Close(context.Background())
	var requests []bson.M
	if err = cursor.All(context.Background(), &requests); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error parsing data"})
		return
	}
	c.JSON(http.StatusOK, requests)
}

func (o *onboardingHandler) ApproveOnboarding(c *gin.Context) {
	var req struct {
		RequestID  string `json:"request_id"`
		OwnerToken string `json:"owner_token"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if !validateServiceOwner(c, req.OwnerToken) {
		return
	}
	collection := o.client.Database("coredb").Collection("onboarding_requests")
	entitlementsCollection := o.client.Database("coredb").Collection("onboarded_tenants")
	var request bson.M
	err := collection.FindOne(context.Background(), bson.M{"request_id": req.RequestID, "status": "pending"}).Decode(&request)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No pending request found"})
		return
	}
	userID := fmt.Sprintf("user-%d", time.Now().UnixNano())
	password := generateRandomPassword()
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	_, err = entitlementsCollection.InsertOne(context.Background(), bson.M{
		"tenant_id":      request["tenant_id"],
		"user_id":        userID,
		"password":       string(hashedPassword),
		"admin_email":    request["admin_email"],
		"admin_username": request["admin_username"],
		"status":         "active",
		"created_at":     time.Now(),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to approve request"})
		return
	}
	_, _ = collection.DeleteOne(context.Background(), bson.M{"request_id": req.RequestID})
	log.Printf("Sending admin credentials: Email: %s, User ID: %s, Password: %s", request["admin_email"], userID, password)
	c.JSON(http.StatusOK, gin.H{"message": "Onboarding request approved, credentials sent"})
}

func validateServiceOwner(c *gin.Context, tokenString string) bool {
	ownerClaims, err := validateToken(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid owner token"})
		return false
	}
	if ownerClaims.Role != "service_owner" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized: Only service owners can view/approve"})
		return false
	}
	return true
}

func validateToken(tokenString string) (*claims, error) {
	claims := &claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}
	if time.Now().After(claims.ExpiresAt.Time) {
		return nil, errors.New("token expired")
	}

	return claims, nil
}

func generateRandomPassword() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 12)
	_, _ = rand.Read(b)
	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}
	return string(b)
}
