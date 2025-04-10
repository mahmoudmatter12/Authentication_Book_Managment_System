package models

import (
	"gorm.io/gorm"
)

type Book struct {
	gorm.Model
	Title    string `json:"title" gorm:"not null;unique"`
	Author   string `json:"author"`
	Category string `json:"category"`
}
