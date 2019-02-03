package model

import (
	"time"
)

type User struct {
	Id        int64 `gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index"`

	Username string `gorm:"size:255;not null;unique"`
	Password string `gorm:"size:255"`
	Email    string `gorm:"size:255;not null;unique"`
	Amount   uint64 `gorm:"not null;default:0"`
}
