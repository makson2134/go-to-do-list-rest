package models

import (
	"time"
	"fmt"
	"golang.org/x/crypto/bcrypt"
)
type User struct {
	ID        uint
	Username  string
	Email     string
	Password  []byte // хэш от bcrypt
	CreatedAt time.Time
	UpdatedAt time.Time

	//Один ко многим
	Tasks []Task
}


func NewUser(username, email, password string) (*User, error) {
	if len(username) < 3 || len(username) > 20 {
		return nil, fmt.Errorf("username should be between 3 and 20 characters")
	}
	
	if len(email) > 100 {
		return nil, fmt.Errorf("email should be shorter than 100 characters")
	}
	
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}
	
	return &User{
		Username: username,
		Email:    email,
		Password: hashedPassword,
	}, nil
}