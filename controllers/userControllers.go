package controllers

import (
	"authSystem/initializers"
	"authSystem/types"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	// "gorm.io/gorm"
)

type UserController struct{}

func NewUserController() *UserController {
	return &UserController{}
}

// GetAllUsers returns a paginated list of users with filtering options
func (uc *UserController) GetAllUsers(c *gin.Context) {
	// Parse pagination parameters with defaults and validation
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 || limit > 50 { // Enforce reasonable limit
		limit = 10
	}
	offset := (page - 1) * limit

	// Parse filter parameters
	email := strings.TrimSpace(c.Query("email"))
	role := strings.TrimSpace(c.Query("role"))

	// Build query
	query := initializers.DB.Model(&types.User{})

	if email != "" {
		query = query.Where("LOWER(email) LIKE ?", "%"+strings.ToLower(email)+"%")
	}
	if role != "" {
		query = query.Where("LOWER(role) LIKE ?", "%"+strings.ToLower(role)+"%")
	}

	// Get total count for pagination
	var total int64
	if err := query.Count(&total).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to count users",
			"details": err.Error(),
		})
		return
	}
	// Fetch paginated results
	var users []types.User
	if err := query.Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to fetch users",
			"details": err.Error(),
		})
		return
	}
	// Return paginated results
	c.JSON(http.StatusOK, gin.H{
		"users": users,
		"Meta": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}
