package types

import (
	"time"
)

type User struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	Email     string `json:"email" gorm:"unique;not null"`
	Password  string `json:"password" gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Role      string `json:"role" gorm:"default:'user'"`
}

type Book struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	Author   string `json:"author"`
	Category string `json:"category"`
}

type AddBookRequest struct {
	Title    string `json:"title"`
	Author   string `json:"author"`
	Category string `json:"category"`
}
