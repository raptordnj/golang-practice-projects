package dto

import "time"

// RegisterRequest represents the user registration payload
type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginRequest represents the user login payload
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// UserResponse represents safe user data returned in responses
type UserResponse struct {
	Id        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// AuthResponse represents the response containing token and user profile
type AuthResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}
