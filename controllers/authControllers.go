package controllers

import (
	"authSystem/initializers"
	"authSystem/models"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"time"

	"github.com/golang-jwt/jwt/v4"
)

func SignUp(c *gin.Context) {
	// Get the email / password from the request
	var body struct {
		Email    string
		Password string
	}

	// read the body from the request -> if it fails then return a 400 error
	if c.Bind(&body) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Message": "Failed to read the request",
		})

		return
	}

	// Check if the user already exists
	var existingUser models.User
	if err := initializers.DB.Where("email = ?", body.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Message": "User already exists",
		})
		return
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(body.Password), 10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to hash the password",
		})
		return
	}

	// creaate the user
	user := models.User{
		Email:    body.Email,
		Password: string(hashedPassword),
		Role:     "user",
	}

	// Save the user to the database
	if err := initializers.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create the user",
		})
		return
	}

	// Return the user
	c.JSON(http.StatusOK, gin.H{
		"message": "User created successfully",
		"user":    user,
	})
}

func Login(c *gin.Context) {
	// get the email and password from the request
	var body struct {
		Email    string
		Password string
		Role     string
	}

	// read the body from the request -> if it fails then return a 400 error
	if c.Bind(&body) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Message": "Failed to read the request",
		})

		return
	}

	// look up the user based on the email
	// Check if the user already exists
	var existingUser models.User
	if err := initializers.DB.Where("email = ?", body.Email).First(&existingUser).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Message": "User does not exist",
		})
		return
	}

	// compare the password with the hashed password
	if err := bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(body.Password)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Message": "Invalid password",
		})
		return
	}
	fmt.Println(existingUser.Role)
	// Generate a JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": existingUser.ID,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate the token",
		})
		return
	}

	// set cookies
	c.SetCookie("Authorization", tokenString, 3600*24, "", "", false, true)
	// Return the user and the token
	c.JSON(http.StatusOK, gin.H{
		"message": "User logged in successfully",

		"user":  existingUser,
		"token": tokenString,
	})

}

func Validate(c *gin.Context) {
	user, _ := c.Get("user")

	c.JSON(http.StatusOK, gin.H{
		"message": "User is authenticated",
		"user":    user,
	})
}

func Logout(c *gin.Context) {
	// remove the cookie
	c.SetCookie("Authorization", "", -1, "", "", false, true) // Set the cookie to expire immediately
	// Optionally, you can also clear the session or perform any other logout-related actions here
	// For example, if you're using sessions, you might want to clear the session data
	// session := sessions.Default(c)
	// session.Clear()
	// session.Save()

	// Return a success message
	c.JSON(http.StatusOK, gin.H{
		"message": "User logged out successfully",
	})
}
