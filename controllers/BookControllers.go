package controllers

import (
	"authSystem/initializers"
	"authSystem/types"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type BookController struct{}

func NewBookController() *BookController {
	return &BookController{}
}

// GetAllBooks returns a paginated list of books with filtering options
func (bc *BookController) GetAllBooks(c *gin.Context) {
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
	title := strings.TrimSpace(c.Query("title"))
	author := strings.TrimSpace(c.Query("author"))
	category := strings.TrimSpace(c.Query("category"))

	// Build query
	query := initializers.DB.Model(&types.Book{})

	if title != "" {
		query = query.Where("LOWER(title) LIKE ?", "%"+strings.ToLower(title)+"%")
	}
	if author != "" {
		query = query.Where("LOWER(author) LIKE ?", "%"+strings.ToLower(author)+"%")
	}
	if category != "" {
		query = query.Where("LOWER(category) LIKE ?", "%"+strings.ToLower(category)+"%")
	}

	// Get total count for pagination
	var total int64
	if err := query.Count(&total).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to count books",
			"details": err.Error(),
		})
		return
	}

	// Execute query with pagination
	var books []types.Book
	if err := query.Offset(offset).Limit(limit).Find(&books).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to fetch books",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": books,
		"meta": gin.H{
			"page":       page,
			"limit":      limit,
			"total":      total,
			"totalPages": (int(total) + limit - 1) / limit,
		},
	})
}

// GetBookByID returns a single book by ID
func (bc *BookController) GetBookByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "Book ID is required",
		})
		return
	}

	bookID, err := strconv.Atoi(id)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid book ID format",
			"details": "ID must be a numeric value",
		})
		return
	}

	var book types.Book
	if err := initializers.DB.First(&book, bookID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"error": "Book not found",
			})
		} else {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to retrieve book",
				"details": err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": book,
	})
}

// CreateBook creates a new book record
func (bc *BookController) CreateBook(c *gin.Context) {
	var req types.AddBookRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Validate required fields
	if strings.TrimSpace(req.Title) == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "Title is required",
		})
		return
	}

	if strings.TrimSpace(req.Author) == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "Author is required",
		})
		return
	}

	// Normalize title for case-insensitive duplicate check
	normalizedTitle := strings.ToLower(strings.TrimSpace(req.Title))

	// Check for existing book with same title
	var existingBook types.Book
	if err := initializers.DB.Where("LOWER(TRIM(title)) = ?", normalizedTitle).First(&existingBook).Error; err == nil {
		c.AbortWithStatusJSON(http.StatusConflict, gin.H{
			"error": "A book with this title already exists",
		})
		return
	} else if err != gorm.ErrRecordNotFound {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to check for existing books",
			"details": err.Error(),
		})
		return
	}

	// Create the book
	book := types.Book{
		Title:    req.Title,
		Author:   req.Author,
		Category: req.Category,
	}

	if err := initializers.DB.Create(&book).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create book",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data": book,
	})
}

// UpdateBook updates an existing book record
func (bc *BookController) UpdateBook(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "Book ID is required",
		})
		return
	}

	bookID, err := strconv.Atoi(id)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid book ID format",
			"details": "ID must be a numeric value",
		})
		return
	}

	var req types.AddBookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Check if book exists
	var book types.Book
	if err := initializers.DB.First(&book, bookID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"error": "Book not found",
			})
		} else {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to retrieve book",
				"details": err.Error(),
			})
		}
		return
	}

	// Normalize title for duplicate check (excluding current book)
	normalizedTitle := strings.ToLower(strings.TrimSpace(req.Title))
	var existingBook types.Book
	if err := initializers.DB.Where("LOWER(TRIM(title)) = ? AND id != ?", normalizedTitle, bookID).First(&existingBook).Error; err == nil {
		c.AbortWithStatusJSON(http.StatusConflict, gin.H{
			"error": "A book with this title already exists",
		})
		return
	} else if err != gorm.ErrRecordNotFound {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to check for existing books",
			"details": err.Error(),
		})
		return
	}

	// Update book fields
	book.Title = req.Title
	book.Author = req.Author
	book.Category = req.Category

	if err := initializers.DB.Save(&book).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update book",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": book,
	})
}

// DeleteBook deletes a book record
func (bc *BookController) DeleteBook(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "Book ID is required",
		})
		return
	}

	bookID, err := strconv.Atoi(id)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid book ID format",
			"details": "ID must be a numeric value",
		})
		return
	}

	// Use transaction for data consistency
	err = initializers.DB.Transaction(func(tx *gorm.DB) error {
		var book types.Book
		if err := tx.First(&book, bookID).Error; err != nil {
			return err
		}

		if err := tx.Delete(&book).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"error": "Book not found",
			})
		} else {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to delete book",
				"details": err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Book deleted successfully",
	})
}