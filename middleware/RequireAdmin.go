package middleware

import (
	"authSystem/initializers"
	"authSystem/types"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"net/http"
	"os"
	"time"

)

func RequireAdmin(c *gin.Context) {
	fmt.Print("RequireAdmin middleware called\n")
	
	// Get the cookie from the request
	tokenString, err := c.Cookie("Authorization")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized - no token provided"})
		return
	}

	// Parse and validate the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized - invalid token"})
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized - invalid claims"})
		return
	}

	// Check token expiration
	exp, ok := claims["exp"].(float64)
	if !ok || time.Unix(int64(exp), 0).Before(time.Now()) {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized - token expired"})
		return
	}

	// Get user ID from claims
	userID, ok := claims["sub"]
	fmt.Println("User ID from claims:", userID)
	if !ok || userID == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized - invalid user ID"})
		return
	}

	// Find the user in database
	var user types.User
	if err := initializers.DB.First(&user, "id = ?", userID).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized - user not found"})
		return
	}

	// Check if the user is an admin
	if user.Role != "admin" {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden - admin access required"})
		return
	}
	// Set the user in the context
	c.Set("user", user)
	// Continue to the next handler
	c.Next()
}
